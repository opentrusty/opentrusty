
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

test-system:
	@echo "Running system integration tests..."
	@INTEGRATION_TEST=true go test -v ./tests/system/...

test-systemd:
	@echo "Running systemd smoke tests..."
	@./tests/e2e/systemd/smoke_test.sh

test-all: test-unit test-integration test-e2e test-systemd

# Test with Structured Reporting
TEST_ARTIFACTS := artifacts/tests
REPORT_GEN := go run scripts/testing/report_gen.go

test-reports: test-report-ut test-report-st test-report-e2e
	@echo "✅ All structured test reports successfully generated in $(TEST_ARTIFACTS)"

test-report-ut:
	@echo "Running unit tests and generating structured reports..."
	@mkdir -p $(TEST_ARTIFACTS)
	@go test -json ./... > $(TEST_ARTIFACTS)/ut-raw.json || true
	@$(REPORT_GEN) -input $(TEST_ARTIFACTS)/ut-raw.json \
		-out-json $(TEST_ARTIFACTS)/ut-report.json \
		-out-md $(TEST_ARTIFACTS)/ut-report.md \
		-out-html $(TEST_ARTIFACTS)/ut-report.html \
		-title "Unit Test Report"
	@rm $(TEST_ARTIFACTS)/ut-raw.json

test-report-st:
	@echo "Running system tests and generating structured reports..."
	@mkdir -p $(TEST_ARTIFACTS)
	@INTEGRATION_TEST=true go test -json ./tests/system/... > $(TEST_ARTIFACTS)/st-raw.json || true
	@$(REPORT_GEN) -input $(TEST_ARTIFACTS)/st-raw.json \
		-out-json $(TEST_ARTIFACTS)/st-report.json \
		-out-md $(TEST_ARTIFACTS)/st-report.md \
		-out-html $(TEST_ARTIFACTS)/st-report.html \
		-title "System Test Report"
	@rm $(TEST_ARTIFACTS)/st-raw.json

test-report-e2e:
	@echo "Running E2E tests and generating structured reports..."
	@mkdir -p $(TEST_ARTIFACTS)
	@./tests/e2e/docker/run-e2e.sh -json > $(TEST_ARTIFACTS)/e2e-raw.json || true
	@$(REPORT_GEN) -input $(TEST_ARTIFACTS)/e2e-raw.json \
		-out-json $(TEST_ARTIFACTS)/e2e-report.json \
		-out-md $(TEST_ARTIFACTS)/e2e-report.md \
		-out-html $(TEST_ARTIFACTS)/e2e-report.html \
		-title "E2E Test Report"
	@rm $(TEST_ARTIFACTS)/e2e-raw.json

docs-test-reports: test-reports
	@echo "Publishing test reports to docs..."
	@mkdir -p docs/testing
	@cp $(TEST_ARTIFACTS)/ut-report.md docs/testing/ut-report.md
	@cp $(TEST_ARTIFACTS)/st-report.md docs/testing/st-report.md
	@cp $(TEST_ARTIFACTS)/e2e-report.md docs/testing/e2e-report.md
	@echo "Test reports published to docs/testing/"


# Benchmark
bench:
	@echo "Running benchmarks..."
	@go test -bench=. ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR) build_docs build_output artifacts
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
