# Tech Debt Strategy

This document classifies identified architectural issues and technical debt based on the critical path of the [Evolution Roadmap](evolution-roadmap.md).

## 1. Must Fix BEFORE OAuth2/OIDC (Phase 1)
*Issues that compromise the security or correctness of the core protocol implementation.*

### 1.1. Explicit Tenant Context in Transport
- **Issue**: `getTenantID` defaults to `"default"` if header is missing.
- **Why it must be fixed**: While acceptable for single-tenant, it violates "Secure by Design". A security protocol should never rely on implicit magic strings. If we are hardening security in Phase 1, we must ensure every request is explicitly scoped, even if only one tenant exists. Implicit defaults mask configuration errors.
- **Remediation**: Remove default. Return `400 Bad Request` if context is missing.

---

## 2. Must Fix BEFORE Multi-Tenancy (Phase 2)
*Issues that effectively break isolation or management in a multi-tenant environment.*

### 2.1. Tenant Context in Access Token Validation
- **Issue**: `UserInfo` endpoint retrieves Tenant ID from the *User* entity, not the *Access Token*.
- **Why it must be fixed**: In complex multi-tenancy, a user might eventually exist in multiple contexts (or federation). Relying on the User's "home" tenant instead of the "token's" tenant context creates ambiguity.
- **Remediation**: snapshot `TenantID` into the Access Token claims/table at issuance time.

---

## 3. Can Be Postponed Safely (Phases 3 & 4)
*Issues that limit flexibility but do not compromise current security or functionality.*

### 3.1. Identity ↔ Password Coupling
- **Issue**: `CreateUser` requires a password argument.
- **Postponed Until**: **Phase 3 (Federation)**.
- **Reasoning**: Phase 1 & 2 rely on standard username/password (plus MFA). We do not need passwordless users until we import users from external IdPs (Federation) or implement pure Passkeys.

### 3.2. OIDC Protocol ↔ Internal Role Storage
- **Issue**: `GenerateIDToken` directly queries `tenantRoleRepo`.
- **Postponed Until**: **Phase 3 (Enterprise Integration)**.
- **Reasoning**: As long as OpenTrusty is the sole source of roles (RBAC), this coupling is efficient and harmless. Abstraction is only needed when roles come from LDAP/SCIM.

### 3.3. Session ↔ Cookie Coupling
- **Issue**: Login handler only sets cookies, doesn't return Session ID in JSON.
- **Postponed Until**: **Phase 4 (Developer Experience / CLI)**.
- **Reasoning**: Browsers (primary target for Phase 1-3) handle cookies natively. CLI/Mobile support is a later concern.

---

## 4. Never Fix (Intentional Constraints)
*"Fake" issues that are actually design choices.*

### 4.1. Stateful Sessions (No JWT)
- **Observation**: Primary sessions require a DB lookup.
- **Decision**: **Keep as is.**
- **Reasoning**: Immediate revocation is a core security requirement (Rule 5.1). Stateless JWT sessions prevent this.

### 4.2. Lack of "Magic Link" / Proprietary Login APIs
- **Observation**: Login is strict OAuth2 or Basic Auth.
- **Decision**: **Keep as is.**
- **Reasoning**: Adherence to RFC specs (Rule 3.1) prevents vendor lock-in and security drift.
