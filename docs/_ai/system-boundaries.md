# System Boundaries

This document defines what this repository owns and explicitly does NOT own.

## Repository Identity

**Repository**: `opentrusty`  
**Purpose**: Core Auth Engine + Management API  
**Domains**: `auth.opentrusty.org` (Port 8080), `api.opentrusty.org` (Port 8081)
**Enforcement**: Strict router segregation via `opentrusty serve [auth|admin]`

## What This Repository Owns

### Authentication Plane (`auth.*`)
- OIDC/OAuth2 protocol endpoints
- Server-rendered login, consent, and error pages
- Session cookie issuance and validation
- Token generation and signing
- **Constraint**: MUST NOT expose Management APIs (404 enforced)

### Management API Plane (`api.*`)
- Tenant lifecycle (create, read, update, delete)
- User provisioning and management
- OAuth client registration
- RBAC role and assignment management
- Audit log access
- **Constraint**: MUST NOT expose Login/OIDC endpoints (404 enforced)

### Shared Domain Core
- Identity service (user management)
- Session service (state management)
- Authorization service (RBAC enforcement)
- Tenant service (isolation logic)
- Database repositories

## What This Repository Does NOT Own

| Component | Owner | Interaction |
|-----------|-------|-------------|
| Control Panel UI | `opentrusty-control-panel` | Consumes Management API |
| Static SPA assets | `opentrusty-control-panel` | None |
| Frontend routing | `opentrusty-control-panel` | None |
| React/Vue/Tailwind | `opentrusty-control-panel` | None |

## Dependencies

### This Repo Depends On
- PostgreSQL (persistence)
- OpenTelemetry collector (observability, optional)

### Other Repos Depend On This Repo
- `opentrusty-control-panel` depends on Management API (`api.*`)

## Forbidden Cross-Overs

| Action | Status |
|--------|--------|
| Embedding SPA code in binary | ❌ FORBIDDEN |
| Serving static UI assets | ❌ FORBIDDEN |
| Importing frontend frameworks | ❌ FORBIDDEN |
| Storing frontend secrets | ❌ FORBIDDEN |
| Implementing UI routing | ❌ FORBIDDEN |
