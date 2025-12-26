# Multi-Process Deployment

This document defines the production deployment model for running OpenTrusty as separate services.

## Architecture

In a production environment (Beta+), OpenTrusty SHOULD be deployed as two distinct services sharing a database:

```mermaid
graph TB
    subgraph "Public Internet"
        LB[Load Balancer / Reverse Proxy]
    end

    subgraph "Trust Zone"
        Auth[Auth Service<br/>(auth.opentrusty.org)]
        API[Management API Service<br/>(api.opentrusty.org)]
        DB[(PostgreSQL)]
    end

    LB -->|Host: auth.*| Auth
    LB -->|Host: api.*| API

    Auth --> DB
    API --> DB
```

## Service Definitions

### 1. Auth Service (`auth`)

- **Entrypoint**: `opentrusty serve auth`
- **Port**: 8080 (default)
- **Responsibility**: OIDC/OAuth2 protocols, User Login
- **Scaling**: CPU-bound (crypto operations: Argon2, RSA signing)
- **Exposure**: High traffic, public-facing

### 2. Management API Service (`api`)

- **Entrypoint**: `opentrusty serve admin`
- **Port**: 8081 (default)
- **Responsibility**: Tenant management, User CRUD, Audit logs
- **Scaling**: I/O-bound (database queries)
- **Exposure**: Low traffic, restricted to admins & control panel

## Shared Infrastructure

Both services MUST share:

1.  **Database**: Same PostgreSQL instance and credentials.
2.  **Secret Key**: Same `OPENTRUSTY_SECRET_KEY` for session encryption/decryption.
3.  **Clock**: NTP synchronization is critical for token validity.

## Configuration Profiles

### Auth Node
```env
OPENTRUSTY_MODE=auth
OPENTRUSTY_PORT=8080
OPENTRUSTY_DATABASE_URL=postgres://...
OPENTRUSTY_SECRET_KEY=...
```

### API Node
```env
OPENTRUSTY_MODE=admin
OPENTRUSTY_PORT=8081
OPENTRUSTY_DATABASE_URL=postgres://...
OPENTRUSTY_SECRET_KEY=...
```

## Security Benefits

1.  **Attack Surface Reduction**: The API service (which has power to delete tenants) is not exposed on the same port/process as the public login page.
2.  **Resource Isolation**: A heavy login spike (DDoS) won't starve the admin API, allowing operators to still manage the system.
3.  **Least Privilege**: Future versions can use different DB credentials for Auth (read-only on config tables) vs API (read-write).
