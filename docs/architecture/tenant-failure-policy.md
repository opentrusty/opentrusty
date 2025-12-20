# Tenant Failure Policy

This policy defines the scoping behavior and error handling requirements for tenant-aware requests in OpenTrusty.

## 1. Scoping Definitions

### Tenant-Scoped Endpoints
All endpoints that interact with Identity, Authorization, or Protocols MUST be tenant-scoped. This includes:
- **Identity**: `POST /auth/register`, `POST /auth/login`, `POST /auth/change-password`.
- **OAuth2 / OIDC**: `/authorize`, `/token`, `/userinfo`, `/.well-known/openid-configuration`.
- **Management**: `/tenants/{id}/*`, `/users/*` (within a tenant).

### Tenant-Agnostic Endpoints
The following endpoints are considered internal platform utilities and MUST NOT require a tenant context:
- **System**: `/health`, `/version`.
- **Platform Admin**: Endpoints for creating *new* tenants (Platform level).

## 2. Failure Behavior

If a request hits a **Tenant-Scoped Endpoint** without a valid tenant identity, the system MUST behave as follows:

| Condition | Response Code | Logic |
| :--- | :--- | :--- |
| **Missing Identifier** | `400 Bad Request` | Neither `X-Tenant-ID` nor `tenant_id` present. |
| **Invalid Format** | `400 Bad Request` | Identifier is malformed. |
| **Unknown Tenant** | `404 Not Found` | Identifier is valid but no such tenant exists in DB. |
| **Access Denied** | `403 Forbidden` | Tenant exists but client/actor is not authorized for it. |

## 3. Policy Adherence

### Is "default tenant" allowed?
**NO.** Implicit defaulting to a "default" tenant is strictly forbidden. 
- A tenant with ID `default` MAY exist in the database for single-tenant or vanity deployments.
- However, the client MUST explicitly request it (e.g., `X-Tenant-ID: default`).
- Failure to provide any identifier MUST result in a `400 Bad Request`, never a fallback to a magic default.

### Can tenant ever be inferred implicitly?
**NO.** Every tenant-scoped request MUST carry an explicit identifier.
- Accepted sources: `X-Tenant-ID` header, `tenant_id` query parameter, or `Host` header (if subdomain mapping is enabled).
- There is no "current tenant" stored in the session or JWT that allows omitting this from the transport entry point; the transport layer must verify the identifier against the authenticated session during every request.

---
**Rule Citation**: Enforces **Rule 2.1 (Isolation)** and **Rule 2.3 (No Cross-Talk)**.
