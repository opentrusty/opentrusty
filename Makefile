
# OpenTrusty Makefile

.PHONY: all build run test clean dev docs-gen

APP_NAME := opentrusty
CMD_PATH := ./cmd/opentrusty
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
test:
	@echo "Running tests..."
	@go test -v ./...

# Benchmark
bench:
	@echo "Running benchmarks..."
	@go test -bench=. ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

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
	swag init -g internal/transport/http/handlers.go --parseDependency --parseInternal --output docs/api
