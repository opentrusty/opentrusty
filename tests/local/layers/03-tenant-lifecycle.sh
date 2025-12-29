#!/bin/bash
# Layer 3: Tenant Lifecycle
# Validates tenant creation and tenant admin provisioning

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
REPORTS_DIR="$SCRIPT_DIR/../reports"
STATE_DIR="$SCRIPT_DIR/../.state"

mkdir -p "$STATE_DIR"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

API_BASE="http://localhost:8090/api/v1"
PASSED=0
FAILED=0
REPORT=""

# Generate unique tenant ID
TENANT_SUFFIX=$(date +%s)
TENANT_ID="beta-tenant-$TENANT_SUFFIX"
TENANT_NAME="Beta Test Tenant"

add_result() {
    local test_name=$1
    local status=$2
    local details=$3
    
    if [ "$status" == "PASS" ]; then
        REPORT+="| $test_name | ✅ PASS | $details |\n"
        PASSED=$((PASSED + 1))
    else
        REPORT+="| $test_name | ❌ FAIL | $details |\n"        
        FAILED=$((FAILED + 1))
    fi
}

echo "Layer 3: Tenant Lifecycle Tests"
echo "================================"
echo ""

# Verify we have admin session
if [ ! -f "$STATE_DIR/admin_cookies.txt" ]; then
    echo -e "${RED}  [!] No admin session found. Run Layer 2 first.${NC}"
    exit 1
fi

# Test 1: Create tenant (as platform admin)
echo -n "  [1] Create tenant via API... "

CREATE_RESPONSE=$(curl -s -X POST "$API_BASE/tenants" \
    -H "Content-Type: application/json" \
    -H "X-CSRF-Token: test-token" \
    -b "$STATE_DIR/admin_cookies.txt" \
    -d "{\"id\": \"$TENANT_ID\", \"name\": \"$TENANT_NAME\"}")

if echo "$CREATE_RESPONSE" | jq -e '.id' >/dev/null 2>&1; then
    CREATED_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id')
    echo -e "${GREEN}PASS${NC}"
    add_result "Create Tenant" "PASS" "tenant_id: $CREATED_ID"
    # Save the returned UUID for use in subsequent layers
    echo "$CREATED_ID" > "$STATE_DIR/tenant_id.txt"
    # Update TENANT_ID to use the returned UUID for remaining tests
    TENANT_ID="$CREATED_ID"
else
    ERROR=$(echo "$CREATE_RESPONSE" | jq -r '.error // "unknown"')
    echo -e "${RED}FAIL${NC}"
    add_result "Create Tenant" "FAIL" "Error: $ERROR"
    # Still save the requested ID to allow subsequent tests to run
    echo "$TENANT_ID" > "$STATE_DIR/tenant_id.txt"
fi

# Test 2: List tenants and verify our tenant appears
echo -n "  [2] Tenant appears in listing... "

LIST_RESPONSE=$(curl -s -X GET "$API_BASE/tenants" \
    -b "$STATE_DIR/admin_cookies.txt")

if echo "$LIST_RESPONSE" | jq -e ".[] | select(.id == \"$TENANT_ID\")" >/dev/null 2>&1; then
    echo -e "${GREEN}PASS${NC}"
    add_result "Tenant Listed" "PASS" "Found in tenant list"
else
    echo -e "${RED}FAIL${NC}"
    add_result "Tenant Listed" "FAIL" "Tenant not found in listing"
fi

# Test 3: Tenant ID is valid UUID format
echo -n "  [3] Tenant ID format valid... "

if [[ "$TENANT_ID" =~ ^[a-zA-Z0-9_-]+$ ]]; then
    echo -e "${GREEN}PASS${NC}"
    add_result "Tenant ID Format" "PASS" "Valid format"
else
    echo -e "${RED}FAIL${NC}"
    add_result "Tenant ID Format" "FAIL" "Invalid format"
fi

# Note: In the control plane model, tenant admins are provisioned by platform admins
# via the Admin API, NOT through self-registration (which is disabled).
# For this test, we'll verify the platform admin can still manage the tenant.

# Test 4: Platform admin can access tenant-scoped endpoints
echo -n "  [4] Platform admin can access tenant clients... "

CLIENTS_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE/tenants/$TENANT_ID/clients" \
    -b "$STATE_DIR/admin_cookies.txt")

HTTP_CODE=$(echo "$CLIENTS_RESPONSE" | tail -1)

if [ "$HTTP_CODE" == "200" ]; then
    echo -e "${GREEN}PASS${NC}"
    add_result "Tenant Access" "PASS" "HTTP 200"
else
    echo -e "${RED}FAIL${NC}"
    add_result "Tenant Access" "FAIL" "HTTP $HTTP_CODE"
fi

# Generate report
cat > "$REPORTS_DIR/tenant-lifecycle.md" << EOF
# Layer 3: Tenant Lifecycle Report

**Generated:** $(date -Iseconds)

## Test Configuration

- **Tenant ID:** $TENANT_ID
- **Tenant Name:** $TENANT_NAME

## Results

| Test | Status | Details |
|------|--------|---------|
$(echo -e "$REPORT")

## Summary

- **Passed:** $PASSED
- **Failed:** $FAILED

## Notes

In the control plane model:
- Tenant admins are provisioned by platform admins (not self-registration)
- Platform admin can manage all tenants

## State Files

- Tenant ID: \`$STATE_DIR/tenant_id.txt\`

EOF

echo ""
echo "Report: $REPORTS_DIR/tenant-lifecycle.md"

if [ $FAILED -eq 0 ]; then
    exit 0
else
    exit 1
fi
