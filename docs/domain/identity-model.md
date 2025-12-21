# Identity & Authorization Model

OpenTrusty follows a strict separation between **Identity** (who you are) and **Authorization** (what you can do). This document outlines the core concepts and design patterns used to enforce security and isolation.

## Core Concepts

### 1. Identity
An **Identity** represents a unique user in the system. 
- In OpenTrusty, all identities are stored in the `users` table.
- **Critical Principle**: OpenTrusty does NOT distinguish between "platform users" and "tenant users" at the identity layer. A user is simply an entry with an email and credentials.
- **Platform Separation Rule**: No tenant represents the platform. All users have a `tenant_id`.

### 2. Role
A **Role** is a logical grouping of permissions.
- Roles are assigned to identities within a specific context (e.g., Platform, Tenant, or Project).
- Roles are **not** global; they express a relationship between an identity and a tier/resource.
- **Platform Separation Rule**: Platform authorization is expressed only via scoped roles.

### 3. Permission
A **Permission** is the granular capability to perform an action (e.g., `tenant:create`, `user:provision`, `oauth2:client_register`).
- Permissions are associated with Roles.
- Business logic checks for permissions, never for roles directly.
- **Platform Separation Rule**: Tenant context must never be elevated to platform context.

---

## Platform Admin Representation

A **Platform Admin** is a regular user (with a tenant) who has been granted a platform-scoped role.

### How It Works:
1.  User belongs to a tenant (e.g., `tenant_id = 'sample'`).
2.  User is granted the `platform_admin` role via `rbac_assignments`:
    - `scope = 'platform'`
    - `scope_context_id = NULL` (the only valid value for platform scope)
3.  The system checks `rbac_assignments` to determine if the user has platform-level privileges.

### What This Means:
- There is **no separate `platform_admin` user table**.
- There is **no magic `tenant_id`** that confers platform privileges.
- Multiple users can be Platform Admins.

---

## Administrative Patterns

Administrative power in OpenTrusty is expressed solely via **Scoped Authorization**. There are no "admin accounts"—only users with administrative roles in specific tenants.

### Platform Administrator
A Platform Admin is simply a user who has an administrative role within the `default` (system) tenant.
- **Context**: `tenant_id = 'default'`
- **Capabilities**: Manage other tenants, system-wide settings, and global resources.

### Tenant Administrator
A Tenant Admin is a user who has an administrative role within a specific custom tenant.
- **Context**: `tenant_id = 'my-org-uuid'`
- **Capabilities**: Manage users, OIDC clients, and branding for that specific tenant.

---

## Scope-Based Authorization (OAuth2/OIDC)

For external applications interacting with OpenTrusty, authorization is mediated via **Scopes**.
- **OpenID Scopes**: `openid`, `profile`, `email` (standard OIDC identity claims).
- **Resource Scopes**: Custom scopes that define access to specific APIs (e.g., `api:read`, `api:admin`).
- Scopes are requested by the client and granted by the user during the consent flow.

---

## Scoped RBAC Implementation

Administrative power is formalized through a set of scoped role and permission tables. This allows for clear separation of concerns across different administrative tiers.

### Tables
1. **rbac_permissions**: Defines the granular capabilities (e.g., `tenant:create`).
2. **rbac_roles**: Groupings of permissions scoped to a specific context.
   - `platform`: Global capabilities (assigned in `default` tenant).
   - `tenant`: Capabilities within a specific tenant.
   - `client`: Capabilities within a specific OAuth2 client/resource context.
3. **rbac_role_permissions**: Maps permissions to roles.

### Patterns
- **Platform Admin**: Assigned the `platform_admin` role within the `default` tenant.
- **Tenant Admin**: Assigned the `tenant_admin` role within their respective tenant.
- **Client Manager**: Assigned a `client` scoped role for a specific resource.

## Anti-Pattern Warnings

To maintain a clean and secure architecture, the following patterns are **strictly forbidden** in OpenTrusty:

> [!CAUTION]
> ### No `is_admin` Flags
> There are no boolean flags like `is_admin` or `is_platform_user` in the database schema. Administrative power must always be derived from a relationship (Role/Permission) found in the `tenant_user_roles` or `user_project_roles` tables.

> [!WARNING]
> ### No Hardcoded Roles
> Business logic should never check for a specific role string (e.g., `if user.Role == "admin"`). Instead, check for the required permission (e.g., `if authz.HasPermission(user, "tenant:delete")`). This allows for flexible role definitions without changing code.

> [!IMPORTANT]
> ### Fail-Closed Resolution
> If a tenant or user context cannot be definitively resolved, the system must fail-closed (return `401 Unauthorized` or `403 Forbidden`) rather than defaulting to a "guest" or "system" context.
