# Auth Plane Architecture

## Purpose
The Auth Plane (`serve auth`) is the security boundary for end-user authentication and protocol compliance. It represents the "Data Plane" of identity.

## Capabilities (Stage 4)

### Endpoints
| Path | Method | Purpose | Auth Required |
|------|--------|---------|---------------|
| `/.well-known/openid-configuration` | GET | OIDC Discovery | No |
| `/oauth2/authorize` | GET | Start OIDC/OAuth2 Flow | via Session |
| `/oauth2/token` | POST | Exchange Code for Token | Basic Auth |
| `/api/v1/auth/login` | POST | User Login | No |
| `/api/v1/auth/logout` | POST | User Logout | Yes |

### Key Invariants
1.  **Strict RFC Compliance**: `redirect_uri` matching must be exact.
2.  **No Admin Logic**: The Auth Plane must NEVER expose tenant management APIs.
3.  **Tenant Agnostic Login**: Users authenticate globally; tenant context is derived *after* login.

## Usage
```bash
./opentrusty serve auth
```
