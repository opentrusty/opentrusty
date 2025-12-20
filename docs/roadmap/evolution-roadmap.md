# OpenTrusty Evolution Roadmap

This document defines the **ONLY** allowed order of feature evolution for the OpenTrusty project. Features MUST be implemented in the order of phases defined below. Jumping ahead to a later phase before completing the current phase is strictly **FORBIDDEN**.

**Core Constraints:**
1. **Security before Convenience**: Hardening takes precedence over ease of use.
2. **Multi-tenancy before Federation**: Isolation must be perfect before external integrations are added.
3. **Protocol before Experience**: The API must be standard-compliant before SDKs or UIs are built.

---

## Phase 1: Security Hardening & Protocol Compliance (Current Focus)

**Goal**: Establish an unassailable security foundation and strict RFC compliance for OAuth2/OIDC.

### ✅ Allowed work
- **MFA Implementation**: TOTP, WebAuthn/FIDO2 support.
- **Session Security**: Device binding, concurrent session limits, revocation API.
- **Protocol Completeness**: Full OIDC Discovery, UserInfo claims, PKCE enforcement for all flows.
- **Observability**: Structured audit logs for *every* write operation.

### ❌ Explicitly Forbidden
- **Social Login / Federation**: No "Login with Google" or upstream IdPs.
- **Admin UI**: No graphical interface for management; API only.
- **SDKs**: No client libraries (Go/JS/Python clients); raw HTTP usage only.
- **Registration APIs**: Public sign-up flows (invite-only for now).

---

## Phase 2: Deep Multi-Tenancy & Isolation

**Goal**: Ensure the system can host hostile tenants side-by-side with zero leakage risk.

### ✅ Allowed work
- **Tenant Lifecycle**: API for systematic onboarding/offboarding of tenants.
- **Per-Tenant Config**: Custom password policies, token lifetimes, and lockout rules per tenant.
- **Custom Domains**: Support for `auth.customer.com` routing to specific tenant contexts.
- **Data Residency**: Support for sharding data by region/tenant at the repository layer.

### ❌ Explicitly Forbidden
- **Cross-Tenant Sharing**: No "shared" users or roles (except System Super-Admin).
- **Federation**: Still no upstream Identity Providers.
- **Frontend Libraries**: No UI component kits.

---

## Phase 3: Federation & Enterprise Integration

**Goal**: Allow OpenTrusty to act as a bridge to other identity systems (Enterprise functionality).

### ✅ Allowed work
- **Upstream OIDC/SAML**: Ability for a tenant to configure "Login with Okta" or "Login with Google".
- **SCIM Provisioning**: Inbound/Outbound user provisioning via SCIM 2.0.
- **Enterprise Features**: LDAP/Active Directory connectors, Group syncing.

### ❌ Explicitly Forbidden
- **Proprietary UI**: No "Hosted Login Page" logic that isn't purely template-driven.
- **Convenience SDKs**: Still relying on standard OIDC libraries.

---

## Phase 4: Developer Experience & Ecosystem

**Goal**: Reduce the friction of integrating with OpenTrusty *after* the core is proven security-perfect.

### ✅ Allowed work
- **Official SDKs**: Thin wrappers around standard OIDC flows for major languages.
- **UI Kits**: React/Vue/Web Component libraries for login forms.
- **Admin Dashboard**: A canonical web interface for managing tenants and users.
- **CLI Tool**: A developer CLI for managing the instance.

### ❌ Explicitly Forbidden
- **Non-Standard Protocol Extensions**: No "magic" proprietary APIs that bypass OAuth2 flows for convenience.
