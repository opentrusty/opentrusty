#!/bin/bash
# run-all.sh - Master orchestrator for Local System Validation
# Runs all test layers from a clean database state

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../.." &> /dev/null && pwd )"
CONTROL_PANEL_DIR="$PROJECT_ROOT/../opentrusty-control-panel"
DEMO_APP_DIR="$PROJECT_ROOT/../opentrusty-demo-app"
REPORTS_DIR="$SCRIPT_DIR/reports"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# PIDs for cleanup
OPENTRUSTY_PID=""
CONTROL_PANEL_PID=""
DEMO_APP_PID=""

# Test environment
export DB_HOST=localhost
export DB_PORT=5434
export DB_USER=opentrusty
export DB_PASSWORD=opentrusty_test_password
export DB_NAME=opentrusty_test
export DB_SSLMODE=disable
export SERVER_HOST=localhost
export SERVER_PORT=8090
export OPENID_KEY_ENCRYPTION_KEY=12345678901234567890123456789012

# Bootstrap admin credentials (platform admin will be auto-created)
export OT_BOOTSTRAP_ADMIN_EMAIL="admin@platform.local"

# Demo App environment
export CLIENT_ID=""
export CLIENT_SECRET=""
export REDIRECT_URI="http://localhost:8081/callback"
export AUTH_URL="http://localhost:8090"

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}Cleaning up...${NC}"
    
    if [ ! -z "$DEMO_APP_PID" ]; then
        kill $DEMO_APP_PID 2>/dev/null || true
    fi
    
    if [ ! -z "$CONTROL_PANEL_PID" ]; then
        kill $CONTROL_PANEL_PID 2>/dev/null || true
    fi
    
    if [ ! -z "$OPENTRUSTY_PID" ]; then
        kill $OPENTRUSTY_PID 2>/dev/null || true
    fi
    
    cd "$SCRIPT_DIR"
    docker compose -f docker-compose.local-test.yml down -v 2>/dev/null || true
    
    echo -e "${GREEN}Cleanup complete.${NC}"
}

trap cleanup EXIT

# Header
print_header() {
    echo ""
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║       OpenTrusty Local System Validation Test Suite          ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

# Check prerequisites
check_prerequisites() {
    echo -e "${BLUE}[0/7] Checking prerequisites...${NC}"
    
    command -v docker >/dev/null 2>&1 || { echo -e "${RED}Docker required but not installed.${NC}"; exit 1; }
    command -v go >/dev/null 2>&1 || { echo -e "${RED}Go required but not installed.${NC}"; exit 1; }
    command -v node >/dev/null 2>&1 || { echo -e "${RED}Node.js required but not installed.${NC}"; exit 1; }
    command -v curl >/dev/null 2>&1 || { echo -e "${RED}curl required but not installed.${NC}"; exit 1; }
    command -v jq >/dev/null 2>&1 || { echo -e "${RED}jq required but not installed.${NC}"; exit 1; }
    
    echo -e "${GREEN}  ✓ All prerequisites met${NC}"
}

# Start PostgreSQL
start_postgres() {
    echo -e "\n${BLUE}Starting PostgreSQL (clean volume)...${NC}"
    
    cd "$SCRIPT_DIR"
    docker compose -f docker-compose.local-test.yml down -v 2>/dev/null || true
    docker compose -f docker-compose.local-test.yml up -d
    
    # Wait for healthy
    echo -n "  Waiting for PostgreSQL..."
    for i in {1..30}; do
        if docker compose -f docker-compose.local-test.yml exec -T postgres pg_isready -U opentrusty -d opentrusty_test >/dev/null 2>&1; then
            echo -e " ${GREEN}ready${NC}"
            return 0
        fi
        sleep 1
        echo -n "."
    done
    echo -e " ${RED}timeout${NC}"
    exit 1
}

# Build OpenTrusty
build_opentrusty() {
    echo -e "\n${BLUE}Building OpenTrusty...${NC}"
    cd "$PROJECT_ROOT"
    go build -o bin/opentrusty ./cmd/server
    echo -e "${GREEN}  ✓ Build complete${NC}"
}

# Run migrations
run_migrations() {
    echo -e "\n${BLUE}Running migrations...${NC}"
    cd "$PROJECT_ROOT"
    ./bin/opentrusty migrate
    echo -e "${GREEN}  ✓ Migrations applied${NC}"
}

# Start OpenTrusty server
start_opentrusty() {
    echo -e "\n${BLUE}Starting OpenTrusty server...${NC}"
    cd "$PROJECT_ROOT"
    ./bin/opentrusty serve all > "$REPORTS_DIR/opentrusty.log" 2>&1 &
    OPENTRUSTY_PID=$!
    
    # Wait for health
    echo -n "  Waiting for server..."
    for i in {1..30}; do
        if curl -sf http://localhost:8090/health >/dev/null 2>&1; then
            echo -e " ${GREEN}ready${NC}"
            return 0
        fi
        sleep 1
        echo -n "."
    done
    echo -e " ${RED}timeout${NC}"
    cat "$REPORTS_DIR/opentrusty.log"
    exit 1
}

# Start Control Panel
start_control_panel() {
    echo -e "\n${BLUE}Starting Control Panel...${NC}"
    
    if [ ! -d "$CONTROL_PANEL_DIR" ]; then
        echo -e "${YELLOW}  ⚠ Control Panel not found at $CONTROL_PANEL_DIR, skipping...${NC}"
        return 0
    fi
    
    cd "$CONTROL_PANEL_DIR"
    npm run dev > "$REPORTS_DIR/control-panel.log" 2>&1 &
    CONTROL_PANEL_PID=$!
    
    # Wait for ready
    echo -n "  Waiting for Control Panel..."
    for i in {1..60}; do
        if curl -sf http://localhost:5173/admin/ >/dev/null 2>&1; then
            echo -e " ${GREEN}ready${NC}"
            return 0
        fi
        sleep 1
        echo -n "."
    done
    echo -e " ${RED}timeout${NC}"
    exit 1
}

# Start Demo App (but don't fail if it doesn't start yet - needs client credentials)
start_demo_app() {
    echo -e "\n${BLUE}Building Demo App...${NC}"
    
    if [ ! -d "$DEMO_APP_DIR" ]; then
        echo -e "${YELLOW}  ⚠ Demo App not found at $DEMO_APP_DIR, skipping...${NC}"
        return 0
    fi
    
    cd "$DEMO_APP_DIR"
    go build -o demo-app .
    echo -e "${GREEN}  ✓ Demo App built${NC}"
}

# Run a test layer
run_layer() {
    local layer_num=$1
    local layer_name=$2
    local layer_script="$SCRIPT_DIR/layers/${layer_num}-${layer_name}.sh"
    
    echo -e "\n${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}[$layer_num/7] Layer: $layer_name${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    if [ -f "$layer_script" ]; then
        chmod +x "$layer_script"
        if "$layer_script"; then
            echo -e "\n${GREEN}  ✓ Layer $layer_num PASSED${NC}"
            return 0
        else
            echo -e "\n${RED}  ✗ Layer $layer_num FAILED${NC}"
            return 1
        fi
    else
        echo -e "${YELLOW}  ⚠ Script not found: $layer_script${NC}"
        return 1
    fi
}

# Main execution
main() {
    print_header
    
    # Initialize reports directory
    mkdir -p "$REPORTS_DIR"
    rm -f "$REPORTS_DIR"/*.md "$REPORTS_DIR"/*.log 2>/dev/null || true
    
    # Setup
    check_prerequisites
    start_postgres
    build_opentrusty
    run_migrations
    start_opentrusty
    start_control_panel
    start_demo_app
    
    # Run all layers
    PASSED=0
    FAILED=0
    
    if run_layer "01" "infra-smoke"; then PASSED=$((PASSED+1)); else FAILED=$((FAILED+1)); fi
    if run_layer "02" "bootstrap"; then PASSED=$((PASSED+1)); else FAILED=$((FAILED+1)); fi
    if run_layer "03" "tenant-lifecycle"; then PASSED=$((PASSED+1)); else FAILED=$((FAILED+1)); fi
    if run_layer "04" "oauth-client"; then PASSED=$((PASSED+1)); else FAILED=$((FAILED+1)); fi
    if run_layer "05" "oidc-e2e"; then PASSED=$((PASSED+1)); else FAILED=$((FAILED+1)); fi
    if run_layer "06" "audit"; then PASSED=$((PASSED+1)); else FAILED=$((FAILED+1)); fi
    if run_layer "07" "isolation-negative"; then PASSED=$((PASSED+1)); else FAILED=$((FAILED+1)); fi
    
    # Summary
    echo ""
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║                        TEST SUMMARY                          ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "  Passed: ${GREEN}$PASSED${NC}"
    echo -e "  Failed: ${RED}$FAILED${NC}"
    echo -e "  Reports: $REPORTS_DIR"
    echo ""
    
    if [ $FAILED -eq 0 ]; then
        echo -e "${GREEN}╔══════════════════════════════════════════════════════════════╗${NC}"
        echo -e "${GREEN}║               🎉 ALL TESTS PASSED - BETA READY 🎉            ║${NC}"
        echo -e "${GREEN}╚══════════════════════════════════════════════════════════════╝${NC}"
        exit 0
    else
        echo -e "${RED}╔══════════════════════════════════════════════════════════════╗${NC}"
        echo -e "${RED}║                    ❌ TESTS FAILED                           ║${NC}"
        echo -e "${RED}╚══════════════════════════════════════════════════════════════╝${NC}"
        exit 1
    fi
}

# Parse args
if [ "$1" == "--no-tests" ]; then
    print_header
    check_prerequisites
    start_postgres
    build_opentrusty
    run_migrations
    start_opentrusty
    start_control_panel
    start_demo_app
    echo -e "\n${GREEN}Environment ready. Press Ctrl+C to stop.${NC}"
    wait
else
    main
fi
