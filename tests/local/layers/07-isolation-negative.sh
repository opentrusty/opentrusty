#!/bin/bash
# Layer 7: Negative & Isolation Tests
# Validates security boundaries and error handling

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
REPORTS_DIR="$SCRIPT_DIR/../reports"
STATE_DIR="$SCRIPT_DIR/../.state"

mkdir -p "$STATE_DIR"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

API_BASE="http://localhost:8090"
PASSED=0
FAILED=0
REPORT=""

# Load from previous layers
TENANT_ID=$(cat "$STATE_DIR/tenant_id.txt" 2>/dev/null || echo "test-tenant")
CLIENT_ID=$(cat "$STATE_DIR/client_id.txt" 2>/dev/null || echo "test-client")
CLIENT_SECRET=$(cat "$STATE_DIR/client_secret.txt" 2>/dev/null || echo "")

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

echo "Layer 7: Negative & Isolation Tests"
echo "===================================="
echo ""

# Test 1: Invalid redirect_uri should be rejected
echo -n "  [1] Invalid redirect_uri rejected... "

INVALID_REDIRECT="http://evil.com/callback"
AUTH_RESPONSE=$(curl -s -w "\n%{http_code}" \
    "$API_BASE/oauth2/authorize?client_id=$CLIENT_ID&redirect_uri=$INVALID_REDIRECT&response_type=code&scope=openid&tenant_id=$TENANT_ID" \
    -b "$STATE_DIR/admin_cookies.txt")

HTTP_CODE=$(echo "$AUTH_RESPONSE" | tail -1)
BODY=$(echo "$AUTH_RESPONSE" | sed '$d')

# Should get 400 or an error response
if [ "$HTTP_CODE" == "400" ] || echo "$BODY" | grep -qi "redirect\|invalid"; then
    echo -e "${GREEN}PASS${NC}"
    add_result "Invalid redirect_uri" "PASS" "Request rejected (HTTP $HTTP_CODE)"
else
    echo -e "${RED}FAIL${NC}"
    add_result "Invalid redirect_uri" "FAIL" "Expected rejection, got HTTP $HTTP_CODE"
fi

# Test 2: Unauthenticated admin API access denied
echo -n "  [2] Unauthenticated admin access denied... "

UNAUTH_RESPONSE=$(curl -s -w "\n%{http_code}" \
    "$API_BASE/api/v1/tenants")

HTTP_CODE=$(echo "$UNAUTH_RESPONSE" | tail -1)

if [ "$HTTP_CODE" == "401" ] || [ "$HTTP_CODE" == "403" ]; then
    echo -e "${GREEN}PASS${NC}"
    add_result "Unauthenticated Access" "PASS" "Access denied (HTTP $HTTP_CODE)"
else
    echo -e "${RED}FAIL${NC}"
    add_result "Unauthenticated Access" "FAIL" "Expected 401/403, got $HTTP_CODE"
fi

# Test 3: Session required for /auth/me
echo -n "  [3] Session required for /auth/me... "

NOSESSION_RESPONSE=$(curl -s -w "\n%{http_code}" \
    "$API_BASE/api/v1/auth/me")

HTTP_CODE=$(echo "$NOSESSION_RESPONSE" | tail -1)

if [ "$HTTP_CODE" == "401" ]; then
    echo -e "${GREEN}PASS${NC}"
    add_result "Session Required" "PASS" "401 Unauthorized returned"
else
    echo -e "${RED}FAIL${NC}"
    add_result "Session Required" "FAIL" "Expected 401, got $HTTP_CODE"
fi

# Test 4: Invalid client_secret rejected at token endpoint
echo -n "  [4] Invalid client_secret rejected... "

TOKEN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE/oauth2/token?tenant_id=$TENANT_ID" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "grant_type=authorization_code" \
    -d "code=fake_code_12345" \
    -d "redirect_uri=http://localhost:8081/callback" \
    -d "client_id=$CLIENT_ID" \
    -d "client_secret=wrong_secret_xyz")

HTTP_CODE=$(echo "$TOKEN_RESPONSE" | tail -1)

if [ "$HTTP_CODE" == "401" ] || [ "$HTTP_CODE" == "400" ]; then
    echo -e "${GREEN}PASS${NC}"
    add_result "Invalid Client Secret" "PASS" "Request rejected (HTTP $HTTP_CODE)"
else
    echo -e "${RED}FAIL${NC}"
    add_result "Invalid Client Secret" "FAIL" "Expected 401/400, got $HTTP_CODE"
fi

# Test 5: Reused/invalid auth code rejected
echo -n "  [5] Invalid auth code rejected... "

# Use obviously invalid code
TOKEN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE/oauth2/token?tenant_id=$TENANT_ID" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "grant_type=authorization_code" \
    -d "code=invalid_code_xyz123" \
    -d "redirect_uri=http://localhost:8081/callback" \
    -d "client_id=$CLIENT_ID" \
    -d "client_secret=$CLIENT_SECRET")

HTTP_CODE=$(echo "$TOKEN_RESPONSE" | tail -1)

if [ "$HTTP_CODE" == "400" ] || [ "$HTTP_CODE" == "401" ]; then
    echo -e "${GREEN}PASS${NC}"
    add_result "Invalid Auth Code" "PASS" "Code rejected (HTTP $HTTP_CODE)"
else
    echo -e "${RED}FAIL${NC}"
    add_result "Invalid Auth Code" "FAIL" "Expected 400/401, got $HTTP_CODE"
fi

# Test 6: Login rejects X-Tenant-ID header (security hardening)
echo -n "  [6] Login rejects X-Tenant-ID header... "

SPOOF_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -H "X-CSRF-Token: test-token" \
    -H "X-Tenant-ID: spoofed-tenant" \
    -d '{"email":"test@test.com","password":"test123"}')

HTTP_CODE=$(echo "$SPOOF_RESPONSE" | tail -1)
BODY=$(echo "$SPOOF_RESPONSE" | sed '$d')

# Should reject with 400 (tenant context must not be provided)
# or 403 if other validation runs first
if [ "$HTTP_CODE" == "400" ] || [ "$HTTP_CODE" == "403" ]; then
    echo -e "${GREEN}PASS${NC}"
    add_result "Tenant Header Spoofing" "PASS" "Spoofing blocked (HTTP $HTTP_CODE)"
else
    echo -e "${RED}FAIL${NC}"
    add_result "Tenant Header Spoofing" "FAIL" "Expected 400/403, got $HTTP_CODE"
fi

# Test 7: Anonymous registration disabled
echo -n "  [7] Anonymous registration disabled... "

REG_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d '{"email":"anon@test.com","password":"test123"}')

HTTP_CODE=$(echo "$REG_RESPONSE" | tail -1)

if [ "$HTTP_CODE" == "403" ]; then
    echo -e "${GREEN}PASS${NC}"
    add_result "Registration Disabled" "PASS" "Registration blocked (HTTP 403)"
else
    echo -e "${RED}FAIL${NC}"
    add_result "Registration Disabled" "FAIL" "Expected 403, got $HTTP_CODE"
fi

# Generate report
cat > "$REPORTS_DIR/isolation-negative.md" << EOF
# Layer 7: Negative & Isolation Tests Report

**Generated:** $(date -Iseconds)

## Security Boundary Tests

These tests validate that security boundaries are properly enforced
and error conditions are handled securely.

## Results

| Test | Status | Details |
|------|--------|---------|
$(echo -e "$REPORT")

## Summary

- **Passed:** $PASSED
- **Failed:** $FAILED

## Tested Boundaries

### Authentication Security
- Session required for protected endpoints
- Anonymous registration disabled
- Tenant header spoofing blocked on login

### OAuth2 Protocol Security
- Invalid redirect_uri rejected
- Invalid client_secret rejected
- Invalid/reused auth codes rejected

### Access Control
- Unauthenticated API access denied

EOF

echo ""
echo "Report: $REPORTS_DIR/isolation-negative.md"

if [ $FAILED -eq 0 ]; then
    exit 0
else
    exit 1
fi
