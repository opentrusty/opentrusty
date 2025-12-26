# Audit Log Specification

OpenTrusty maintains an immutable audit trail for all security-sensitive operations.

## 1. Event Structure
All events MUST capture:
- **Actor**: The `user_id` or `service_id` performing the action.
- **Scope**: Whether the action is `platform` or `tenant` scoped.
- **Action**: The specific operation (e.g., `tenant:created`, `user:provisioned`, `client:secret_regenerated`).
- **Target**: The resource type and ID being affected.
- **Metadata**: Contextual information (IP address, user agent).
- **Timestamp**: High-precision UTC timestamp.

## 2. Mandatory Events

| Event Type | Plane | Trigger |
|------------|-------|---------|
| `auth:login_success` | Auth | Successful session establishment |
| `auth:login_failed` | Auth | Password mismatch or account lockout |
| `tenant:created` | Admin | New tenant provisioned by Platform Admin |
| `user:provisioned` | Admin | New user added to a tenant |
| `role:assigned` | Admin | Role assignment update |
| `client:created` | Admin | New OAuth2 client registration |
| `client:secret_rotated` | Admin | Client secret regeneration |

## 3. Storage & Integrity
- **Immutability**: Audit logs are "append-only" in the database.
- **Separation**: Audit logging is handled by a dedicated `audit.Logger` interface to facilitate future export to external SIEM systems (e.g., ELK, Splunk).
