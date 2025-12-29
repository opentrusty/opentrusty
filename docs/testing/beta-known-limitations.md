# Beta Known Limitations

This document tracks known architectural and functional limitations identified during the Stage 6.5 Beta Validation.

## 1. Protocol & Security

- **Hardcoded Issuer**: The OIDC `issuer` is currently hardcoded to `http://localhost:8080` in `cmd/server/main.go`. This must be configurable via environment variables before production release.
- **SSL/TLS**: Local development runs on HTTP. Production environments MUST use a reverse proxy (Nginx/Traefik) for TLS termination.
- **Token Signing Key Rotation**: There is currently no automated mechanism for signing key rotation without system downtime.
- **PKCE Optionality**: While PKCE is encouraged and implemented, the current handler may not strictly block all clients if they don't provide it, depending on client configuration (Beta verification uses mandatory PKCE).

## 2. Management & UI

- **Audit Log Visibility**: Audit logs are recorded in the database but have no dedicated UI view in the Control Panel yet.
- **Client Secret Recovery**: If a `client_secret` is lost after initial registration, it cannot be "viewed" again (by design); it must be regenerated.
- **Branding**: Login and consent pages use a default OpenTrusty theme. Tenant-specific branding APIs are defined but not fully wired to the UI.

## 3. Operations

- **Session Cleanup**: A background goroutine cleans up expired sessions, but there is no fine-grained control over session lifetime per tenant yet.
- **Database Migrations**: Migrations are applied manually via `migrate` command. There is no automated migration-on-startup feature for production safety.

---
*Last Updated: Stage 6.5 Beta Validation*
