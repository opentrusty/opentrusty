# Authority Model

This document defines the roles, scopes, and authority hierarchy within OpenTrusty.
**AI agents MUST read and comply with this document before modifying authorization code.**

## Scopes & Contexts

Authority is derived from the combination of **Role** and **Scope**.

| Scope | Context Required | Description |
| :--- | :--- | :--- |
| `platform` | `NULL` | Global authority over the entire installation. |
| `tenant` | `tenant_id` | Authority limited to a specific tenant. |
| `client` | `client_id` | Authority limited to a specific OAuth2 client/machine. |

---

## Defined Roles

### 1. Platform Admin (`platform_admin`)

-   **Scope**: `platform`
-   **Context**: None
-   **Capabilities**:
    -   Create and delete Tenants
    -   Manage system-wide configurations
    -   Assign Platform roles to other users
    -   **View ALL audit logs across ALL tenants (read-only)**
-   **Restrictions**:
    -   MUST NOT mutate tenant data by default
    -   MUST NOT manage tenant users directly (must provision via Tenant Owner)
    -   MUST NOT see secrets, credentials, or sensitive payloads
    -   All actions MUST be audited

> Platform Admin is an **operator**, not a tenant participant.
> Platform Admin ≠ Tenant Owner.

### 2. Tenant Owner (`tenant_owner`)

-   **Scope**: `tenant`
-   **Context**: `tenant_id`
-   **Capabilities**:
    -   Full authority within the tenant
    -   Manage tenant settings
    -   Manage tenant users (add/remove, assign roles)
    -   View tenant-scoped audit logs
    -   Register and manage OAuth2 clients
-   **Invariants**:
    -   First user created when a tenant is provisioned by Platform Admin
    -   Every tenant MUST have exactly one `tenant_owner`
    -   Cannot be removed by `tenant_admin`

### 3. Tenant Admin (`tenant_admin`)

-   **Scope**: `tenant`
-   **Context**: `tenant_id`
-   **Capabilities**:
    -   Operational administrator
    -   Manage users within their Tenant
    -   Register OAuth2 clients for their Tenant
    -   View audit logs for their Tenant
-   **Restrictions**:
    -   CANNOT delete the Tenant
    -   CANNOT remove or modify the Tenant Owner
    -   CANNOT transfer tenant ownership
    -   CANNOT see or modify other Tenants

### 4. Tenant Member (`tenant_member`)

-   **Scope**: `tenant`
-   **Context**: `tenant_id`
-   **Capabilities**:
    -   View basic Tenant information
    -   Access applications authorized for the Tenant
    -   Self-manage their own profile/credentials
-   **Restrictions**:
    -   CANNOT view audit logs
    -   CANNOT manage other users
    -   CANNOT register clients

---

## Audit Log Visibility Model

| Viewer | Audit Access | Notes |
| :--- | :--- | :--- |
| Platform Admin | ✅ All tenants (read-only) | `tenant_id` always visible; access is audited |
| Tenant Owner | ✅ Own tenant only | Full tenant audit visibility |
| Tenant Admin | ✅ Own tenant only | Operational audit visibility |
| Tenant Member | ❌ None | No audit access |

**Security Invariants:**
-   Platform Admin audit access is **intentional** for compliance/operations
-   Platform Admin audit access is **read-only** (no mutation)
-   Platform Admin audit access **does NOT break** tenant isolation at data mutation layer
-   All Platform Admin audit access MUST itself generate an audit entry
-   Sensitive fields (secrets, credentials) MUST be redacted uniformly

---

## Permission Logic

-   Permissions are additive.
-   A user may hold multiple roles across different scopes.
-   Authorization checks MUST use `HasPermission()`, not role name checks.
-   Platform scope assignments grant implicit access to all tenant-scoped read operations (wildcard `*`).
