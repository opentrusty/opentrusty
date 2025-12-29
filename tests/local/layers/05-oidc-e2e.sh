#!/bin/bash
# Layer 5: End-to-End OIDC Flow
# Validates the complete OAuth2 Authorization Code + PKCE flow

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
TENANT_ID=$(cat "$STATE_DIR/tenant_id.txt" 2>/dev/null || echo "")
CLIENT_ID=$(cat "$STATE_DIR/client_id.txt" 2>/dev/null || echo "")
CLIENT_SECRET=$(cat "$STATE_DIR/client_secret.txt" 2>/dev/null || echo "")

# For OIDC, use the platform admin as the end user (has valid session)
REDIRECT_URI="http://localhost:8081/callback"

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

echo "Layer 5: End-to-End OIDC Flow Tests"
echo "===================================="
echo ""

# Validate prerequisites
if [ -z "$TENANT_ID" ] || [ -z "$CLIENT_ID" ] || [ -z "$CLIENT_SECRET" ]; then
    echo -e "${RED}Missing state from previous layers. Run layers 1-4 first.${NC}"
    exit 1
fi

echo "  Using Client ID: $CLIENT_ID"
echo "  Using Tenant: $TENANT_ID"
echo ""

# Generate PKCE values
VERIFIER=$(openssl rand -base64 32 | tr -d '/+=' | head -c 43)
CHALLENGE=$(echo -n "$VERIFIER" | openssl dgst -sha256 -binary | base64 | tr '/+' '_-' | tr -d '=')
STATE="state_$(date +%s)"
NONCE="nonce_$(date +%s)"

# Test 1: User has valid session (from Layer 2)
echo -n "  [1] User session valid... "

ME_RESPONSE=$(curl -s "$API_BASE/api/v1/auth/me" \
    -b "$STATE_DIR/admin_cookies.txt")

if echo "$ME_RESPONSE" | jq -e '.user.user_id' >/dev/null 2>&1; then
    USER_ID=$(echo "$ME_RESPONSE" | jq -r '.user.user_id')
    echo -e "${GREEN}PASS${NC}"
    add_result "User Session" "PASS" "user_id: $USER_ID"
else
    echo -e "${RED}FAIL${NC}"
    add_result "User Session" "FAIL" "No valid session"
    exit 1
fi

# Test 2: Authorization request (with tenant_id as query param for OAuth2)
echo -n "  [2] Authorization request... "

AUTH_URL="$API_BASE/oauth2/authorize?client_id=$CLIENT_ID&redirect_uri=$(printf '%s' "$REDIRECT_URI" | jq -sRr @uri)&response_type=code&scope=openid%20profile%20email&state=$STATE&nonce=$NONCE&code_challenge=$CHALLENGE&code_challenge_method=S256&tenant_id=$TENANT_ID"

# Don't follow redirects - expect 302 to callback URL
AUTH_RESPONSE=$(curl -s -w "\n%{http_code}\n%{redirect_url}" \
    -b "$STATE_DIR/admin_cookies.txt" \
    "$AUTH_URL")

# Parse response - use sed for macOS compatibility (head -n -1 doesn't work on BSD)
BODY=$(echo "$AUTH_RESPONSE" | sed '$d' | sed '$d')
HTTP_CODE=$(echo "$AUTH_RESPONSE" | tail -2 | head -1)
REDIR_URL=$(echo "$AUTH_RESPONSE" | tail -1)

CODE=""
if [ "$HTTP_CODE" == "302" ] && echo "$REDIR_URL" | grep -q "code="; then
    CODE=$(echo "$REDIR_URL" | grep -oE 'code=[^&]+' | cut -d= -f2)
    echo -e "${GREEN}PASS${NC}"
    add_result "Authorization Request" "PASS" "Got authorization code"
    echo "$CODE" > "$STATE_DIR/auth_code.txt"
else
    echo -e "${RED}FAIL (HTTP $HTTP_CODE)${NC}"
    add_result "Authorization Request" "FAIL" "HTTP $HTTP_CODE, no code in redirect"
fi

# Test 3: Verify state parameter in redirect
echo -n "  [3] State parameter in redirect... "

if echo "$REDIR_URL" | grep -q "state=$STATE"; then
    echo -e "${GREEN}PASS${NC}"
    add_result "State Validation" "PASS" "State matches"
else
    echo -e "${RED}FAIL${NC}"
    add_result "State Validation" "FAIL" "State not found in redirect"
fi

# Test 4: Token exchange
echo -n "  [4] Token exchange... "

if [ -z "$CODE" ]; then
    echo -e "${RED}SKIP (no code)${NC}"
    add_result "Token Exchange" "FAIL" "No authorization code available"
else
    TOKEN_RESPONSE=$(curl -s -X POST "$API_BASE/oauth2/token?tenant_id=$TENANT_ID" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "grant_type=authorization_code" \
        -d "code=$CODE" \
        -d "redirect_uri=$REDIRECT_URI" \
        -d "client_id=$CLIENT_ID" \
        -d "client_secret=$CLIENT_SECRET" \
        -d "code_verifier=$VERIFIER")
    
    if echo "$TOKEN_RESPONSE" | jq -e '.access_token' >/dev/null 2>&1; then
        ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')
        ID_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.id_token')
        echo -e "${GREEN}PASS${NC}"
        add_result "Token Exchange" "PASS" "Got access_token and id_token"
        echo "$ACCESS_TOKEN" > "$STATE_DIR/access_token.txt"
        echo "$ID_TOKEN" > "$STATE_DIR/id_token.txt"
    else
        ERROR=$(echo "$TOKEN_RESPONSE" | jq -r '.error // .error_description // "unknown"')
        echo -e "${RED}FAIL${NC}"
        add_result "Token Exchange" "FAIL" "Error: $ERROR"
    fi
fi

# Test 5: Validate ID Token claims
echo -n "  [5] ID Token claims validation... "

ID_TOKEN=$(cat "$STATE_DIR/id_token.txt" 2>/dev/null || echo "")

if [ -z "$ID_TOKEN" ]; then
    echo -e "${RED}SKIP (no token)${NC}"
    add_result "ID Token Validation" "FAIL" "No ID token available"
else
    # Decode JWT payload (base64url -> base64 -> decode)
    PAYLOAD_B64=$(echo "$ID_TOKEN" | cut -d'.' -f2)
    # Add padding if needed
    PADDING=$((4 - ${#PAYLOAD_B64} % 4))
    if [ $PADDING -ne 4 ]; then
        PAYLOAD_B64="${PAYLOAD_B64}$(printf '%*s' $PADDING '' | tr ' ' '=')"
    fi
    PAYLOAD=$(echo "$PAYLOAD_B64" | tr '_-' '/+' | base64 -d 2>/dev/null || echo "{}")
    
    ISS=$(echo "$PAYLOAD" | jq -r '.iss // "missing"')
    AUD=$(echo "$PAYLOAD" | jq -r '.aud // "missing"')
    SUB=$(echo "$PAYLOAD" | jq -r '.sub // "missing"')
    
    CLAIMS_OK=true
    CLAIM_DETAILS=""
    
    if [ "$ISS" == "missing" ]; then
        CLAIMS_OK=false
        CLAIM_DETAILS+="iss:missing "
    else
        CLAIM_DETAILS+="iss:ok "
    fi
    
    if [ "$AUD" == "missing" ]; then
        CLAIMS_OK=false
        CLAIM_DETAILS+="aud:missing "
    elif [ "$AUD" != "$CLIENT_ID" ]; then
        CLAIMS_OK=false
        CLAIM_DETAILS+="aud:mismatch "
    else
        CLAIM_DETAILS+="aud:ok "
    fi
    
    if [ "$SUB" == "missing" ]; then
        CLAIMS_OK=false
        CLAIM_DETAILS+="sub:missing "
    else
        CLAIM_DETAILS+="sub:ok "
    fi
    
    if [ "$CLAIMS_OK" == "true" ]; then
        echo -e "${GREEN}PASS${NC}"
        add_result "ID Token Validation" "PASS" "$CLAIM_DETAILS"
    else
        echo -e "${RED}FAIL${NC}"
        add_result "ID Token Validation" "FAIL" "$CLAIM_DETAILS"
    fi
    
    # Save decoded payload
    echo "$PAYLOAD" | jq . > "$STATE_DIR/id_token_claims.json" 2>/dev/null || echo "$PAYLOAD" > "$STATE_DIR/id_token_claims.json"
fi

# Generate report
cat > "$REPORTS_DIR/oidc-e2e.md" << EOF
# Layer 5: End-to-End OIDC Flow Report

**Generated:** $(date -Iseconds)

## Test Configuration

- **Tenant ID:** $TENANT_ID
- **Client ID:** $CLIENT_ID
- **Redirect URI:** $REDIRECT_URI
- **PKCE Challenge Method:** S256

## Results

| Test | Status | Details |
|------|--------|---------|
$(echo -e "$REPORT")

## Summary

- **Passed:** $PASSED
- **Failed:** $FAILED

## ID Token Claims

\`\`\`json
$(cat "$STATE_DIR/id_token_claims.json" 2>/dev/null || echo "Not available")
\`\`\`

EOF

echo ""
echo "Report: $REPORTS_DIR/oidc-e2e.md"

if [ $FAILED -eq 0 ]; then
    exit 0
else
    exit 1
fi
