# Hidden/Implicit Couplings Report

This document lists "invisible" architectural couplings that violate the domain model and block future evolution.

## 1. Identity Implementation ↔ Password Credential

- **Location**: `internal/identity/service.go:CreateUser`
- **Coupling**: The `CreateUser` function signature requires a `password` string and internally calls the password hasher.
- **Violation**: **Identity 1.1**. Identity entity is not distinct from credentials.
- **Blocks**:
    - **Phase 3 (Federation)**: Cannot create a user from "Sign in with Google" because they define no password.
    - **Phase 1 (MFA/Passkeys)**: Cannot create a "Passkey-only" user.

**Remediation**: Split `CreateUser` into `ProvisionIdentity(profile)` and `SetCredential(userId, type, secret)`.

## 2. HTTP Transport ↔ Single Tenant Assumption

- **Location**: `internal/transport/http/handlers.go:getTenantID`
- **Coupling**: The helper function defaults to `"default"` if no tenant context is found.
- **Violation**: **Tenant 2.1**. The system assumes a "Default Tenant" exists, effectively making it Single-Tenant by default.
- **Blocks**:
    - **Phase 2 (Deep Multi-Tenancy)**: High risk of data leakage. If a customized tenant login page fails to send the header, the user might accidentally log into the default tenant (or fail securely).

**Remediation**: Remove the default. Return error 400 (Bad Request) if Tenant ID is missing for tenant-scoped endpoints.

## 3. Session Management ↔ Browser Cookies

- **Location**: `internal/transport/http/handlers.go:Login`
- **Coupling**: The Login handler strictly sets an HTTP Cookie (`http.SetCookie`). It does not return the Session ID in the response body.
- **Violation**: **Protocol 3.1** (Spirit of). Assumes the client is always a Browser.
- **Blocks**:
    - **Phase 4 (CLI Tool)**: A CLI tool cannot easily login because parsing `Set-Cookie` headers is non-trivial and brittle compared to JSON response.
    - **Mobile Apps**: Native apps prefer token-based session handling over cookie jars.

**Remediation**: Return the Session ID in the JSON response body as well as the Cookie.

## 4. OIDC Protocol ↔ Internal Role Storage

- **Location**: `internal/oauth2/service.go:GenerateIDToken`
- **Coupling**: The OIDC service directly imports `internal/tenant` and queries `tenantRoleRepo` to populate the `roles` claim.
- **Violation**: **Protocol layer** depends on **specific storage implementation** of Authorization.
- **Blocks**:
    - **Phase 3 (Enterprise Integration)**: If roles come from an LDAP/AD upstream group mapping, this hard-coded repo call will fail or be bypassed.

**Remediation**: Inject a `ClaimsProvider` interface into the OAuth2 service, allowing different strategies for fetching user claims (DB, LDAP, Remote Policy).
