# Release Gates & Testing Standards

## Overview
OpenTrusty enforces a strict "No Tests, No Release" policy. This document defines the automated gates that must pass for any code to merge into `main` or be released as a versioned artifact.

## CI Gates (Pull Requests)
All Pull Requests must pass the following checks before merging.

### Core Repository (`opentrusty`)
-   **Unit Tests**: `make test-unit` (Go domain logic must pass).
-   **Integration Tests**: `make test-integration` (Postgres interactions must pass).
-   **Security Scan**: `gosec` static analysis.

### Control Panel Repository (`opentrusty-control-panel`)
-   **Static Analysis**: `npm run lint` (ESLint).
-   **Build Verification**: `npm run build` (TypeScript Type Check).

## Release Gates (Tags `v*`)
All Release Tags trigger a full validation suite. A release is considered "broken" if any of these fail.

### Core Repository (`opentrusty`)
-   **All CI Gates** (Unit, Integration, Security).
-   **E2E Tests**: `make test-e2e` (Full Dockerized flow verification).
-   **Documentation**: API docs must match code (`swag init` freshness check).

### Control Panel Repository (`opentrusty-control-panel`)
-   **All CI Gates** (Lint, Build).
-   **Start-to-Finish Smoke Tests**: `npx playwright test` (Verifies Critical User Journeys).

## Reporting Standards
-   **Unit/Integration**: Output as structured Markdown/HTML in CI artifacts.
-   **E2E (Core)**: Docker logs and structured test results attached to Release.
-   **E2E (UI)**: Playwright HTML Report attached to Release.

## Exceptions
-   **Alpha/Beta Releases**: May relax *Performance* or *Systemd* specific smoke tests, but MUST pass functional E2E.
-   **RC/GA Releases**: MUST pass ALL gates without exception.
