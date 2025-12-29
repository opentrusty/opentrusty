#!/bin/bash
# Layer 2: Bootstrap & Platform Admin
# Validates platform admin auto-provisioning and login

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

# Bootstrap credentials from environment (set by run-all.sh)
ADMIN_EMAIL="${OT_BOOTSTRAP_ADMIN_EMAIL:-admin@platform.local}"
ADMIN_PASSWORD="${OT_BOOTSTRAP_ADMIN_PASSWORD:-}"

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

echo "Layer 2: Bootstrap & Platform Admin Tests"
echo "=========================================="
echo ""

# Check for password in environment or from server log
if [ -z "$ADMIN_PASSWORD" ]; then
    # Try to extract from server log
    if [ -f "$REPORTS_DIR/opentrusty.log" ]; then
        # Format: Password: <value>
        ADMIN_PASSWORD=$(grep "^Password:" "$REPORTS_DIR/opentrusty.log" 2>/dev/null | head -1 | sed 's/Password: //' || echo "")
    fi
fi

if [ -z "$ADMIN_PASSWORD" ]; then
    echo -e "${RED}  [!] No bootstrap password found. Set OT_BOOTSTRAP_ADMIN_PASSWORD env var.${NC}"
    add_result "Bootstrap Password Available" "FAIL" "Password not found in env or log"
else
    add_result "Bootstrap Password Available" "PASS" "Password available"
fi

# Test 1: Platform admin login
# NOTE: Login endpoint rejects X-Tenant-ID header (derived from user record)
# NOTE: Must include X-CSRF-Token header for CSRF protection
echo -n "  [1] Platform admin login... "

LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/auth/login" \
    -H "Content-Type: application/json" \
    -H "X-CSRF-Token: test-token" \
    -c "$STATE_DIR/admin_cookies.txt" \
    -d "{\"email\": \"$ADMIN_EMAIL\", \"password\": \"$ADMIN_PASSWORD\"}")

if echo "$LOGIN_RESPONSE" | jq -e '.user_id' >/dev/null 2>&1; then
    USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.user_id')
    echo -e "${GREEN}PASS${NC}"
    add_result "Platform Admin Login" "PASS" "user_id: $USER_ID"
else
    ERROR=$(echo "$LOGIN_RESPONSE" | jq -r '.error // "unknown"')
    echo -e "${RED}FAIL${NC}"
    add_result "Platform Admin Login" "FAIL" "Error: $ERROR"
fi

# Test 2: Verify session exists (check cookie file)
echo -n "  [2] Session cookie created... "
if [ -f "$STATE_DIR/admin_cookies.txt" ] && grep -q "opentrusty_session" "$STATE_DIR/admin_cookies.txt"; then
    echo -e "${GREEN}PASS${NC}"
    add_result "Session Cookie" "PASS" "Cookie file contains opentrusty_session"
else
    echo -e "${RED}FAIL${NC}"
    add_result "Session Cookie" "FAIL" "No session cookie found"
fi

# Test 3: Get current user to verify session
echo -n "  [3] Verify session via /auth/me... "

ME_RESPONSE=$(curl -s -X GET "$API_BASE/auth/me" \
    -b "$STATE_DIR/admin_cookies.txt")

if echo "$ME_RESPONSE" | jq -e '.user' >/dev/null 2>&1; then
    ME_EMAIL=$(echo "$ME_RESPONSE" | jq -r '.user.email')
    echo -e "${GREEN}PASS${NC}"
    add_result "Session Verification" "PASS" "email: $ME_EMAIL"
else
    ERROR=$(echo "$ME_RESPONSE" | jq -r '.error // "unknown"')
    echo -e "${RED}FAIL${NC}"
    add_result "Session Verification" "FAIL" "Error: $ERROR"
fi

# Test 4: Check role assignments (platform admin should have platform scope)
echo -n "  [4] Platform admin role present... "

if echo "$ME_RESPONSE" | jq -e '.role_assignments[] | select(.scope == "platform")' >/dev/null 2>&1; then
    echo -e "${GREEN}PASS${NC}"
    add_result "Platform Role" "PASS" "Has platform-scoped role"
else
    echo -e "${RED}FAIL${NC}"
    add_result "Platform Role" "FAIL" "No platform-scoped role found"
fi

# Generate report
cat > "$REPORTS_DIR/bootstrap-test.md" << EOF
# Layer 2: Bootstrap & Platform Admin Report

**Generated:** $(date -Iseconds)

## Test Configuration

- **Admin Email:** $ADMIN_EMAIL
- **Password Source:** Environment or server log

## Results

| Test | Status | Details |
|------|--------|---------|
$(echo -e "$REPORT")

## Summary

- **Passed:** $PASSED
- **Failed:** $FAILED

## State Files

- Admin cookies: \`$STATE_DIR/admin_cookies.txt\`
- Admin user ID: \`$STATE_DIR/admin_user_id.txt\`

EOF

echo ""
echo "Report: $REPORTS_DIR/bootstrap-test.md"

if [ $FAILED -eq 0 ]; then
    exit 0
else
    exit 1
fi
