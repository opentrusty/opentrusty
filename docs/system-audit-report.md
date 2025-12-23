# System Audit Report - OpenTrusty

**Generated:** 2025-12-23
**Type:** Read-Only Static Analysis

This report identifies security and architectural issues across the codebase.

---

## A. AuthN (Authentication)

| ID | Location | Finding |
|----|----------|---------|
| A.1 | `handlers.go:414` (Logout) | Bypasses `AuthMiddleware`; manually checks cookie instead of using centralized session validation. |
| A.2 | `oauth2_handler.go:89-96` (Authorize) | Checks `userID` manually after middleware should have validated. Redundant check creates maintenance risk. |
| A.3 | `oidc_handler.go:30` (Discovery) | Publicly accessible. Expected for OIDC, but note for DDoS risk assessment. |
| A.4 | `oidc_handler.go:53` (JWKS) | Publicly accessible. Expected for OIDC, but note for DDoS risk assessment. |

---

## B. AuthZ (Authorization)

| ID | Location | Finding |
|----|----------|---------|
| B.1 | `tenant_handler.go:202-220` (AssignTenantRole) | **CRITICAL.** No `HasPermission` check. Any authenticated user can assign roles. |
| B.2 | `oauth2_client_handler.go:54-106` (RegisterClient) | **CRITICAL.** No `HasPermission` check. Any authenticated user can register OAuth2 clients for any tenant. |
| B.3 | `handlers.go:461,495` (GetCurrentUser, GetProfile) | Uses `ScopePlatform` for self-service profile reads. May not be intended if profile access should be tenant-scoped. |
| B.4 | `tenant/service.go:114` (RevokeRole) | Cannot determine `ActorID` for audit log without context enrichment. |

---

## C. Tenant Isolation

| ID | Location | Finding |
|----|----------|---------|
| C.1 | `oauth2/models.go:129-140` (AccessToken) | **HIGH.** Missing `TenantID` field. Tokens are not bound to a tenant at the data model level. |
| C.2 | `oauth2/models.go:148-159` (RefreshToken) | **HIGH.** Missing `TenantID` field. Same as C.1. |
| C.3 | `store/postgres/token_repository.go` | No `tenant_id` column in `access_tokens`/`refresh_tokens` queries. Isolation relies solely on hash uniqueness. |
| C.4 | `store/postgres/user_repository.go:82` (GetByID) | Retrieves user by ID without `tenant_id` constraint. Global lookup. |
| C.5 | `store/postgres/client_repository.go:150` (GetByID) | Retrieves client by ID without `tenant_id` constraint. Global lookup. |
| C.6 | `oauth2/service.go:222` (ExchangeCodeForToken) | Does not verify that `code.UserID` belongs to the tenant of the `client`. |
| C.7 | `identity/service.go:348` (GetUser) | Retrieves user by ID globally; does not verify caller's tenant. |

---

## D. API Validation

| ID | Location | Finding |
|----|----------|---------|
| D.1 | `tenant_handler.go:64,216,250,280` | Leaks `err.Error()` directly to HTTP response. May expose internal details. |
| D.2 | `tenant_handler.go:146,151,156,173` | Leaks prefixed error messages (e.g., `"failed to provision user: " + err.Error()`). |
| D.3 | `oauth2_client_handler.go:97` | Leaks `"failed to register client: " + err.Error()`. |
| D.4 | `handlers.go:311` | Leaks `"failed to set password: " + err.Error()`. |
| D.5 | `tenant_handler.go:105` | `tenantID` from URL param used without UUID format validation. |
| D.6 | `tenant_handler.go:204` | `userID` from URL param used without UUID format validation. |

---

## E. Magic Values

| ID | Location | Finding |
|----|----------|---------|
| E.1 | `identity/bootstrap.go:66` | Hardcoded UUID `"20000000-0000-0000-0000-000000000001"` for `platform_admin` role. |
| E.2 | `tenant/tenant.go:31` | `DefaultTenantID = "default"`. **Violates anti-pattern rule.** Should not exist or be used. |
| E.3 | `oauth2_client_handler.go:80-82` | Hardcoded token lifetimes: `AccessTokenLifetime: 3600`, `RefreshTokenLifetime: 2592000`, `IDTokenLifetime: 3600`. |
| E.4 | `handlers.go:626` | Hardcoded cookie `MaxAge: 86400`. |
| E.5 | `oauth2/service.go:71` | Fallback encryption key: `"insecure_dev_key_must_change_!!"`. |
| E.6 | `oauth2/service.go:193` | Hardcoded authorization code lifetime: `5 * time.Minute`. |
| E.7 | `ratelimit.go:40` | Hardcoded cleanup interval: `10 * time.Minute`. |
| E.8 | `handlers.go:127` | Hardcoded request timeout: `60 * time.Second`. |

---

## F. Domain Boundary

| ID | Location | Finding |
|----|----------|---------|
| F.1 | `oauth2/service.go:319-328` | `OIDCProvider` hook is optional; silent failure mode if ID token generation fails. Could mask production issues. |
| F.2 | `transport/http/handlers.go` | Handler layer contains router setup (`NewRouter`). Violates separation; router config should be in a dedicated file. |
| F.3 | `tenant/service.go` vs `authz/service.go` | Two separate role assignment paths: one for tenant roles, one for RBAC. Potential for inconsistency. |

---

## G. Observability & Safety

| ID | Location | Finding |
|----|----------|---------|
| G.1 | `ratelimit.go:69-72` | Cleanup wipes entire IP map. Can cause bursts of re-computation and potential bypass mid-cycle. |
| G.2 | `tenant/service.go:119-128` | `RevokeRole` audit log cannot reliably determine `ActorID` (revoker). |
| G.3 | `oauth2/service.go:312-314` | `refreshRepo.Create` error is swallowed (`if err == nil`). Silent failure mode. |
| G.4 | `session/service.go:112` | `generateSessionID` ignores error from `rand.Read`. Potential for weak session IDs on system entropy failure. |
| G.5 | `identity/service.go:395-405` | `isValidEmail` and `isStrongPassword` are simplistic. RFC 5321 violation risk for email; weak password policy. |

---

## Summary Statistics

| Category | Critical | High | Medium | Low |
|----------|----------|------|--------|-----|
| AuthN    | 0        | 0    | 1      | 3   |
| AuthZ    | 2        | 0    | 2      | 0   |
| Tenant   | 2        | 3    | 2      | 0   |
| API      | 0        | 2    | 4      | 0   |
| Magic    | 0        | 2    | 6      | 0   |
| Domain   | 0        | 0    | 3      | 0   |
| Safety   | 0        | 1    | 4      | 0   |
| **Total**| **4**    | **8**| **22** | **3**|
