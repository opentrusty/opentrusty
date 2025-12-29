#!/bin/bash
# Layer 1: Infrastructure Sanity
# Validates all services are reachable and healthy

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
REPORTS_DIR="$SCRIPT_DIR/../reports"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

PASSED=0
FAILED=0
REPORT=""

check_service() {
    local name=$1
    local url=$2
    local expected=$3
    
    echo -n "  Checking $name... "
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")
    
    if [ "$response" == "$expected" ]; then
        echo -e "${GREEN}PASS ($response)${NC}"
        REPORT+="| $name | $url | $expected | $response | ✅ PASS |\n"
        PASSED=$((PASSED + 1))
        return 0
    else
        echo -e "${RED}FAIL (got $response, expected $expected)${NC}"
        REPORT+="| $name | $url | $expected | $response | ❌ FAIL |\n"
        FAILED=$((FAILED + 1))
        return 1
    fi
}

check_port() {
    local name=$1
    local host=$2
    local port=$3
    
    echo -n "  Checking $name ($host:$port)... "
    
    if nc -z "$host" "$port" 2>/dev/null; then
        echo -e "${GREEN}OPEN${NC}"
        REPORT+="| $name | $host:$port | open | open | ✅ PASS |\n"
        PASSED=$((PASSED + 1))
        return 0
    else
        echo -e "${RED}CLOSED${NC}"
        REPORT+="| $name | $host:$port | open | closed | ❌ FAIL |\n"
        FAILED=$((FAILED + 1))
        return 1
    fi
}

echo "Layer 1: Infrastructure Sanity Tests"
echo "====================================="
echo ""

# Check PostgreSQL
check_port "PostgreSQL" "localhost" "5434" || true

# Check OpenTrusty health
check_service "OpenTrusty Health" "http://localhost:8090/health" "200" || true

# Check OIDC Discovery
check_service "OIDC Discovery" "http://localhost:8090/.well-known/openid-configuration" "200" || true

# Check JWKS (correct path is /jwks.json, not /.well-known/jwks.json)
check_service "JWKS Endpoint" "http://localhost:8090/jwks.json" "200" || true

# Check Control Panel (optional - may not be running)
if curl -s http://localhost:5173/admin/ >/dev/null 2>&1; then
    check_service "Control Panel" "http://localhost:5173/admin/" "200" || true
else
    echo "  Control Panel not running (optional)"
fi

# Generate report
cat > "$REPORTS_DIR/infra-smoke.md" << EOF
# Layer 1: Infrastructure Sanity Report

**Generated:** $(date -Iseconds)

## Results

| Service | URL/Port | Expected | Actual | Status |
|---------|----------|----------|--------|--------|
$(echo -e "$REPORT")

## Summary

- **Passed:** $PASSED
- **Failed:** $FAILED

EOF

echo ""
echo "Report: $REPORTS_DIR/infra-smoke.md"

# Return success only if all critical checks passed
if [ $FAILED -eq 0 ]; then
    exit 0
else
    exit 1
fi
