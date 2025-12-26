# Tenant Context Resolution

**Status:** Normative  
**Version:** 1.0  
**Last Updated:** 2025-12-26

## Purpose

This document eliminates tenant context ambiguity in the OpenTrusty identity control plane by defining the single, authoritative source of tenant resolution for every request type.

## Critical Context

**OpenTrusty is an administrative control plane, not an end-user login portal.**

Only these roles log into the OpenTrusty UI:
- `platform_admin` (tenant-agnostic)
- `tenant_owner` (tenant-scoped)
- `tenant_admin` (tenant-scoped)

**Tenant members (`member` role) never log into OpenTrusty UI.** End users authenticate via external applications using OAuth2/OIDC flows, not via OpenTrusty's admin interface.

## Tenant Context Rules

### Table: Tenant Resolution by Context

| Request Context         | Tenant Source         | Status |
|------------------------|-----------------------|--------|
| Unauthenticated request (`/auth/login`) | User record lookup | ✅ Required |
| Unauthenticated request | `X-Tenant-ID` header | ❌ Forbidden |
| Unauthenticated request | Session               | ❌ Not possible |
| Authenticated request   | Session (`tenant_id`) | ✅ Required |
| Authenticated request   | `X-Tenant-ID` header | ❌ Rejected (400) |
| OAuth2 flow             | `client_id → tenant_id` | ✅ Required |

### Rule A: Unauthenticated Endpoints

Endpoints that do NOT require session authentication:
-   `/health`
-   `/.well-known/openid-configuration`
-   `/jwks.json`
-   `/auth/login` (tenant derived from user record)
-   `/auth/register` (should be disabled for production; see Control Plane Login Model)

**MUST NOT** accept tenant context from `X-Tenant-ID` header or any other client-supplied source.

**For `/auth/login` specifically:**
- Tenant is derived by looking up the user record: `SELECT tenant_id FROM users WHERE email = ?`
- The user's `tenant_id` field is immutable and set at account creation
- No tenant selection is possible or necessary

**Rationale:** Tenant context without authenticated identity is a spoofing vector. For login, the user's tenant affiliation is a database fact, not a client assertion.

### Rule B: Authenticated Endpoints

Endpoints protected by `AuthMiddleware`:
-   `/api/v1/auth/me`
-   `/api/v1/user/profile`
-   `/api/v1/tenants/*`
-   All other authenticated routes

**MUST** derive tenant context **exclusively from the session**.

The session model contains:
```go
type Session struct {
    TenantID *string  // Nullable: NULL for platform admins
    UserID   string
    // ...
}
```

`AuthMiddleware` (located in `internal/transport/http/middleware.go`) is responsible for:
1. Validating the session
2. Extracting `session.TenantID`
3. Injecting it into `context.Context` via `tenantIDKey`

**If a client sends `X-Tenant-ID` header on an authenticated request, the server MUST:**
-   Reject with `400 Bad Request`
-   Log a security warning (no PII)

**Rationale:** Accepting tenant context from headers post-authentication creates a spoofing vector and undermines session-based isolation.

### Rule C: OAuth2/OIDC Flows

OAuth2 and OIDC endpoints (under `/oauth2/` prefix):
-   `/oauth2/authorize`
-   `/oauth2/token`
-   `/oauth2/revoke`

**Tenant resolution hierarchy:**
1. **Client-to-Tenant mapping:** `client_id` (from API request) → `tenant_id` (via `oauth2_clients` table lookup)
2. **Authorization Code:** When redeeming a code, tenant is carried from the code's associated client

**Headers are never consulted** for tenant resolution in OAuth2 flows.

**Rationale:** OAuth2 security model binds tenant context to the client credential, not to request headers.

### Rule D: Control Plane Login (`/auth/login`)

**Status:** Normative

`/auth/login` is a control-plane authentication endpoint for administrative operators.

**Client Requirements:**
- Clients **MUST NOT** provide tenant context via:
  - `X-Tenant-ID` header
  - Subdomain
  - Query parameter
  - Request body field

**Server Behavior:**
- Tenant context **MUST** be derived server-side **after** successful authentication
- Authentication is based **solely** on email + password
- Tenant context resolution:
  - **Platform admins**: `session.tenant_id = NULL` (tenant-agnostic privileges via RBAC)
  - **Tenant admins/owners**: `session.tenant_id = <user.tenant_id>` (immutable assignment from user record)

**Rejection Rule:**
- If client supplies tenant context during login → **400 Bad Request**

**Critical Distinction:**
> **OpenTrusty UI users are operators, not tenant members.**  
> Business users (tenant members) never log into OpenTrusty UI; they authenticate via OAuth2/OIDC flows to external applications.

**Forbidden Pattern:**
```typescript
// ❌ FORBIDDEN
POST /auth/login
X-Tenant-ID: tenant-abc
{ "email": "...", "password": "..." }
```

**Correct Pattern:**
```typescript
// ✅ CORRECT
POST /auth/login
{ "email": "admin@example.com", "password": "..." }

// Backend derives tenant_id from user record
// Session created with appropriate tenant_id or NULL
```

## Middleware Responsibilities

### `TenantMiddleware`

**Location:** `internal/transport/http/middleware.go:L70-96`

**Behavior:**
-   Attempts to extract tenant from `X-Tenant-ID` header or `tenant_id` query parameter
-   Stores in context if found
-   **Does NOT enforce** tenant presence (handler-specific)

**Usage:** Applied to routes that MAY need tenant context for routing (e.g., `/oauth2/*`)

### `RequireTenant`

**Location:** `internal/transport/http/middleware.go:L99-108`

**Behavior:**
-   Enforces that tenant context exists in `context.Context`
-   Returns `400 Bad Request` if missing

**Usage:** Applied to tenant-scoped unauthenticated endpoints (currently none in routing definition except login/register, which will be revised)

### `AuthMiddleware`

**Location:** `internal/transport/http/middleware.go:L111-164`

**Behavior:**
1. Validates session cookie
2. Loads session from database
3. Validates tenant isolation: if request contains `X-Tenant-ID` header, cross-checks against session tenant (L149-156)
4. **Injects session tenant into context** (L160)

**Post-hardening behavior:**
-   If `X-Tenant-ID` header is present when session exists → **reject with 400**

## Forbidden Patterns

The following patterns are **explicitly forbidden**:

❌ **Accepting `X-Tenant-ID` on authenticated routes**
```go
// FORBIDDEN
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
    tenantID := r.Header.Get("X-Tenant-ID")  // ❌ NO
    // ...
}
```

✅ **Correct pattern:**
```go
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
    tenantID := GetTenantID(r.Context())  // ✅ From session via AuthMiddleware
    // ...
}
```

❌ **Requiring tenant header when session exists**
```typescript
// FORBIDDEN (Frontend)
await client.GET("/auth/me", {
    headers: { "X-Tenant-ID": "..." }  // ❌ Session already has this
});
```

✅ **Correct pattern:**
```typescript
await client.GET("/auth/me", {});  // ✅ Tenant from session cookie
```

## Control Plane Login Model

**OpenTrusty UI login NEVER requires tenant input from the user.**

### Account-Level Tenant Binding

Admin accounts in OpenTrusty have **immutable tenant assignment**:

1. **Platform Admins** (`platform_admin` role):
   - `tenant_id = NULL` in `users` table
   - Can operate across all tenants via RBAC assignments
   - Managed via bootstrap process or by other platform admins

2. **Tenant Admins** (`tenant_owner`, `tenant_admin` roles):
   - `tenant_id = <specific UUID>` in `users` table
   - Bound to **exactly one tenant** at account creation
   - Cannot access other tenants under any circumstance

3. **Tenant Members** (`member` role):
   - **DO NOT log into OpenTrusty UI**
   - Authenticate via external applications using OAuth2/OIDC

### Login Flow (Email/Password Only)

```
User submits → POST /auth/login
{
  "email": "admin@example.com",
  "password": "..."
}
         ↓
Backend queries:
  SELECT * FROM users WHERE email = ?
  (Global lookup or with tenant index)
         ↓
If found:
  - Create session with user.tenant_id
  - Return session cookie
         ↓
All subsequent requests:
  - Tenant context = session.tenant_id
  - No headers consulted
```

**No tenant selection, no tenant dropdown, no X-Tenant-ID header.**

### Why Multi-Tenant User Selection Is Forbidden

#### Security Rationale

1. **Spoofing Risk**: Accepting tenant context from client headers creates an attack surface where:
   - Malicious clients can attempt to impersonate other tenants
   - Session fixation attacks become easier
   - Audit trails become ambiguous (which tenant context was intended?)

2. **Privilege Escalation**: If users could "choose" their tenant at login:
   - A compromised account could pivot to unintended tenants
   - RBAC checks would need to validate both user identity AND tenant selection
   - Cross-tenant data leaks become more likely

#### Control Plane Design Rationale

3. **Admin Identity Is Tenant-Scoped**: 
   - Admin accounts are **organizational identities**, not portable user credentials
   - A "Tenant Admin for Acme Corp" is fundamentally different from "Tenant Admin for Beta Inc"
   - Allowing the same email to select different tenants violates the separation of duties principle

4. **Operational Clarity**:
   - In a control plane, admins represent **specific organizations**
   - Tenant affiliation is an **employment relationship**, not a preference
   - Changing tenant context requires provisioning a new admin account in the target tenant

5. **Contrast with End-User Systems**:
   - **End-user SaaS** (e.g., project management tool): Users may belong to multiple workspaces
   - **Admin control plane**: Admin accounts are organizational assets, tenant-bound by design

### Anonymous Registration Status

**The `/auth/register` endpoint is currently UNGUARDED and allows anonymous account creation.**

> **⚠️ SECURITY RISK**: This endpoint MUST be disabled or restricted to bootstrap-only usage before production deployment.

**Recommended Actions:**

1. **Option A: Disable Entirely**
   ```go
   func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
       respondError(w, http.StatusForbidden, "anonymous registration is disabled; admins must be provisioned by platform admins")
       return
   }
   ```

2. **Option B: Bootstrap Guard**
   ```go
   // Only allow if:
   // - No platform admins exist (first-run bootstrap)
   // - OR request has valid bootstrap token
   ```

3. **Option C: Admin-Only Provisioning**
   - Remove `/auth/register` from public routes
   - Replace with `/admin/tenants/{id}/users` (authenticated, authorized)

**Current State**: Option A recommended until proper admin provisioning UI is built.

## Security Properties

This model enforces:

1. **Session-Based Isolation**: Tenant context cannot be manipulated post-authentication or during authentication
2. **Defense in Depth**: Even if a client sends spoofed headers, middleware rejects them
3. **Fail-Closed**: Missing tenant context on authenticated routes fails the request
4. **Audit Trail**: Tenant context is immutably tied to the session, logged in audit events
5. **Account Integrity**: Admin accounts cannot pivot between tenants (prevents lateral movement)

## References

-   **Middleware Implementation:** `internal/transport/http/middleware.go`
-   **Session Model:** `internal/session/session.go`
-   **Authority Model:** `docs/_ai/authority-model.md`
-   **Invariants:** `docs/_ai/invariants.md`
