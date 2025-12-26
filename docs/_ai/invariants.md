# System Invariants

This document defines the non-negotiable security and architectural invariants of OpenTrusty.
Any code change that violates these invariants is **forbidden**.

## 1. Tenant Isolation Invariants

-   **MUST** enforce strict tenant isolation at the database level.
    -   Every query targeting tenant data **MUST** include a `tenant_id` WHERE clause.
-   **MUST NOT** use "magic" tenant IDs (e.g., "default", "system", "0000") to represent the platform.
-   **MUST NOT** allow a tenant-scoped session to access platform-scoped resources.
-   **MUST** ensure that `tenant_id` is immutable once assigned to a resource.

## 2. Authorization Invariants

-   **MUST** express Platform authorization ONLY via scoped roles (Scope: `platform`).
-   **MUST** express Tenant authorization ONLY via scoped roles (Scope: `tenant`).
-   **MUST NOT** derive privileges from the presence or absence of a `tenant_id` in the users table alone; privileges come from `rbac_assignments`.
-   **MUST** validate that a token's scope matches the requested resource's scope.

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

