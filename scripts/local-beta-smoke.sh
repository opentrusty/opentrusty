#!/bin/bash
# local-beta-smoke.sh
# Automates health checks for the OpenTrusty Beta environment.

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo "🚀 Starting OpenTrusty Local Beta Smoke Test..."

function check_status() {
    local service=$1
    local url=$2
    local expected_code=$3
    
    # Use 127.0.0.1 instead of localhost for better reliability with dev servers
    local check_url=$(echo "$url" | sed 's/localhost/127.0.0.1/')
    
    echo -n "Checking $service ($url)... "
    # Added --connect-timeout and -4 to avoid resolution hangs on macOS
    # Use || true to prevent set -e from exiting the script if curl fails (e.g. Connection Refused)
    response=$(curl -s -4 --connect-timeout 2 -o /dev/null -w "%{http_code}" "$check_url" || true)
    
    if [ "$response" == "$expected_code" ]; then
        echo -e "${GREEN}PASS ($response)${NC}"
    else
        echo -e "${RED}FAIL ($response)${NC}"
        echo "  [!] Hint: Ensure the service is running and reachable at 127.0.0.1"
        return 1
    fi
}

# 1. PostgreSQL (via pg_isready or port check)
# Since we use docker-compose, we can check the port 5433
echo -n "Checking PostgreSQL (127.0.0.1:5433)... "
if nc -4 -z 127.0.0.1 5433; then
    echo -e "${GREEN}UP${NC}"
else
    echo -e "${RED}DOWN${NC}"
    exit 1
fi

# 2. OpenTrusty Core (Auth/Admin Plane)
check_status "OpenTrusty Core" "http://localhost:8080/health" "200"

# 3. Control Panel (Vite Dev Server)
# Note: /admin/ is the base path
check_status "Control Panel" "http://localhost:5173/admin/" "200"

# 4. Demo Client
check_status "Demo Client" "http://localhost:8081/" "200"

echo ""
echo -e "${GREEN}✅ All services are healthy and reachable.${NC}"
echo "Ready for Beta validation."
