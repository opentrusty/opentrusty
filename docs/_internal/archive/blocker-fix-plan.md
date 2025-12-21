# Blocker Fix Plan

This document outlines the remediation plan for issues identified as **BLOCKERS** in the `conformance-report.md`. These issues represent critical violations of the `architecture-rules.md` and `domain-model.md` that jeopardize the system's security or evolutionary capacity.

## Blockers Identified

### 1. Implicit Tenant Context (Transport Layer)
*Elevated from [RISK] to [BLOCKER] due to Security Standard violation.*

- **Rule Violation**: **Rule 2.1 (Isolation)** ("Multi-tenancy MUST be enforced... It MUST NOT be a purely logical application-layer filter") & **Rule 2.3 (No Cross-Talk)**.
- **Domain**: **Tenant** / **Protocol**.
- **Violation Type**: **Boundary Violation**.
    - The Transport Layer (`http`) is making a domain decision (defaulting to "default" tenant) instead of rejecting an invalid request.
    - It blurs the boundary between "Application" logic and "Platform" enforcement.
- **Risk**:
    - **Data Leakage**: A client forgetting a header gets silently routed to the wrong tenant instead of receiving an error.
    - **Single-Tenant Assumption**: Hardcodes a "default" tenant, making genuine multi-tenancy brittle.

### 2. Identity-Credential Coupling (Service Layer) [DONE]
*Elevated from [WARNING] to [BLOCKER] for Phase 3 (Federation).*

- **Rule Violation**: **Rule 1.1 (Persistence)** ("Identity... MUST be distinct from the authentication credentials").
- **Domain**: **Identity**.
- **Violation Type**: **Conceptual Violation**.
    - The Service Layer conflates "Creating an Identity" with "Setting a Password".
- **Risk**:
    - **Blocks Evolution**: Impossible to support Passwordless, Social Login, or OIDC Federation properly without hacking the core service.

---

## Fix Sequence

The blockers MUST be fixed in the following order to minimize regression risk and maximize security:

### Step 1: Remediation of Implicit Tenant Context
**Priority**: **IMMEDIATE** (Security Critical)

1.  **Stop the Bleeding**: Modify `internal/transport/http/handlers.go` helper `getTenantID`.
2.  **Enforce Explicit Context**: If `X-Tenant-ID` header or `tenant_id` query param is missing:
    - Return `400 Bad Request` immediately.
    - **Exception**: Health checks, Public Landing Pages (if any).
3.  **Audit**: Ensure all existing clients (if any) send the header.

**Why First?**
- This is a silent security failure mode.
- It requires no database changes, only transport logic changes.
- It "seals" the isolation boundary before we build more features on top of it.

### Step 2: Decoupling of Identity & Credentials [DONE]
**Priority**: **HIGH** (Architectural Debt)

> [!IMPORTANT]
> **Constraint**: No further Identity domain refactor unless required by a new **BLOCKER**.
> **Rationale**: Prevent churn and protect the newly established architectural boundary.

1.  **Split the Atom**: Refactor `identity.Service.CreateUser` into two distinct atomic operations:
    - `CreateIdentity(ctx, tenantID, profile)` -> Returns `User`
    - `SetMetadata` / `SetCredential(ctx, userID, type, secret)`
2.  **Orchestrate**: Update `Register` handler to call them in a transaction (or saga).
3.  **Validate**: Ensure `User` can exist without `Credentials` (for Federation start) or `Credentials` can be added later.

**Why Second?**
- It involves refactoring core business logic and transaction boundaries.
- It is necessary for Phase 3 (Federation) but postponing it doesn't open a security hole *today*, unlike Step 1.
