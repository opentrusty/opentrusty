# Session Hardening Audit

## 1. Cookie Security
- **HttpOnly**: `true` (Default). Prevents client-side JS from accessing the session cookie.
- **Secure**: `false` (Default). **ACTION**: Should be set to `true` by default for non-localhost/HTTP environments.
- **SameSite**: `Lax` (Default). Provides CSRF protection for cross-site requests while allowing navigation.
- **Entropy**: Cryptographically secure 32-byte IDs via `crypto/rand`.

## 2. Session Lifecycle
- **Rotation**: **ACTION**: Must ensure old session is destroyed on successful login.
- **Lifetime**: 24h (Default).
- **Idle Timeout**: 30m (Default).
- **Cleanup**: Automatic cleanup of expired sessions is needed.

## 3. Plane Isolation
- **Status**: Currently uses a single cookie name `opentrusty_session`.
- **Recommendation**: Separate cookie names `ot_auth_session` and `ot_admin_session` to prevent credential leakage across planes, or implement context-aware validation.

## 4. JS Accessibility
- HttpOnly flag is enforced.
- API check: Verified that `SessionID` is never returned in JSON bodies.
