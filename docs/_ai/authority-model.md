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

### 1. Identity vs. Membership (Strict Separation)

OpenTrusty maintains a hard boundary between **Identity** (who you are) and **Membership** (where you belong).

*   **Identities** (`identity.User`) are global. They contain credentials, profile data, and security status (lockout, etc.). They **do not** contain a `tenant_id`.
*   **Memberships** (`tenant.Membership`) link an Identity to a Tenant. A user can be a member of zero or more tenants.
*   **Authority** is derived from **Roles** assigned to an Identity within a specific **Scope** (Platform or Tenant).

**Invariant**: A user has authority in a tenant ONLY if they have both a `Membership` record AND a relevant `Role` assignment in that tenant.

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

| Viewer | Audit Access | Discovery Pattern |
| :--- | :--- | :--- |
| **Platform Admin** | ✅ Scoped (Read-only) | **Explicit Declaration Required** (Tenant, Window, Reason) |
| **Tenant Owner** | ✅ Own tenant only | Default access to own tenant |
| **Tenant Admin** | ✅ Own tenant only | Default access to own tenant |
| **Tenant Member** | ❌ None | No access |

---

## Control Panel Access Control

The Control Panel (Management Plane) is restricted to administrative identities.

| Role | Access | Authentication Flow |
| :--- | :--- | :--- |
| `platform_admin` | ✅ FULL | Direct Login (Session) |
| `tenant_owner` | ✅ TENANT ONLY | Direct Login (Session) |
| `tenant_admin` | ✅ TENANT ONLY | Direct Login (Session) |
| `tenant_member` | ❌ BLOCKED | OAuth2 / OIDC Flow ONLY |

**Security Guard**: `AuthMiddleware` and `Login` handler strictly reject `tenant_member` principal logins to the Control Panel with a `403 Forbidden`. End-users must authenticate via applications using protocol flows.

---

## Permission Logic

-   Permissions are additive.
-   A user may hold multiple roles across different scopes.
-   Authorization checks MUST use `HasPermission()`, not role name checks.
-   Platform scope assignments grant implicit access to all tenant-scoped read operations (wildcard `*`).
