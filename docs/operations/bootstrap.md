# Platform Admin Bootstrap

This document describes the bootstrap mechanism for initializing the first Platform Admin in OpenTrusty.

## Rationale

To maintain a secure-by-default posture, OpenTrusty does not implicitly grant administrative privileges to the first registered user. Instead, an explicit, auditable bootstrap process is required.

## Bootstrap Conditions

1. **Clean Slate**: The bootstrap logic only executes if no user currently holds the `platform_admin` role at the `platform` scope.
2. **Explicit Intent**: The process must be triggered either via environment variables or a direct CLI command.

## Mechanism

### 1. Environment Variable (Automated)

The server will attempt to bootstrap the admin on every startup if the following environment variables are provided:

- `OT_BOOTSTRAP_ADMIN_EMAIL`: The email of the user to be elevated.
- `OT_BOOTSTRAP_ADMIN_TENANT_ID`: The tenant ID where the user resides.

Example:
```bash
export OT_BOOTSTRAP_ADMIN_EMAIL="admin@example.com"
export OT_BOOTSTRAP_ADMIN_TENANT_ID="sample"
./opentrusty
```

### 2. CLI Command (Manual/Dev)

You can explicitly trigger the bootstrap using the `bootstrap` command. This uses the same environment variables but provides immediate feedback and exits.

```bash
OT_BOOTSTRAP_ADMIN_EMAIL="admin@example.com" OT_BOOTSTRAP_ADMIN_TENANT_ID="sample" ./opentrusty bootstrap
```

## Security & Idempotency

- **Idempotent**: If a platform admin already exists, the bootstrap process will skip without making any changes.
- **Auditable**: Every successful bootstrap is recorded in the audit logs with the event type `platform_admin_bootstrapped`.
- **Scoped**: The privilege is granted only at the `platform` scope, ensuring clear separation between platform and tenant concerns.
