# Storage Model

This document defines OpenTrusty's storage model and its guarantees regarding state persistence.

## 1. External PostgreSQL Dependency

OpenTrusty treats PostgreSQL as an **external infrastructure dependency**. It is a stateful service that exists outside the application boundary.

- **Primary Source of Truth**: PostgreSQL is the ONLY persistent source of truth for all OpenTrusty data (Identities, Tenants, OAuth2 Sessions, Audit Logs).
- **Decoupled Lifecycle**: The database lifecycle (provisioning, backup, scaling) is independent of the OpenTrusty application lifecycle.
- **Dev vs. Production**: 
    - **Development/Testing**: A Docker Compose-managed PostgreSQL instance is provided for convenience.
    - **Production**: PostgreSQL should be treated as a managed service or an appropriately hardened standalone cluster.

## 2. Compatibility & Portability

OpenTrusty aims for maximum portability across different PostgreSQL environments.

- **No Vendor Extensions**: The schema relies on standard PostgreSQL features. We strictly avoid proprietary extensions to ensure compatibility with all standard distributions.
- **Managed Services**: OpenTrusty is fully compatible with managed PostgreSQL services, including:
    - Amazon RDS / Aurora
    - Google Cloud SQL
    - Azure Database for PostgreSQL
    - Supabase (standard Postgres layer)
    - DigitalOcean Managed Databases

## 3. Responsibility Boundaries

The operator is responsible for the following database operations:

- **High Availability (HA)**: Configuring replication, failover, and connection pooling (if required at the infrastructure level).
- **Backups**: Implementing and verifying regular point-in-time recovery (PITR) or snapshots.
- **Upgrades**: Performing major version upgrades of the PostgreSQL engine.
- **Security**: Configuring TLS for the connection and enforcing network-level access control.

OpenTrusty is responsible for:
- **Schema Management**: Handling migrations and ensuring the schema is correct for the running version of the application.
- **Query Efficiency**: Ensuring indices and queries are optimized for standard Postgres execution plans.

## 4. Long-Term Compatibility Guarantees

OpenTrusty commits to the following for its storage layer:

1. **Standard Postgres Compliance**: We will always target a reasonable range of stable PostgreSQL versions (currently 15+).
2. **Migration Stability**: Schema migrations are atomic and designed to be non-destructive during upgrades.
3. **Transparent Persistence**: All table definitions and indices are documented in source code and are accessible for standard auditing tools.
