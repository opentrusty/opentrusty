#!/bin/bash
set -e

# run-e2e.sh
# Orchestrates E2E tests in a Docker environment.

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../../.." &> /dev/null && pwd )"

cd "$SCRIPT_DIR"

# Determine Docker Compose command
if docker compose version > /dev/null 2>&1; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

echo "--- Starting E2E Docker Environment ---"
$DOCKER_COMPOSE -f docker-compose.test.yml up -d --build

# Cleanup function
cleanup() {
    echo "--- Tearing down E2E Docker Environment ---"
    $DOCKER_COMPOSE -f "$SCRIPT_DIR/docker-compose.test.yml" down -v
}
# trap cleanup EXIT

echo "--- Waiting for OpenTrusty to be ready ---"
MAX_RETRIES=30
COUNT=0
until $(curl -sf http://localhost:8080/health > /dev/null); do
    if [ $COUNT -eq $MAX_RETRIES ]; then
      echo "Timeout waiting for service"
      docker-compose -f docker-compose.test.yml logs opentrusty_test
      exit 1
    fi
    echo "Waiting... ($COUNT/$MAX_RETRIES)"
    sleep 2
    COUNT=$((COUNT+1))
done
sleep 5
docker ps -a --filter name=opentrusty
docker logs docker-opentrusty_test-1
echo "--- Running Database Migrations ---"
docker exec docker-opentrusty_test-1 ./opentrusty migrate

echo "--- Running Go E2E Tests ---"
cd "$PROJECT_ROOT"
export OPENTRUSTY_API_URL="http://localhost:8080"
go test "$@" ./tests/e2e/ --tags=e2e

echo "--- E2E Tests Passed! ---"
