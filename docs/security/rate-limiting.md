# Rate Limiting Policy

OpenTrusty employs multi-layered rate limiting to protect against Brute Force, Denial of Service (DoS), and API abuse.

## 1. Global Protection
- **Middleware**: `RateLimitMiddleware` is applied to all API routes.
- **Default Limit**: 100 requests per minute per IP address.
- **Enforcement**: Returns `429 Too Many Requests` with a `Retry-After` suggestion.

## 2. Sensitive Endpoint Throttling
The following endpoints have stricter limits enforced via specialized middleware or handler-level checks:

| Endpoint | Limit | Purpose |
|----------|-------|---------|
| `/api/v1/auth/login` | 5 attempts / min | Prevent password brute-forcing |
| `/api/v1/oauth2/authorize` | 10 requests / min | Prevent authorization code spam |
| `/api/v1/oauth2/token` | 20 requests / min | Protect token exchange service |
| `/api/v1/user/change-password` | 3 attempts / hour | Prevent account takeover attempts |

## 3. Implementation Details
- **Mechanism**: Token Bucket algorithm (via `golang.org/x/time/rate`).
- **Identity**: Limits are tracked by source IP address for unauthenticated requests and `user_id` for authenticated requests.
- **Burst Capacity**: Allows for sub-second bursts to accommodate legitimate UI interactions (e.g., loading multiple assets).
