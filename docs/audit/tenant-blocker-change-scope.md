# Change Scope: B-TENANT-01 (Implicit Tenant Context)

This document outlines the specific code locations affected by the fix for blocker **B-TENANT-01**, based on the **Tenant Resolution Contract**.

## 1. Files & Functions to Change

### REQUIRED (Immediate Fix)

#### [internal/transport/http/middleware.go](file:///Users/mw/workspace/repo/github.com/opentrusty/opentrusty/internal/transport/http/middleware.go)
- **[NEW] `TenantMiddleware`**: 
    - Implements logic to extract `tenant_id` from `X-Tenant-ID` header or query parameter. 
    - Rejects request with `400 Bad Request` if missing.
    - Injects `TenantID` into context.
- **[MODIFY] `AuthMiddleware`**: 
    - Updated to verify that the `TenantID` in the session matches the `TenantID` in the request-scoped context (Cross-tenant session prevention).

#### [internal/transport/http/handlers.go](file:///Users/mw/workspace/repo/github.com/opentrusty/opentrusty/internal/transport/http/handlers.go)
- **[MODIFY] `NewRouter`**: 
    - Inject `TenantMiddleware` into the global or group-level middleware stack.
- **[DELETE] `getTenantID`**: 
    - Remove the helper that defaults to `"default"`.
- **[MODIFY] `Register`, `Login`, `Logout`**: 
    - Extract `tenant_id` from `r.Context()` instead of calling `getTenantID`.

#### [internal/transport/http/oauth2_handler.go](file:///Users/mw/workspace/repo/github.com/opentrusty/opentrusty/internal/transport/http/oauth2_handler.go)
- **[MODIFY] `Authorize`, `Token`, `UserInfo`**: 
    - Ensure these endpoints are wrapped by `TenantMiddleware`.
    - Extract `tenant_id` from context for all service calls.

---

## 2. What MUST NOT be touched

- **`internal/identity/service.go`**: Service logic already accepts `tenantID` as an argument.
- **`internal/store/postgres/*`**: Persistence layers already correctly scope by `tenant_id`.
- **`cmd/server/main.go`**: Service wiring is unaffected.

---

## 3. Classification

| Change | Classification | Rationale |
| :--- | :--- | :--- |
| **Strict Header Enforcement** | **REQUIRED** | Core fix for B-TENANT-01. |
| **Middleware Isolation** | **REQUIRED** | Satisfies Tenant Resolution Contract. |
| **Subdomain-based Resolution** | **OPTIONAL** | Forbidden in this phase (Scope Crep). |
| **Database-backed Tenant Validation** | **OPTIONAL** | Can be deferred to Phase 2 (Isolation Verification). |

---
**Rule Citation**: These changes satisfy **Rule 2.1 (Isolation)** and **Rule 2.3 (No Cross-Talk)**.
