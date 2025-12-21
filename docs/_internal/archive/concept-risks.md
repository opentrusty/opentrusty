# Architectural Risk Analysis

This document identifies discrepancies between the core project definitions (`docs/fundamentals.md`) and the current implementation.

## 1. Identity Mixing

**Identified Issue**
The `internal/identity` service currently handles both **Identity Management** (Profile, User Lifecycle) and **Authentication** (Password Hashing, Credential Storage, Login Logic).
- `CreateUser` accepts a raw password and hashes it.
- `Authenticate` performs password verification.

**Risk**
- **Coupling**: Identity logic is tightly coupled to "Password" credential type. Harder to plug in other credential types (e.g., WebAuthn, FIDO2, Federation) later.
- **Future Violation**: Fundamentals state Identity ≠ Credentials. The current service violates this separation of concerns.

## 2. Authentication Protocol Leakage

**Identified Issue**
The `internal/oauth2` service (Protocol) has direct dependencies on `internal/audit` and `internal/tenant` (Role Repository).
- `GenerateIDToken` fetches roles directly from `tenantRoleRepo`.
- `ExchangeCodeForToken` logs audit events directly.

**Risk**
- **Mixing AuthN & AuthZ**: The Protocol layer is performing Authorization work (fetching and embedding roles). While standard for ID Tokens, it creates a dependency chain where AuthN cannot function if AuthZ storage is down.
- **Ambiguity**: Is the ID Token a claim of identity or a claim of permission? Current implementation implies both.

## 3. Tenant Isolation Assumptions

**Identified Issue**
Tenant ID is often defaulted to `"default"` in the HTTP layer (`getTenantID` helper) and even in some repository methods if not strictly enforced.
- **Explicit Defaulting**: The system is "Single Tenant by Default" rather than "Multi-Tenant by Design" in some pathways.

**Risk**
- **Data Leak**: If a developer forgets to pass `TenantID` from the context, it might silently fall back to `"default"`, potentially exposing the default tenant's data or stranding data in the wrong tenant.
- **Assumption**: The system assumes a "default" tenant always exists.

## 4. Authorization Scope Ambiguity

**Identified Issue**
The `internal/authz` service and `internal/tenant` service both handle "Roles".
- `authz` handles "Project Roles" (RBAC for resources).
- `tenant` handles "Tenant Roles" (Admin permissions).

**Risk**
- **Confusion**: Two different "Role" concepts exist (Project-scoped vs Tenant-scoped) but might be conflated in the `roles` claim of the ID Token.
- **Complexity**: Developers might check the wrong role type for a permission.

## 5. Domain Object Coupling

**Identified Issue**
The `User` struct in `internal/identity/user.go` contains a strict `Profile` struct with specific fields (`GivenName`, `Picture`, etc.).

**Risk**
- **Rigidity**: If an application requires custom profile attributes (e.g., `EmployeeID`, `Department`), the core Identity schema must be modified. Identity is coupled to a specific application's profile requirements.
