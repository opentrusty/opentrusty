# Admin Authorization Matrix

This document defines the access boundaries for different administrative roles within the OpenTrusty platform.

## Roles & Scopes

| Role | Scope | Description |
|------|-------|-------------|
| **Platform Admin** | `ScopePlatform` | Root administrative access across the entire platform. |
| **Tenant Owner** | `ScopeTenant` | Full administrative control over a specific tenant. |
| **Tenant Admin** | `ScopeTenant` | Day-to-day administrative access to a specific tenant. |

## Permission Mapping

| Permission | Platform Admin | Tenant Owner | Tenant Admin |
|------------|----------------|--------------|--------------|
| `platform:manage_tenants` | **Yes** | No | No |
| `platform:view_all` | **Yes** | No | No |
| `tenant:manage_users` | **Yes** | **Yes** | **Yes** |
| `tenant:manage_clients` | **Yes** | **Yes** | **Yes** |
| `tenant:view` | **Yes** | **Yes** | **Yes** |
| `tenant:delete` | **Yes** | **Yes** | No |

## Invariants
1. **Cross-Tenant Isolation**: A Tenant Admin assigned to Tenant A can NEVER access or modify resources in Tenant B. This is enforced via `AuthMiddleware` derive tenant context and `HasPermission` scope context validation.
2. **Platform vs Tenant Boundary**: Platform permissions (e.g., creating a new tenant) are ONLY available to users with a `ScopePlatform` assignment.
3. **Session context**: Authority is derived EXCLUSIVELY from the session record in the database, not from request headers.
