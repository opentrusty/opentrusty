# Security Review Checklist: Stage 5

## 1. Session & Cookie Security
- [x] `HttpOnly` attribute set for all session cookies.
- [x] `Secure` attribute defaults to `true` for non-localhost environments.
- [x] `SameSite` attribute set to `Lax` (or `Strict` where appropriate).
- [x] Session rotation implemented (old session destroyed on login).
- [x] Logical isolation between Auth and Admin session namespaces.

## 2. CSRF & CORS
- [x] CSRF protection enforced for state-changing Admin API endpoints (via `X-CSRF-Token`).
- [x] CORS policies strictly defined (no `*` in production).
- [x] `X-Tenant-ID` header spoofing rejected on authenticated routes.

## 3. Credential Handling
- [x] Argon2id used for all password hashing.
- [x] `client_secret` and tokens hashed (SHA-256) before storage.
- [x] No plaintext secrets or passwords in logs.
- [x] Critical secrets (Encryption Key) verified at startup (fail-fast).

## 4. OIDC Protocol Compliance
- [x] Authorization codes are one-time use.
- [x] Authorization codes have short expiry (5m).
- [x] Strict exact matching for `redirect_uri`.
- [x] PKCE supported and verified.

## 5. Authorization & Multi-Tenancy
- [x] Cross-tenant lateral isolation verified in `HasPermission`.
- [x] Platform vs Tenant vertical isolation verified.
- [x] Tenant context derived exclusively from server-side session.

## 6. Observability
- [x] Audit logs implemented for all state-changing operations.
- [x] Audit logs capture Actor, Action, Target, and Timestamp.
- [x] No PII in audit log metadata.
