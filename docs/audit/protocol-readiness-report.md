# Protocol Readiness Report: B-PROTOCOL-01

This document assesses whether the OAuth2 and OpenID Connect (OIDC) protocol boundaries in OpenTrusty are explicitly modeled or remain implicit.

## 1. Assessment Results

### 1.1 Protocol Entry Isolation
- **Finding**: **IMPLICIT**.
- **Evidence**: All OAuth2/OIDC handlers (`Authorize`, `Token`, `UserInfo`, `GetOIDCConfiguration`) are methods of the monolithic `Handler` struct in `internal/transport/http/handlers.go`. This struct is shared with business logic for identity provisioning, profile updates, and tenant management.
- **Violation**: **Rule 2 (Domain-Driven Architecture)**. Protocol handling is co-located with management logic, preventing clean isolation of security policies and dependencies.

### 1.2 Protocol State Modeling
- **Finding**: **EXPLICIT**.
- **Evidence**: `internal/oauth2/models.go` defines dedicated domain objects for `AuthorizationCode`, `AccessToken`, and `RefreshToken`. These are distinct from browser `Session` objects.
- **Status**: Satisfied.

### 1.3 Protocol Error Modeling
- **Finding**: **IMPLICIT**.
- **Evidence**: In `internal/transport/http/oauth2_handler.go`, handlers manually switch on domain errors to produce RFC-compliant JSON (`invalid_request`, `invalid_client`, etc.). The logic for what constitutes a protocol error is trapped in the transport layer rather than being an intrinsic part of the Protocol Domain.
- **Violation**: **Rule 4 (OAuth2 / OIDC Compliance)**. "Strict RFC behavior" is currently a manual transport-layer translation rather than a domain-enforced invariant.

---

## 2. NEW BLOCKER: B-PROTOCOL-01
**Title**: Implicit Protocol Boundary (Monolithic Handler)

### Violated Rule(s)

> **Rule 4. OAuth2 / OIDC Compliance**: Strict RFC behavior MUST be enforced... [It] MUST NOT be logged or returned.
> **Rule 2. Domain-Driven Architecture**: Domain logic lives ONLY in `internal/*`. HTTP layer: Parse request -> Call domain service -> Map domain error → HTTP response.
> -- *docs/architecture-rules.md*

### Why Protocol Implementation Must Wait

The current implementation is an "implicit" protocol layer. Proceeding with features (e.g., Refresh Tokens, PKCE enforcement, OIDC Federation) on this foundation will result in:

1. **Brittle Error Handling**: Every new protocol feature will require manual error-mapping logic in the transport layer, increasing the risk of non-compliant responses (RFC 6749 requires specific `error` and `error_description` fields).
2. **Security Risk**: The co-location of protocol logic with management logic in a single `Handler` struct increases the attack surface. A bug in a management endpoint could theoretically compromise the protocol state if not strictly isolated.
3. **Hardcoded Semantics**: Critical Discovery metadata (e.g., `issuer`, `jwks_uri`) is currently hardcoded in the handler. These MUST be modeled as domain-driven entities that support multi-tenancy before the protocol is "ready".

**Decision**: **NOT READY**. We must formalize the **Protocol Domain** and isolate its transport adapters before finishing the RFC implementation.
