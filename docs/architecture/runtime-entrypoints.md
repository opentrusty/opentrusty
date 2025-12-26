# Runtime Entrypoints

This document defines the runtime entrypoint architecture for OpenTrusty deployment.

## Current State (Alpha)

The current binary serves all planes on a single HTTP port:

```
opentrusty server    # Serves auth + admin API on PORT
opentrusty migrate   # Runs database migrations
opentrusty bootstrap # Creates initial platform admin
```

## Target State (Beta+)

The binary MUST support mode-based entrypoints for production deployment:

```bash
opentrusty serve auth   # Authentication Plane only (auth.*)
opentrusty serve admin  # Management API only (api.*)
opentrusty serve all    # Both planes (development only)
```

## Entrypoint Responsibilities

| Mode | Domain | Endpoints | Use Case |
|------|--------|-----------|----------|
| `auth` | `auth.*` | OIDC/OAuth2, login pages, session cookies | End-user authentication |
| `admin` | `api.*` | REST admin APIs, tenant/user/client management | Control Panel consumption |
| `all` | localhost | Both auth + admin | Local development only |

## Security Implications

### Separate Processes

In production, `auth` and `admin` SHOULD run as separate processes:

```
┌─────────────────┐     ┌─────────────────┐
│  opentrusty     │     │  opentrusty     │
│  serve auth     │     │  serve admin    │
│  (port 8080)    │     │  (port 8081)    │
└────────┬────────┘     └────────┬────────┘
         │                       │
         └───────────┬───────────┘
                     │
              ┌──────┴──────┐
              │  PostgreSQL  │
              └─────────────┘
```

### Runtime Resource Model

In production, `auth` and `admin` run as **independent processes** and DO NOT share any in-memory resources.

#### Database
- Each entrypoint maintains its own database connection pool.
- Both connect to the same PostgreSQL cluster.
- All cross-service consistency is enforced at the database schema and transaction level.

#### Audit Logging
- Both entrypoints emit audit events using a shared audit schema.
- Audit records are written to a common backend (e.g., PostgreSQL audit tables).
- No audit writer or buffer is shared in memory.

#### Sessions
- Browser sessions are persisted in a shared session store (database-backed).
- Session cookies are validated by querying persistent session data.
- No session state is cached or shared in memory across processes.

This design ensures:
- Strong process isolation
- Clear failure boundaries
- Horizontal scalability
- Elimination of cross-process trust assumptions

### Isolated Concerns

| Concern | Auth Mode | Admin Mode |
|---------|-----------|------------|
| Credential handling | ✅ | ❌ |
| Token issuance | ✅ | ❌ |
| Login pages | ✅ | ❌ |
| Tenant CRUD | ❌ | ✅ |
| User provisioning | ❌ | ✅ |
| Client registration | ❌ | ✅ |

## Configuration

Environment variables for entrypoint selection:

```bash
# Auth mode
OPENTRUSTY_MODE=auth
OPENTRUSTY_AUTH_PORT=8080
OPENTRUSTY_AUTH_DOMAIN=auth.opentrusty.org

# Admin mode
OPENTRUSTY_MODE=admin
OPENTRUSTY_ADMIN_PORT=8081
OPENTRUSTY_ADMIN_DOMAIN=api.opentrusty.org
```

## Implementation Status

| Feature | Status | Target Release |
|---------|--------|----------------|
| Single binary serve | ✅ Implemented | Alpha |
| `migrate` subcommand | ✅ Implemented | Alpha |
| `bootstrap` subcommand | ✅ Implemented | Alpha |
| `serve auth` mode | ⏳ Planned | Beta |
| `serve admin` mode | ⏳ Planned | Beta |
| Host-based routing | ⏳ Planned | Beta |

## Future: Domain Routing

When multi-process mode is implemented:

```nginx
# Example reverse proxy configuration
server {
    server_name auth.opentrusty.org;
    location / {
        proxy_pass http://opentrusty-auth:8080;
    }
}

server {
    server_name api.opentrusty.org;
    location / {
        proxy_pass http://opentrusty-admin:8081;
    }
}
```
