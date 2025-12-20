# Change Summary: Blocker B-IDENTITY-01

This document summarizes the architectural changes implemented to resolve the Identity-Credential coupling blocker.

## 1. Summary of Code Changes

### Domain Layer (`internal/identity`)
- **[MODIFY] [user.go](file:///Users/mw/workspace/repo/github.com/opentrusty/opentrusty/internal/identity/user.go)**: Refactored the `UserRepository` interface. Split the monolithic `Create(user, credentials)` method into two atomic operations: `Create(user)` and `AddCredentials(credentials)`.
- **[MODIFY] [service.go](file:///Users/mw/workspace/repo/github.com/opentrusty/opentrusty/internal/identity/service.go)**: 
    - Replaced `CreateUser(tenantID, email, password, profile)` with `ProvisionIdentity(tenantID, email, profile)`. 
    - Extracted password logic into a standalone `AddPassword(userID, password)` method.
    - Removed `PasswordHasher` dependency from the identity provisioning logic.

### Persistence Layer (`internal/store/postgres`)
- **[MODIFY] [user_repository.go](file:///Users/mw/workspace/repo/github.com/opentrusty/opentrusty/internal/store/postgres/user_repository.go)**: 
    - Updated `Create` to execute a single-table `INSERT` into `users`.
    - Implemented `AddCredentials` to handle `INSERT` into the `credentials` table.
    - Removed the cross-table transaction logic from the low-level `Create` method.

### Adapter Layer (`internal/transport/http`)
- **[MODIFY] [handlers.go](file:///Users/mw/workspace/repo/github.com/opentrusty/opentrusty/internal/transport/http/handlers.go)**: Refactored the `Register` handler to perform the two-step orchestration: `ProvisionIdentity` followed by `AddPassword`.
- **[MODIFY] [tenant_handler.go](file:///Users/mw/workspace/repo/github.com/opentrusty/opentrusty/internal/transport/http/tenant_handler.go)**: Refactored `ProvisionTenantUser` to use the decoupled service methods.

## 2. Rule Mapping

| Change | Blocker ID | Violated Rule ID |
| :--- | :--- | :--- |
| Interface Split | B-IDENTITY-01 | **Rule 1.1** (Persistence) |
| Service Refactor | B-IDENTITY-01 | **Rule 1.1** (Decoupling Identity from Secret) |
| Atomic Persistence | B-IDENTITY-01 | **Rule 1.1** (Distinct representation) |

## 3. Boundary & Coupling Analysis

### What coupling was removed?
**Identity-Credential Direct Dependency**. 
Previously, a `User` entity could not be persisted or represented at the service level without a corresponding `Argon2id` password hash. The `identity.Service` was hardcoded to assume that birth of an identity equals birth of a password.

### What new boundary now exists?
**Identity Lifecycle Boundary**.
There is now a clear functional boundary between **Identity Provisioning** (defining who an actor is within a tenant) and **Credential Management** (how that actor proves themselves). 

- **Identity Domain**: Owns `User`, `Profile`, and `TenantID`.
- **Credential Domain**: Owns secret verification (Passwords, MFA, etc.).

This allows the system to now satisfy **Rule 1.2 (Decoupling)** and enables Phase 3 (Federation) where an identity may be provisioned before any secret is known, or provisioned with no internal secret at all.
