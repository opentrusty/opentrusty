# OIDC Capabilities (Stage 6)

This document outlines the OIDC features supported by OpenTrusty as of Stage 6. Features not listed here are intentionally unsupported to maintain a minimal, secure, and auditable footprint.

## Supported Features

| Feature | Status | Specification |
|---------|--------|---------------|
| Authorization Code Flow | **Supported** | RFC 6749, OIDC Core |
| PKCE (S256) | **Required** | RFC 7636 |
| ID Token | **Supported** | OIDC Core (RS256 Signed) |
| Standard Scopes | **Partial** | `openid`, `profile`, `email` |
| Client Authentication | **Supported** | `client_secret_basic`, `none` (for SPAs) |
| Redirect URI | **Required** | Exact matching enforced |

## Intentionally Unsupported (Stage 6)

The following features are disabled or not implemented to reduce attack surface and complexity for the first external integration:

- **Implicit Flow**: Disallowed for security reasons (use PKCE).
- **Resource Owner Password Credentials**: Disallowed (prevents credential scraping).
- **Dynamic Client Registration**: Not implemented (prevents client spam).
- **Refresh Tokens**: Not explicitly issued for this Stage verification.
- **Custom Claims**: Limited to standard identity claims.
- **UserInfo Endpoint**: Aggregated facts are currently provided in the ID Token.

## Security Invariants
1. **PKCE is mandatory** for all authorization code exchanges.
2. **State parameter** is required to prevent CSRF in the redirect flow.
3. **Redirect URIs** must be pre-registered and matched exactly.
