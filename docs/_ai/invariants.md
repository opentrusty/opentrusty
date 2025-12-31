# System Invariants

This document defines the non-negotiable security and architectural invariants of OpenTrusty.
Any code change that violates these invariants is **forbidden**.

## 1. Tenant Isolation Invariants

-   **MUST** enforce strict tenant isolation at the database level.
    -   Every query targeting tenant data **MUST** include a `tenant_id` WHERE clause.
-   **MUST NOT** use "magic" tenant IDs (e.g., "default", "system", "0000") to represent the platform.
-   **MUST NOT** allow a tenant-scoped session to access platform-scoped resources.
### I-102: Explicit Tenant Membership
- Every link between a user and a tenant MUST be recorded in the `tenant_members` table.
- Implicit membership via user-level flags is forbidden.
- The `identity.User` struct MUST NOT contain a `tenant_id` field.
-   **MUST** require an explicit `tenant_memberships` record for any user-tenant relationship.

## 2. Authorization Invariants

-   **MUST** express Platform authorization ONLY via scoped roles (Scope: `platform`).
-   **MUST** express Tenant authorization ONLY via scoped roles (Scope: `tenant`).
-   **MUST NOT** derive privileges from the presence or absence of a user record alone; privileges come from `rbac_assignments` and require explicit `tenant_memberships` for tenant-scoped actions.
-   **MUST** validate that a token's scope matches the requested resource's scope.
-   **MUST** strictly block Control Panel (Management Plane) login for users with only the `tenant_member` role.

## 3. Session & Token Invariants

-   **MUST** generate session IDs using cryptographically secure random number generators (CSPRNG) or UUIDv4.
-   **MUST** store sessions in the database; strictly NO stateless JWT sessions for core administration.
-   **MUST** verify the `aud` (Audience) and `iss` (Issuer) claims in all OIDC tokens.
-   **MUST** revoke all associated Refresh Tokens when a User session is terminated or an Access Token is revoked.

## 4. Secret Management

-   **MUST NOT** log secrets (passwords, tokens, keys) in plain text.
-   **MUST NOT** return hashed passwords in API responses.
-   **MUST** store client secrets as hashes, never in plain text.

## 5. Client Trust Invariants

-   **MUST** treat all HTTP clients (including Control Panel UI) as untrusted.
-   **MUST** enforce authorization server-side for every API request.
-   **MUST NOT** expose internal state or secrets to any client.
-   **MUST NOT** assume UI visibility equals authorization.
-   **MUST NOT** rely on client-side validation for security decisions.

## 6. Repository Scope Invariants

-   **AI MUST** recognize this repository represents ONLY:
    -   Authentication Plane (OIDC/OAuth2 endpoints, login pages)
    -   Management API Plane (REST admin APIs)
-   **AI MUST NOT** add Control Panel UI code, SPA assets, or frontend build systems to this repository.
-   **AI MUST NOT** add admin SPA routes or static file serving.
-   **AI MUST** treat any reference to UI as an external, untrusted client.
-   **AI MUST** ensure all UI interactions occur ONLY via Management API.

## 7. Protocol Surface Invariants

Login pages are **NOT** UI components. They are protocol surfaces.

-   **MUST** be server-rendered within the core binary.
-   **MUST** belong to the Authentication Plane.
-   **MUST NOT** be served from or delegated to the Control Panel UI.
-   **MAY** be customized via tenant branding configuration.
## 8. Tenant Ownership Invariants

-   **MUST** ensure every Tenant has exactly one `tenant_owner`.
-   **MUST** assign the `tenant_owner` role to the first user provisioned during tenant creation.
-   **MUST NOT** allow a Tenant to exist without an active `tenant_owner`.
-   **MUST NOT** allow a `tenant_admin` to delete a Tenant or remove the last `tenant_owner`.
-   **MUST** require Platform Admin to explicitly provision an owner when creating a Tenant.
## Audit Log Immutability

Audit logs are the authoritative record of security events. To ensure non-repudiation and system integrity:

1. **Append-Only**: Audit logs MUST be immutable once written.
2. **No Suppression**: No operations (API or CLI) shall exist to delete, modify, or suppress audit entries.
3. **Universality**: Every security-sensitive action MUST be recorded.
4. **Audit-of-Audit**: Every platform administrative access to tenant-scoped audit data MUST generate a primary audit record containing the actor, target, reason, and scope of access.
