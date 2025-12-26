# OIDC Abuse Prevention

OpenTrusty enforces strict OIDC protocol compliance to prevent common authorization flow attacks.

## 1. Authorization Code Protection
- **One-Time Use**: Codes are marked as `is_used = true` immediately upon first exchange. Any subsequent request with the same code MUST be rejected.
- **Short Lifetime**: Codes expire after 5 minutes (RFC 6749 Section 4.1.2 recommendation is < 10min).
- **Binding**: Codes are strictly bound to the `client_id` and `redirect_uri` requested during the authorization step.

## 2. Redirect URI Security
- **Exact Match**: Only exact string matches against registered `redirect_uris` are allowed. No wildcard or partial matches.
- **Protocol Enforced**: Non-HTTPS redirect URIs are only allowed for `localhost`.

## 3. PKCE (Proof Key for Code Exchange)
- **Support**: SHA-256 (`S256`) and `plain` methods are supported.
- **Public Clients**: PKCE is highly recommended and will be enforced for all non-confidential clients in future stages.

## 4. State & Nonce
- **State**: Required for all authorization requests to prevent CSRF in the redirect flow.
- **Nonce**: Bound to the issued `id_token` to prevent replay attacks.

## 5. Information Leakage
- **Error Responses**: RFC-compliant error codes (e.g., `invalid_request`, `unauthorized_client`) are returned without leaking internal stack traces or database details.
