# Production Readiness

> [!WARNING]
> OpenTrusty is **NOT** secure by default for public internet exposure without specific configuration and infrastructure.
> You **MUST** address the checklist below before deploying to production.

## Checklist

### 1. Infrastructure (Mandatory)
- [ ] **TLS Termination**: OpenTrusty serves plain HTTP using standard `net/http`. You **MUST** run it behind a reverse proxy (Nginx, Caddy, AWS ALB) that handles HTTPS.
- [ ] **Database Connection**: Ensure `SSLMode` is set to `require` or `verify-full` in `config.yaml` for production databases. Do NOT use `disable`.
- [ ] **Secret Management**: Pass sensitive values (DB Passwords) via Environment Variables, NOT config files committed to git.

### 2. Configuration (Mandatory)
- [ ] **Establish Trusted Issuer**: The `ISSUER` URL must be the public-facing HTTPS URL (e.g., `https://auth.example.com`). This value is embedded in signed tokens and cannot change.
- [ ] **Key Management**: OpenTrusty currently generates ephemeral RSA keys on startup (Phase II.1).
    - **CRITICAL**: For multi-node deployments, you currently have **NO persistence**.
    - **Status**: **NOT READY for multi-node production**. Single-instance only until Phase III (Key Persistence).

### 3. Application Security
- [ ] **CORS**: Verify CORS headers if your frontend is on a different domain.
- [ ] **Redirect URIs**: Ensure all Client Redirect URIs are HTTPS (except for `localhost` during dev).
- [ ] **Scopes**: Verify that only necessary scopes are allowed for each client.

### 4. Observability
- [/] **Audit Logs**: Enabled via `slog`. Ensure these logs are shipped to a secure SIEM or long-term storage.
- [ ] **Metrics**: No integrated Prometheus metrics exporter yet (Planned Phase IV).

## Supported Architectures

### Recommended
- **Single Instance** (Due to ephemeral keys)
- **PostgreSQL** (Managed Service preferred)
- **Reverse Proxy** (TLS + Rate Limiting)

### Not Supported Yet
- **High Availability / Horizontal Scaling** (Requires Key Persistence)
- **Geo-Replication**
