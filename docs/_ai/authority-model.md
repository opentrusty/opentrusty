# Authority Model

This document defines the roles, scopes, and authority hierarchy within OpenTrusty.

## Scopes & Contexts

Authority is derived from the combination of **Role** and **Scope**.

| Scope | Context Required | Description |
| :--- | :--- | :--- |
| `platform` | `NULL` | Global authority over the entire installation. |
| `tenant` | `tenant_id` | Authority limited to a specific tenant. |
| `client` | `client_id` | Authority limited to a specific OAuth2 client/machine. |

## Defined Roles

### 1. Platform Admin (`platform_admin`)
-   **Scope**: `platform`
-   **Context**: None
-   **Capabilities**:
    -   Create and delete Tenants.
    -   Manage system-wide configurations.
    -   Assign Platform roles to other users.
    -   **CANNOT** access Tenant data unless explicitly granted a Tenant role (Separation of Concern).

### 2. Tenant Admin (`tenant_admin`)
-   **Scope**: `tenant`
-   **Context**: `tenant_id`
-   **Capabilities**:
    -   Manage users within their Tenant.
    -   Register OAuth2 clients for their Tenant.
    -   Configure Tenant-specific settings.
    -   **CANNOT** see or modify other Tenants.

### 3. Tenant Member (`member`)
-   **Scope**: `tenant`
-   **Context**: `tenant_id`
-   **Capabilities**:
    -   View basic Tenant information.
    -   Access applications authorized for the Tenant.
    -   Self-manage their own profile/credentials.

## Permission Logic
-   Permissions are additive.
-   A user may hold multiple roles across different scopes (e.g., `member` of Tenant A and `tenant_admin` of Tenant B).
