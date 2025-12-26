# Capabilities

> [!IMPORTANT]
> This document defines the **supported** functional contract of the OpenTrusty Identity Provider. 
> Features not listed here are explicitly **NOT supported** or **Out of Scope** for the current version.

## OAuth2 Support
OpenTrusty implements a strict subset of OAuth 2.0 (RFC 6749) optimized for security.

### Supported Grant Types
| Grant Type | Compliance | Notes |
| :--- | :--- | :--- |
| `authorization_code` | RFC 6749 §4.1 | The **ONLY** supported flow for user authentication. |
| `refresh_token` | RFC 6749 §6 | Supported for offline access. |

### Not Supported Grant Types
- `implicit` (Insecure, deprecated by OAuth 2.1)
- `password` (Resource Owner Password Credentials - Insecure)
- `client_credentials` (Machine-to-Machine not currently exposed)

### Security Extensions
- **PKCE** (RFC 7636): Enforced for public clients.
- **State Parameter**: Required for CSRF protection.
- **Client Authentication**: `client_secret_post` (Form POST) and Basic Auth.

## OpenID Connect (OIDC) Support
OpenTrusty acts as an OpenID Provider (OP) compliant with OIDC Core 1.0.

### Core Features
- **Flow**: Code Flow (`response_type=code`).
- **Signing Algorithm**: RS256 (RSA Signature with SHA-256).
- **Discovery**: `/.well-known/openid-configuration` (OIDC Discovery).
- **Keys**: `/jwks.json` (RFC 7517).

### Claims
The `id_token` includes the following standard claims:
- `iss`: Issuer Identifier
- `sub`: Subject Identifier (Stable, scoped to Tenant)
- `aud`: Audience (Client ID)
- `exp`: Expiration Time
- `iat`: Issued At
- `nonce`: String value used to associate a Client session with an ID Token (Replay protection)
- `at_hash`: Access Token Hash

### Not Supported
- `response_type=token` or `id_token` (Implicit Flow)
- Encryption (JWE)
- Dynamic Client Registration

## Multi-Tenancy
Multi-tenancy is a **core domain invariant**.

- **Isolation**: Strictly enforced at the Database and API layer.
- **Cross-Tenant Access**: Impossible by design. A session from Tenant A cannot access resources in Tenant B.
- **Resolution**: Tenant ID is resolved via `X-Tenant-ID` header or URL query parameter.

> **Note**: OpenTrusty UI login is tenant-agnostic at request time; tenant context is established post-authentication. See `architecture/tenant-context-resolution.md` for control plane login specifics.

## Security Model
### Authentication
- **Session Management**: Server-side sessions backed by `HttpOnly`, `Secure`, `SameSite=Lax` cookies.
- **NO JWT Sessions**: Browser sessions do NOT use JWTs.
- **Password Storage**: Argon2id (RFC 9106).
- **Account Lockout**: Automatic lockout after configurable failed attempts.

### Audit
- **Method**: Structured Logging (`slog`).
- **Events**: Login Success/Failure, Token Issuance, Password Change.
- **Privacy**: No PII or secrets in logs.

## Non-Goals
These features are intentionally **Out of Scope** for OpenTrusty:
- **User Self-Service UI**: OpenTrusty provides the Engines/APIs; the UI is the integrator's responsibility.
- **Dynamic Client Registration**: Clients are managed via administrative APIs/SQL.
- **Social Login**: No federation with Google/GitHub/Facebook.
- **Fine-grained Policy**: Complex authorization (Rego/OPA) is delegated to downstream apps.
