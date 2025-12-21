
# OpenTrusty Makefile

.PHONY: all build run test clean dev docs-gen

APP_NAME := opentrusty
CMD_PATH := ./cmd/server
BUILD_DIR := ./bin

all: build

# Build the binary
build:
	@echo "Building $(APP_NAME)..."
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)

# Run the application
run: build
	@echo "Running $(APP_NAME)..."
	@$(BUILD_DIR)/$(APP_NAME)

# Test
test: test-unit

test-unit:
	@echo "Running unit tests..."
	@go test -v ./...

test-e2e:
	@echo "Running Docker-based E2E tests..."
	@./tests/e2e/docker/run-e2e.sh

test-integration:
	@echo "Running integration tests..."
	@go test -v ./internal/store/postgres/ --tags=integration

test-systemd:
	@echo "Running systemd smoke tests..."
	@./tests/e2e/systemd/smoke_test.sh

test-all: test-unit test-integration test-e2e test-systemd

# Benchmark
bench:
	@echo "Running benchmarks..."
	@go test -bench=. ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR) build_docs build_output
	@rm -f docs/api/docs.go docs/api/swagger.json docs/api/swagger.yaml

# Development Setup (Docker)
dev:
	@echo "Starting development environment..."
	@docker-compose up -d

dev-down:
	@echo "Stopping development environment..."
	@docker-compose down

# Tidy dependencies
tidy:
	@go mod tidy

# Generate OpenAPI Documentation
docs-gen:
	@echo "Generating OpenAPI 3.1 Specification..."
	@if ! command -v swag > /dev/null; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@export PATH=$${PATH}:$$(go env GOPATH)/bin; \
	swag init -g internal/transport/http/handlers.go --parseDependency --parseInternal --output docs/api
