# CSRF/CORS Defense Model

## 1. CSRF (Cross-Site Request Forgery)

### Strategy
- **Admin API**: Use Double Submit Cookie pattern or a custom `X-CSRF-Token` header. Since the Control Panel is a SPA, we will implement a custom header approach.
- **Login/Consent**: Use traditional CSRF tokens in forms.

### Endpoint Defense Table

| Endpoint | Method | Defense | Rationale |
|----------|--------|---------|-----------|
| `/api/v1/auth/login` | POST | CSRF Token | Form-based login |
| `/api/v1/auth/logout` | POST | CSRF Token | Prevent forced logout |
| `/api/v1/tenants/*` | POST/PUT/DELETE | `X-CSRF-Token` Header | SPA Management API |
| `/api/v1/user/*` | POST/PUT | `X-CSRF-Token` Header | SPA Management API |
| `/oauth2/token` | POST | None (Protocol) | Client authentication required |
| `/oauth2/authorize` | GET | `state` param | OIDC protocol protection |

## 2. CORS (Cross-Origin Resource Sharing)

### Strategy
- **Auth Plane**: Allow `*` for OIDC Discovery and JWKS. Restrict Authorization/Token endpoints to registered `redirect_uris`.
- **Admin Plane**: ONLY allow the Control Panel's domain (e.g., `console.opentrusty.org`). NO `*` allowed.

### CORS Configuration

| Plane | Allowed Origins | Allowed Methods | Credentials |
|-------|-----------------|-----------------|-------------|
| **Auth** | `*` (Discovery/JWKS) | GET | No |
| **Auth** | Restricted (Redirect URIs) | POST | No |
| **Admin** | `console.opentrusty.org` | GET, POST, PUT, DELETE | **Yes** |
