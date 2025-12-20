# Second Blocker Identification: B-TENANT-01

This document identifies the next critical architectural blocker to be resolved: the implicit "default" tenant model in the Transport layer.

## Blocker ID: B-TENANT-01
**Title**: Implicit Tenant Context (Transport Defaulting)

### 1. Violated Rule(s)

> **Rule 2.1. Isolation**: Multi-tenancy MUST be enforced at the entry point of every tenant-scoped request. It MUST NOT be a purely logical application-layer filter.
> **Rule 2.3. No Cross-Talk**: It MUST be impossible to query or modify data belonging to Tenant A using a context authenticated for Tenant B.
> -- *docs/architecture-rules.md*

### 2. Evidence of Violation
- **`internal/transport/http/handlers.go`**: The `getTenantID` helper function defaults to `"default"` if no `X-Tenant-ID` header or `tenant_id` query parameter is present.
- **Impact**: This turns multi-tenancy into an "opt-in" feature at the transport layer, violating the requirement for enforcement at the entry point.

### 3. Why this MUST be fixed before OAuth2/OIDC

1. **Security Invariant**: OAuth2 token issuance and validation depend on knowing exactly which tenant is requesting the token. If the transport layer defaults to "default", a client could accidentally (or maliciously) obtain tokens for the "default" tenant when they intended to target a specific one, or vice versa, leading to **Data Leakage**.
2. **Isolation Guarantee**: One of the primary value propositions of OpenTrusty is strict multi-tenancy. If the system "fails open" to a default tenant, it compromises the isolation guarantee before the protocol even starts.
3. **Protocol Semantic**: OAuth2/OIDC discovery and authorization flows are often tenant-specific (e.g., `/.well-known/openid-configuration` URLs). If the system cannot reliably resolve the tenant from the transport context, these flows will return incorrect metadata or allow cross-tenant token exploitation.

Resolving **B-TENANT-01** ensures that the "Isolation Boundary" is enforced strictly at the edge, making multi-tenancy a hard requirement for all following protocol implementations.
