# Architecture Map

This document defines the current boundaries of the system.

## Domain Boundaries (`internal/`)

The application is structured by domain. Cross-domain dependencies should be minimized and explicit.

| Directory | Domain Responsibility | Dependencies (Allowed) |
| :--- | :--- | :--- |
| `internal/audit` | Audit logging (Who did what) | `store`, `observability` |
| `internal/authz` | Authorization Enforcement (RBAC) | `store`, `identity` |
| `internal/config` | Configuration loading | *None* |
| `internal/identity` | User management, Credentials | `store`, `tenant` |
| `internal/oauth2` | OAuth2 Protocol Logic | `store`, `identity`, `tenant` |
| `internal/observability` | Tracing, Metrics, Logging | *None* |
| `internal/oidc` | OIDC Protocol Logic | `store`, `oauth2` |
| `internal/session` | Session Management | `store`, `identity` |
| `internal/store` | Data Access Layer (PostgreSQL) | *None* (Leaf) |
| `internal/tenant` | Tenant Lifecycle | `store` |
| `internal/transport` | HTTP/GRPC Handlers (Edge) | **ALL Domains** |

## Layering

1.  **Transport Layer** (`internal/transport`): Handles HTTP/GRPC requests. Decodes input, calls Domain Layer, encodes output. **NO business logic.**
2.  **Domain Layer** (`internal/identity`, `internal/oauth2`, etc.): Core business logic. **NO direct HTTP dependencies.**
3.  **Storage Layer** (`internal/store`): Database interactions. **NO business logic.**

## Protocol Logic

-   **OAuth2** logic resides strictly in `internal/oauth2`.
-   **OIDC** logic resides strictly in `internal/oidc`.
-   **Session** handling resides in `internal/session`.

## External Consumers

The following systems consume this core's APIs but are **NOT part of this repository**:

| Consumer | Repository | Relationship |
| :--- | :--- | :--- |
| Control Panel UI | `opentrusty-control-panel` | **Untrusted API client** |

**Critical Rule**: The core binary contains **NO UI, NO templates, NO frontend assets**.
All administrative UIs are external consumers that interact via documented HTTP APIs.

