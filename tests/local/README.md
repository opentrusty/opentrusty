# Local System Validation Test Suite

Automated test suite for OpenTrusty Beta validation.

## Prerequisites

- Docker (with Compose V2)
- Go 1.21+
- Node.js 18+
- curl
- jq

## Quick Start

```bash
# Run all tests (from opentrusty repo root)
./tests/local/run-all.sh

# Start environment without running tests (for manual testing)
./tests/local/run-all.sh --no-tests
```

## Test Layers

| Layer | Name | Description |
|-------|------|-------------|
| 1 | infra-smoke | Verify all services start and respond |
| 2 | bootstrap | Platform admin auto-provisioning |
| 3 | tenant-lifecycle | Tenant creation and admin login |
| 4 | oauth-client | OAuth2 client registration |
| 5 | oidc-e2e | Full OIDC authorization code flow |
| 6 | audit | Audit log verification |
| 7 | isolation-negative | Security boundary tests |

## Ports

| Service | Port |
|---------|------|
| PostgreSQL | 5434 |
| OpenTrusty | 8080 |
| Control Panel | 5173 |
| Demo App | 8081 |

## Reports

Generated in `tests/local/reports/`:
- `infra-smoke.md`
- `bootstrap-test.md`
- `tenant-lifecycle.md`
- `oauth-client-test.md`
- `oidc-e2e.md`
- `audit-verification.md`
- `isolation-negative.md`

## Expected Duration

~5-10 minutes for full suite.

## Cleanup

Automatic cleanup on exit. Manual cleanup:
```bash
cd tests/local
docker compose -f docker-compose.local-test.yml down -v
```
