# OpenTrusty Evolution Roadmap

This document defines the **ONLY** allowed order of feature evolution for the OpenTrusty project. Features MUST be implemented in the order of phases defined below. Jumping ahead to a later phase before completing the current phase is strictly **FORBIDDEN**.

**Core Constraints:**
1. **Security before Convenience**: Hardening takes precedence over ease of use.
2. **Multi-tenancy before Federation**: Isolation must be perfect before external integrations are added.
3. **Protocol before Experience**: The API must be standard-compliant before SDKs or UIs are built.
4. **API before UI**: Management API endpoints MUST be stable before Control Panel UI depends on them.

**Architectural Boundaries:**
- **Authentication Plane**: OIDC/OAuth2 endpoints and server-rendered login pages (part of core binary)
- **Management API Plane**: Admin REST APIs for tenant/user/client management (part of core binary)
- **Control Panel UI**: Separate repository (`opentrusty-control-panel`), NOT embedded in core

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
- **Control Panel UI**: No graphical admin interface in this repo; Management API only.
- **SDKs**: No client libraries (Go/JS/Python clients); raw HTTP usage only.
- **Registration APIs**: Public sign-up flows (invite-only for now).

**Clarification**: Server-rendered login/consent pages belong to the **Authentication Plane** and are part of core. They are NOT "Admin UI".

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
- **Control Panel UI Features**: No UI work until Management API is stable.

**Phase 2 Enables**:
- **Branding Metadata API**: Endpoints for tenant-customizable login page branding.
- Login page branding is **Authentication Plane** (server-rendered), not Control Panel.

---

## Phase 3: Federation & Enterprise Integration

**Goal**: Allow OpenTrusty to act as a bridge to other identity systems (Enterprise functionality).

### ✅ Allowed work
- **Upstream OIDC/SAML**: Ability for a tenant to configure "Login with Okta" or "Login with Google".
- **SCIM Provisioning**: Inbound/Outbound user provisioning via SCIM 2.0.
- **Enterprise Features**: LDAP/Active Directory connectors, Group syncing.

### ❌ Explicitly Forbidden
- **Proprietary UI Logic**: No hosted login page logic that isn't template-driven.
- **Convenience SDKs**: Still relying on standard OIDC libraries.
- **Control Panel UI Coupling**: UI remains separate artifact; no embedding.

**Phase 3 Enables**:
- **Branding Templates**: Server-side templates for login/consent pages.
- **Management API Stability**: API surface considered stable for UI consumption.

---

## Phase 4: Developer Experience & Ecosystem

**Goal**: Reduce the friction of integrating with OpenTrusty *after* the core is proven security-perfect.

### ✅ Allowed work
- **Official SDKs**: Thin wrappers around standard OIDC flows for major languages.
- **Control Panel UI**: Full admin dashboard in `opentrusty-control-panel` repository.
- **UI DX Improvements**: React/Vue/Web Component libraries for login forms (in UI repo).
- **CLI Tool**: A developer CLI for managing the instance.

### Dependencies
- Control Panel UI MUST consume stable Management API (Phase 3+).
- Control Panel UI is a **separate deployment artifact**, never embedded in core.

### ❌ Explicitly Forbidden
- **Non-Standard Protocol Extensions**: No "magic" proprietary APIs that bypass OAuth2 flows for convenience.
- **UI in Core**: Control Panel code MUST NOT be added to `opentrusty` repository.
