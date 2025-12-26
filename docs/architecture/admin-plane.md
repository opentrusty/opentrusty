# Admin Plane Architecture

## Purpose
The Admin Plane (`serve admin`) exposes the Management API. It allows Platform Admins to create tenants and Tenant Admins to manage their users. It effectively acts as the "Control Plane".

## Capabilities (Stage 4)

### Endpoints
| Path | Method | Purpose | Auth Required |
|------|--------|---------|---------------|
| `/health` | GET | Service Health | No |
| `/api/v1/auth/me` | GET | Session Check | Yes |
| `/api/v1/tenants` | GET | List Tenants | Platform Admin |
| `/api/v1/tenants` | POST | Create Tenant | Platform Admin |

### Key Invariants
1.  **Strict Authorization**: All endpoints (except health) require a valid Session Cookie AND appropriate RBAC permissions.
2.  **Audit Logging**: Every write operation (Create/Update/Delete) MUST be audited.
3.  **No Protocol Logic**: The Admin Plane does NOT issue tokens or handle OIDC flows.

## Usage
```bash
./opentrusty serve admin
```
