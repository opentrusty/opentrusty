# Test Report Schema

**Status**: Normative
**Owner**: OpenTrusty Maintainers

This document defines the canonical JSON schema for all test reports generated in OpenTrusty. All test runners (Unit, Service, E2E) MUST output results adhering strictly to this schema to ensure interoperability with the Release Gates validation infrastructure.

## 1. Schema Definition

All test reports must be valid JSON objects matching the following structure.

### 1.1 Root Object

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `test_type` | `string` | **YES** | One of: `"unit"`, `"integration"`, `"e2e"`, `"system"` |
| `suite_name` | `string` | **YES** | Logical name of the test suite (e.g., `"Core Auth"`, `"Postgres Store"`) |
| `total_tests` | `integer` | **YES** | Total number of tests executed |
| `passed` | `integer` | **YES** | Number of passing tests |
| `failed` | `integer` | **YES** | Number of failing tests |
| `skipped` | `integer` | **YES** | Number of skipped tests |
| `duration_ms` | `integer` | **YES** | Total execution time in milliseconds |
| `environment` | `object` | **YES** | Metadata about the runtime environment (see 1.2) |
| `source` | `object` | **YES** | Metadata about the code being tested (see 1.3) |
| `timestamp_utc` | `string` | **YES** | Execution completion time in ISO 8601 format (UTC) |
| `failures` | `array` | **YES** | List of failure objects (see 1.4). Empty array if no failures. |

### 1.2 Environment Object (`environment`)

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `os` | `string` | **YES** | Operating System (e.g., `"linux"`, `"darwin"`, `"windows"`) |
| `arch` | `string` | **YES** | Architecture (e.g., `"amd64"`, `"arm64"`) |
| `runtime` | `string` | **YES** | Runtime version (e.g., `"go1.21.0"`, `"docker-24.0.5"`) |
| `ci_job_url` | `string` | NO | Link to the CI run logs (if available) |

### 1.3 Source Object (`source`)

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `git_commit` | `string` | **YES** | Full SHA-1 hash of the commit being tested |
| `git_tag` | `string` | **YES** | Tag name (e.g., `"v1.0.0-rc.1"`) or `"none"` |
| `branch` | `string` | NO | Branch name (e.g., `"main"`, `"feat/oauth-fix"`) |

### 1.4 Failure Object (`failures[]`)

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `test_name` | `string` | **YES** | Full name of the test (e.g., `"TestOAuth2_AuthCode_Success"`) |
| `category` | `string` | **YES** | Logical grouping (e.g., `"Security"`, `"Database"`, `"Unknown"`) |
| `reason` | `string` | **YES** | Sanitized error message. **MUST NOT** contain PII or secrets. |
| `location` | `string` | NO | File and line number (e.g., `"internal/oauth2/handler_test.go:45"`) |

## 2. Validation Rules

1.  **Strictness**: Additional fields are allowed but ignored. Missing required fields render the report **INVALID**, causing the Release Gate to fail immediately.
2.  **Sanitization**: The `reason` field in failures must be sanitized to remove sensitive data (API keys, passwords, PII) before generation.
3.  **Determinism**: Fields like `timestamp_utc` must use a consistent format (ISO 8601) to allow correct sorting and parsing.

## 3. Artifact Naming Conventions

To ensure CI pipelines can automatically detect and process reports, files MUST form the following naming convention:

| Test Type | Allowed Filenames |
| :--- | :--- |
| Unit Tests | `test-report-unit.json` |
| Integration Tests | `test-report-integration.json` |
| System Tests | `test-report-system.json` |
| E2E Tests | `test-report-e2e.json` |
| Security Scan | `test-report-security.json` |

Any file adhering to this naming convention will be automatically picked up by the `scripts/generate-test-report.sh` rendering tool.
