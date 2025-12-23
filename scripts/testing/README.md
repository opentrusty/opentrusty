# OpenTrusty Structured Test Reporting

This directory contains the tools for generating structured, professional test reports from Go test execution.

## Pipeline Overview

1. **Test Execution**: Tests are run using `go test -json`, which produces machine-readable output events.
2. **Metadata Extraction**: The `report_gen.go` tool parses the Go source code to extract structured annotations from test doc comments (e.g., `TestPurpose`, `Security`, `Expected`).
3. **Correlation**: The tool merges the execution results (pass/fail/time) with the source code metadata.
4. **Report Generation**: The tool generates both a detailed JSON report for machine consumption and a formatted Markdown report for human review.

## Annotation Standard

Tests should be annotated as follows to be included in reports correctly:

```go
// TestPurpose: Description of what is being tested.
// Scope: Unit Test | Service Test | Integration Test
// Security: Security invariant being protected (optional).
// Expected: Expected outcome of the test.
// Test Case ID: External ID for tracking (optional).
func TestComponent_Scenario_Behavior(t *testing.T) { ... }
```

## How to Generate Reports

Use the provided `Makefile` targets:

```bash
# Generate all reports (Unit and System)
make test-reports

# Generate specifically Unit Test reports
make test-report-ut

# Generate specifically System Test reports
make test-report-st
```

## Output Locations

Reports are saved to `artifacts/tests/`:
- `ut-report.json` / `ut-report.md`: Unit test results.
- `st-report.json` / `st-report.md`: System test results.
