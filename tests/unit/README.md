# Unit Tests

This directory contains cross-package unit tests and documentation on test organization.

## Organization Policy

1. **Package Unit Tests**: Standard Go unit tests (`*_test.go`) should remain within the package they test to allow testing of internal (unexported) members.
2. **Cross-Package Integration**: Use `tests/unit` ONLY for tests that span multiple `internal` packages but do NOT require external dependencies (DB, network).
3. **Build Tags**:
   - `integration`: Tests requiring a running database.
   - `e2e`: Tests requiring the full application stack running.

## Running Tests

```bash
# Run unit tests only (CI Gate)
make test-unit

# Run all tests (requires environment)
make test-all
```
