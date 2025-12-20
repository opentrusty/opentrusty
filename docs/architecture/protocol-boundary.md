# Protocol Boundary: OAuth2 / OIDC

This document defines the strict architectural boundaries for the OAuth2 and OpenID Connect (OIDC) protocols in OpenTrusty, addressing **B-PROTOCOL-01**.

## 1. Protocol Residence

- **Location**: All protocol-specific logic MUST reside exclusively within `internal/oauth2` and `internal/oidc`.
- **Classification**: These are **Isolated Protocol Modules**. 
- **Definition**: The protocol layer is a **Consumer** of the Platform Core. It is an optional capability that translates platform facts (Identity, Authorization) into standardized external formats (RFC 6749, OIDC Core).

## 2. Forbidden Dependencies

The Protocol Module represents a pure domain of "Delegated Authorization". To maintain modularity and security, it MUST NOT depend on:

- **Transport Handlers**: No knowledge of `net/http`, headers, or cookies. It MUST operate on pure Data Transfer Objects (DTOs).
- **Session Management**: No direct dependency on browser-specific session implementations. It MUST treat authentication status as an abstract input (`IdentityContext`).
- **Platform Management**: No awareness of how tenants are created, how clients are registered via UI, or how the system is configured.
- **External UI**: No dependency on HTML templates, CSS, or frontend routing logic.

## 3. Allowed Interactions

The Protocol Module interacts with the Platform Core through strictly defined interfaces:

- **Identity Domain**: Allowed to call `identity.Service` to retrieve profile facts (claims) and verify user status.
- **Tenant Context**: MUST use the platform-provided `TenantID` for all operations. All protocol entities (Clients, Tokens, Codes) are natively multi-tenant.
- **Authentication Mechanism**: Rely on the platform to provide a "Verified Principal". The protocol layer does not perform password hashing or handle login flows.
- **Persistence**: Allowed to define and use dedicated repositories (`ClientRepository`, `TokenRepository`, `CodeRepository`) for protocol state storage.

## 4. Direction of Dependency

The integrity of the system relies on a unidirectional dependency flow:

### What Depends on Protocol
- **Transport Adapters**: The HTTP layer depends on the Protocol Module to handle the logic of `authorize` and `token` exchanges.
- **OIDC Discovery**: The discovery generator depends on the Protocol Module to provide metadata.

### What Protocol Depends On
- **Core Domain**: Protocol depends on `identity`, `tenant`, and `store` for data and invariants.
- **Cryptography**: Protocol depends on base platform crypto utilities (e.g., for JWT signing).

### Prohibited Dependencies (NEVER)
- **Identity -> Protocol**: The `internal/identity` package MUST NEVER import `internal/oauth2`. A user's identity exists independently of any delegation protocol.
- **Tenant -> Protocol**: Multi-tenancy must function even if OAuth2 is completely disabled.

---
**Rule Citation**: This boundary enforces **Rule 2 (Domain-Driven Architecture)** and **Rule 7 (Open Source & Non-Profit Values)** by preventing monolithic coupling and ensuring the protocol remains a clean "plug-in" capability.

**Forbidden**: Protocol modules MUST NOT introduce or manage user, tenant, or credential lifecycles. They may only consume existing domain abstractions.