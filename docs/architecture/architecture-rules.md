# OpenTrusty Architecture Rules

This document outlines the normative architectural rules for the OpenTrusty project. These rules represent the "constitution" of the system and MUST be followed by all contributors. Failure to adhere to these rules constitutes a rejection criteria for any code change.

## 1. Identity

- **1.1. Persistence**: An **Identity** MUST serve as a persistent, unique representation of an actor within a **Tenant**. It MUST be distinct from the authentication credentials used to verify it.
- **1.2. Decoupling**: An Identity MUST NOT be owned by or coupled to a specific Application (`OAuthClient`). Identities are owned by the **Tenant**.
- **1.3. Uniqueness**: Identity identifiers (e.g., ID, Email) MUST be unique within the scope of their **Tenant**.
- **1.4. Lifecycle**: Identity lifecycle (Creation, Deactivation) MUST be independent of Application lifecycle. Deleting an Application MUST NOT delete its users.

## 2. Tenant

- **2.1. Isolation**: The **Tenant** MUST be the primary security boundary. All data (Users, Sessions, Tokens, configuration) MUST be scoped to a specific Tenant.
- **2.2. Enforcement**: Multi-tenancy MUST be enforced at the persistence layer (Row-Level Security or mandatory `tenant_id` columns). It MUST NOT be a purely logical application-layer filter.
- **2.3. No Cross-Talk**: Data from one Tenant MUST NOT represent, reference, or leak into another Tenant under any circumstance.

## 3. Protocol (OAuth2 / OIDC)

- **3.1. Standardization**: Authentication MUST be implemented exclusively via standard **OAuth2** and **OpenID Connect** flows. Proprietary login APIs or "magic links" outside these standards SHOULD be avoided unless strictly necessary for bootstrapping.
- **3.2. Strict Compliance**: Implementation MUST adhere strictly to RFC specifications (e.g., exact string matching for `redirect_uri`).
- **3.3. Client Secrets**: `client_secret` MUST NOT be passed in URL parameters. It MUST NOT be logged in plain text. It MUST be stored using cryptographic hashing.
- **3.4. Token Scope**: Access Tokens MUST be scoped to the specific Tenant and Permissions requested. They MUST NOT grant global super-admin privileges unless explicitly scoped for system administration.

## 4. Authorization

- **4.1. Source of Truth**: The Identity Provider (OpenTrusty) MUST serve as the Source of Truth for *assignments* (Who has What Role).
- **4.2. Enforcement point**: The Resource Server (Application) MUST serve as the Point of Enforcement (Is this Role allowed to do X?). The IdP MUST NOT encode granular application-specific permission logic (e.g., "can_click_blue_button").
- **4.3. Reusability**: Roles associated with a Tenant (e.g., "Employee", "Manager") MUST be reusable across multiple Applications. They MUST NOT be hard-coded to a single `OAuthClient`.

## 5. Security

- **5.1. Session Management**: Primary user sessions (Browser Login) MUST be **Stateful** (Database-backed). Stateless tokens (JWT) MUST NOT be used for primary session management due to revocation complexity.
- **5.2. Cryptography**: 
    - Passwords MUST be hashed using **Argon2id**.
    - Weak algorithms (MD5, SHA1, unchecked bcrypt) MUST NOT be used.
- **5.3. Cookies**: All authentication cookies MUST be marked `HttpOnly`, `Secure`, and `SameSite=Lax` (or `Strict`).
- **5.4. Secrets Management**: Secrets (API Keys, Client Secrets, Private Keys) MUST be encrypted at rest or hashed. They MUST NEVER be committed to version control.

## 6. Observability

- **6.1. Structured Logging**: All logs MUST be structured (JSON/Key-Value). Unstructured text logs are forbidden.
- **6.2. Audit Trails**: All security-critical events (Login Success/Failure, Token Issuance, Password Change, Role Assignment) MUST be emitted to the Audit Log.
- **6.3. Audit Context**: Audit events MUST include the **Who** (ActorID), **Where** (TenantID, IP Address), **When** (Timestamp), and **What** (Action/Resource).
- **6.4. Privacy**: Personally Identifiable Information (PII) and Secrets (Passwords, Tokens) MUST be redacted or masked in all logs.

## 7. Development Practices

- **7.1. Clean Architecture**: Business logic MUST reside in the Domain layer (`internal/{domain}`). It MUST NOT leak into the Transport layer (HTTP Handlers).
- **7.2. Dependencies**: The project SHOULD favor the Go Standard Library. External dependencies MUST be justified by significant complexity reduction or security necessity.

## 8. Docs Governance

OpenTrusty documentation is classified into three tiers to ensure clarity of commitment and professional release management:

### 1. Versioned Contract Docs
Documentation that defines the behavior, security properties, and integration contracts of a specific release. Permanent and versioned alongside the code.
- **Location**: `docs/api/`, `docs/architecture/`, `docs/security/`, `docs/fundamentals/`, `docs/domain/`, `docs/deployment/`, `docs/audit/`, `docs/operations/`.
- **Policy**: Must be published per release. Any change to these requires a corresponding logic verification or conscious architectural decision.

### 2. Governance Docs
Documents defining the "rules of the game" and project management.
- **Location**: `docs/governance/`, `GOVERNANCE.md`, `CONTRIBUTING.md`.
- **Policy**: "Latest Only" - reflects current project state and policies regardless of binary version.

### 3. Internal / Historical Docs
Working documents, internal plans, and historical context that are not part of the public product.
- **Location**: `docs/_internal/`.
- **Policy**: Never published to the public documentation site. Used for developer context and audit trails of decision-making.
