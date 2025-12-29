# Beta Test Matrix

This document maps the capabilities defined in `docs/_ai/capabilities.md` to specific technical tests (Unit, System, and E2E-UI).

## 1. Platform Admin UI

| Test ID | Type | Capability | Execution | Expected Result | Status |
|---------|------|------------|-----------|-----------------|--------|
| UI-01 | E2E | Initial Admin Login | Playwright (`01-bootstrap`) | Dashboards visible after login with bootstrap credentials. | PASS |
| UI-02 | E2E | Tenant Creation | Playwright (`02-tenant`) | Tenant and initial user created and visible. | PASS |
| UI-03 | E2E | Tenant Isolation | Playwright (`02-tenant`) | Admins can only see data for their scope. | PASS |
| UI-09 | E2E | Platform Overview Metrics | Playwright (`06-platform-features`) | Dynamic counts for tenants/users/clients visible. | PASS |
| UI-10 | E2E | Platform Admin List | Playwright (`06-platform-features`) | List of platform admins is accessible. | PASS |
| UI-11 | E2E | System Settings OIDC | Playwright (`06-platform-features`) | Discovery, Issuer, and JWKS URLs visible. | PASS |
| UI-12 | E2E | Platform Navigation | Playwright (`06-platform-features`) | Sidebar links resolve to correct platform views. | PASS |
| UI-13 | E2E | Platform Audit Logs | Playwright (`06-platform-features`) | Platform-wide activity table is visible. | PASS |

## 2. Tenant Owner/Admin UI

| Test ID | Type | Capability | Execution | Expected Result | Status |
|---------|------|------------|-----------|-----------------|--------|
| UI-04 | E2E | Register OIDC Client | Playwright (`03-client`) | Client ID and Secret generated and shown. | PASS |
| UI-08 | E2E | Audit Log Verification | Playwright (`05-audit-v2`) | `client_created` and `login_success` events logged. | PASS |
| UI-14 | E2E | Tenant Overview Data | Playwright (`07-tenant-features`) | Tenant-specific metrics and OIDC endpoints shown. | PASS |
| UI-15 | E2E | Client Wizard Access | Playwright (`07-tenant-features`) | Create Client wizard opens correctly. | PASS |
| UI-16 | E2E | Branding Placeholder | Playwright (`07-tenant-features`) | "Coming Soon" styling for branding page. | PASS |
| UI-17 | E2E | Tenant User List | Playwright (`07-tenant-features`) | Table of users within the tenant is visible. | PASS |
| UI-18 | E2E | Tenant Navigation | Playwright (`07-tenant-features`) | Sidebar links resolve correctly in tenant context. | PASS |

## 3. OAuth2 / OIDC Protocol

| Test ID | Type | Capability | Execution | Expected Result | Status |
|---------|------|------------|-----------|-----------------|--------|
| UI-05 | E2E | Full OIDC Flow | Playwright (`04-oidc-flow`) | Third-party app authenticates via OpenTrusty. | **SKIPPED** |

> [!NOTE] 
> **Skipped: UI-05**
> **Reason:** Backend lacks an interactive login page for the auth plane. The `/oauth2/authorize` endpoint currently returns a 401 JSON error instead of redirecting to a UI.
> **Enabled In:** Next major release following the implementation of the Auth Plane UI.

---

## Reproduction Steps

To execute the automated E2E-UI suite:

1.  **Ensure Backend is Running:**
    ```bash
    # In opentrusty repo
    make run
    ```
2.  **Run Playwright Tests:**
    ```bash
    # In opentrusty-control-panel repo
    make e2e
    ```

Check `artifacts/tests/ui/` for HTML reports and artifacts.
