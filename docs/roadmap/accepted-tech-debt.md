# Accepted Tech Debt

This document lists architectural issues that are **ACCEPTED** as technical debt. These issues will NOT be fixed in Phase 1 or 2.

## 1. Identity ↔ Password Coupling

- **Status**: ACCEPTED TECH DEBT
- **Coupling**: The `CreateUser` service method requires a password argument, making it impossible to create users without password credentials.
- **Why we are NOT fixing it now**: 
    - Fixing it requires a significant refactor of the Identity Service and Registration flow (breaking change).
    - **No Phase 1/2 feature depends on it**. Password authentication is the only supported method for Phases 1 & 2.
    - It is **secure** as implemented (state of art hashing).
- **Triggers for Fix**: 
    - **Phase 3 (Federation)**: When we need to support "Sign in with Google", we will need to create identities without passwords.
    - **Phase 1 Extension (WebAuthn/Passkeys)**: If we implement pure passwordless signup.

## 2. OIDC Protocol ↔ Internal Role Storage

- **Status**: ACCEPTED TECH DEBT
- **Coupling**: The OAuth2 Service (`GenerateIDToken`) directly imports the Tenant domain to fetch roles from the `tenant_user_roles` table.
- **Why we are NOT fixing it now**:
     - Decoupling requires introducing a generic `ClaimsProvider` interface and dependency injection complexity.
     - **No current feature depends on it**. We only support our own internal RBAC system.
     - It does not violate security constraints (roles are scoped correctly).
- **Triggers for Fix**:
    - **Phase 3 (Enterprise Integration)**: When roles must be sourced from an external LDAP or SCIM directory instead of our database.

## 3. Session ↔ Cookie Coupling

- **Status**: ACCEPTED TECH DEBT
- **Coupling**: The Login handler relies exclusively on HTTP Cookies for session management and does not return the Session ID in the JSON response body.
- **Why we are NOT fixing it now**:
    - **Browsers are the only target** for Phases 1, 2, and 3.
    - Cookies are the most secure storage mechanism for browsers (`HttpOnly`).
    - Adding JSON token return encourages insecure storage (LocalStorage) by developers.
- **Triggers for Fix**:
    - **Phase 4 (Developer Experience)**: When we officially support a CLI tool or native mobile SDK that cannot use a cookie jar easily.
