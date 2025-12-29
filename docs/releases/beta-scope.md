# Beta Release Scope

This document defines the **binding capability freeze** for the OpenTrusty Beta release.

---

## Included Capabilities

### Platform Admin
- [x] Platform Overview with dynamic metrics
- [x] Tenant list and creation
- [x] Tenant user provisioning
- [x] Platform admin list (view only)
- [x] Platform audit logs
- [x] System settings display (OIDC endpoints)

### Tenant Admin
- [x] Tenant overview with OIDC endpoints
- [x] User list and provisioning
- [x] OAuth client registration wizard
- [x] Client credentials display (one-time)
- [x] Tenant audit logs
- [x] Branding placeholder

### OAuth2 / OIDC
- [x] Client Credentials grant
- [x] Authorization Code grant (backend only, no interactive UI)
- [x] JWKS endpoint
- [x] Discovery endpoint

---

## Explicit Exclusions

The following are **NOT** included in Beta:

| Feature | Reason |
|---------|--------|
| Interactive OIDC login UI | Auth Plane UI not implemented |
| Password reset flow | Deferred to post-Beta |
| Email verification | Deferred to post-Beta |
| MFA/2FA | Deferred to GA |
| Tenant branding customization | UI placeholder only |
| Platform admin invitation | API exists, UI limited |
| Refresh token rotation | Deferred to post-Beta |
| PKCE enforcement | Optional in Beta |

---

## Known Limitations

1. **OIDC Authorization Flow**: The `/oauth2/authorize` endpoint returns 401 JSON for unauthenticated requests instead of redirecting to a login page.

2. **Single Binary Mode**: Backend runs as a single process; separate auth/admin planes not yet independently deployable.

3. **No HA Support**: Single-node deployment only.

4. **Limited Audit Filtering**: Audit log UI does not support filtering by date or event type.

---

## Non-Goals for Beta

- Kubernetes deployment manifests
- Helm charts
- Terraform modules
- Multi-region replication
- Backup/restore procedures
- Performance benchmarks

---

## Binding Until

This scope is binding until **General Availability (GA)** or explicit amendment approved by project maintainers.
