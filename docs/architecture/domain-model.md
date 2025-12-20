# OpenTrusty Domain Model

This document outlines the canonical domain entities derived from the system's constitution (`docs/fundamentals.md`).

## 1. Tenant

- **Purpose**: The root container for all data isolation, representing a distinct customer or environment.
- **Belongs To**: System (Root Entity).
- **Scope**: Global (Top-level identifier).
- **Mutability**: **Mutable** (Name, Config, Status).

## 2. User (Identity)

- **Purpose**: A persistent digital representation of an actor within a specific tenant.
- **Belongs To**: **Tenant**.
- **Scope**: **Tenant-Scoped** (Unique by `TenantID` + `Email/ID`).
- **Mutability**: **Mutable** (Profile attributes, Lockout state).

## 3. Credential

- **Purpose**: A secret or proof (e.g., Password Hash, WebAuthn Key) used to verify an Identity.
- **Belongs To**: **User**.
- **Scope**: **Tenant-Scoped** (via User).
- **Mutability**: **Mutable** (Can be rotated/changed).

## 4. Session

- **Purpose**: Represents an active, interactive authentication state (e.g., browser login).
- **Belongs To**: **User**.
- **Scope**: **Tenant-Scoped**.
- **Mutability**: **Mutable** (Last Seen timestamp updates, Expiry extensions).

## 5. OAuthClient

- **Purpose**: An application (RP) registered to request authentication or access on behalf of users.
- **Belongs To**: **Tenant**.
- **Scope**: **Tenant-Scoped**.
- **Mutability**: **Mutable** (Config, Secrets, Redirect URIs).

## 6. AuthorizationCode

- **Purpose**: A short-lived, transient artifact proving user consent, exchanged for tokens.
- **Belongs To**: **User** and **OAuthClient**.
- **Scope**: **Tenant-Scoped**.
- **Mutability**: **Mutable** (State changes to `Used`, otherwise Immutable data).

## 7. AccessToken

- **Purpose**: A standardized, time-limited credential permitting access to resources/APIs.
- **Belongs To**: **User** (Resource Owner) and **OAuthClient**.
- **Scope**: **Tenant-Scoped**.
- **Mutability**: **Immutable** (Revocation status is external state, the token itself is fixed).

## 8. Role

- **Purpose**: A named collection of permissions or a label representing a function/job.
- **Belongs To**: **Tenant** (for Admin Roles) or **Project/Resource** (for functional RBAC).
- **Scope**: **Tenant-Scoped** or **Resource-Scoped**.
- **Mutability**: **Mutable** (Name, Description, Permission set).

## 9. Scope (OAuth2)

- **Purpose**: A string identifier requesting specific access rights or information (e.g., `openid`, `profile`, `roles`).
- **Belongs To**: System (Standardized) or **OAuthClient** (Allowed scopes).
- **Scope**: **Global** (Definitions) / **Tenant-Scoped** (Assignment).
- **Mutability**: **Immutable** (The definition is static config).

## 10. Forbidden Couplings

### User owned by OAuthClient
- **Constraint**: A User MUST NOT be scoped or "belong" to a specific `OAuthClient`.
- **Danger**: This creates data silos (Application Identity) instead of Centralized Identity. It destroys Single Sign-On (SSO) capabilities and forces users to manage separate credentials for every application within the same organization.

### Roles defined by OAuthClient
- **Constraint**: Roles MUST NOT be hard-coupled to a specific `OAuthClient` ID in the core domain models.
- **Danger**: If roles are strictly app-specific, the Identity Provider becomes a tight coupling point for application logic. It prevents role re-use across microservices (e.g., a "Manager" role should be recognized by both the Dashboard App and the Reporting API).

### OAuthClient owning Users
- **Constraint**: Deleting an `OAuthClient` MUST NOT cascade to delete `User` entities.
- **Danger**: This implies the application owns the identity. In an IdP, the Identity is paramount and independent of the services it accesses. Deleting a client is a configuration change; deleting a user is a distinct lifecycle event.
