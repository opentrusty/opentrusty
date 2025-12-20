# First Blocker Identification: B-IDENTITY-01

This document identifies the primary architectural blocker to be resolved within the Identity domain of OpenTrusty.

## Blocker ID: B-IDENTITY-01
**Title**: Identity-Credential Coupling (Hardcoded Password Dependency)

### 1. Violated Rule(s)
> **Rule 1.1. Persistence**: An **Identity** MUST serve as a persistent, unique representation of an actor within a **Tenant**. It MUST be distinct from the authentication credentials used to verify it.
> -- *docs/architecture-rules.md*

### 2. Affected Packages & Files
- `internal/identity/service.go`: The `CreateUser` method signature and logic are hardcoded to require a password.
- `internal/identity/user.go`: The domain models for `User` and `Credentials` are co-located, and the `UserRepository` interface forces atomic creation of both.
- `internal/store/postgres/user_repository.go`: The persistence layer implements a transaction that assumes every Identity comes with a Password.

### 3. Classification: Boundary Violation
This is a **Boundary Violation**. The Identity domain (which should only care about *who* an actor is and *which tenant* they belong to) has absorbed the Authentication domain (which cares about *how* they prove their identity). By forcing a password secret into the `Identity.Service.CreateUser` method, the boundary between Identity and Security has been breached.

### 4. Why this MUST be the first fix
While there are transport-level risks (e.g., Tenant defaulting), `B-IDENTITY-01` is the most fundamental **architectural** blocker for the following reasons:

1. **Protocol Rigidity**: It prevents OpenTrusty from fulfilling its mission as a modern OIDC provider. You cannot currently "provision" a user (via admin or invitation) without a password, nor can you integrate external Federation (Phase 3) where the password is held by a third party.
2. **Multi-Tenant Ambiguity**: In an enterprise multi-tenant setup, an identity might migrate between authentication methods (e.g., from Password to SAML). If the ID and Password are tied at the root creation method, this migration requires deleting and recreating the identity, which destroys audit trails and persistent IDs.
3. **Foundation for Security Hardening**: Phase 1 requires MFA and WebAuthn. If the core service assumes a password is the "primary" credential, adding secondary or substitute credentials will result in "wrapper debt" where the system always expects a password that might not exist.

Fixing this decoupling immediately clarifies the `User` lifecycle and paves the way for both strict multi-tenancy and standard-compliant federation.
