# Beta Validation Test Cases (A–I)

This document enumerates the mandatory test cases for OpenTrusty Beta classification.

## Test Case A: Platform Bootstrapping
- **Preconditions**: Clean database, environment variables set.
- **Actions**: Run `go run cmd/server/main.go serve all`.
- **Expected Results**:
    - System starts successfully.
    - Log indicates "bootstrap successful".
    - Initial platform admin user exists in `users` table.

## Test Case B: Tenant Creation (UI)
- **Preconditions**: Logged into Control Panel as Platform Admin.
- **Actions**: Navigate to "Platform -> Tenants", click "New Tenant", enter "Beta Tenant", submit.
- **Expected Results**:
    - Tenant appears in list.
    - `tenant_id` (UUID v7) is generated.

## Test Case C: OAuth2 Client Registration (UI)
- **Preconditions**: Tenant created.
- **Actions**: Select "Beta Tenant", navigate to "Clients", click "Register Client", select "Web Application", enter redirect URI `http://localhost:8081/callback`.
- **Expected Results**:
    - `client_id` and `client_secret` are generated and displayed once.
    - Client appears in the tenant's client list.

## Test Case D: User Provisioning
- **Preconditions**: Tenant created.
- **Actions**: Navigate to "Tenant -> Users", click "Create User", enter email `beta-user@example.com`, set password.
- **Expected Results**:
    - User is created and assigned to the tenant.
    - User can successfully log in via the Auth Plane.

## Test Case E: OIDC Discovery & Metadata
- **Preconditions**: OpenTrusty Core is running.
- **Actions**: `curl http://localhost:8080/.well-known/openid-configuration`.
- **Expected Results**:
    - Valid JSON returned.
    - `issuer` matches expected local URL.
    - `authorization_endpoint` and `token_endpoint` are present.

## Test Case F: End-to-End OIDC Flow (Demo App)
- **Preconditions**: Client registered, User provisioned, Demo App running.
- **Actions**: Open Demo App (`localhost:8081`), click "Login", enter User credentials on OpenTrusty login page, approve consent.
- **Expected Results**:
    - Redirected back to Demo App.
    - Demo App displays "Login Successful!".
    - Raw ID Token is visible.

## Test Case G: Token Exchange & Validation
- **Preconditions**: Test Case F completed.
- **Actions**: Inspect the JWT payload of the `id_token` in the Demo App.
- **Expected Results**:
    - `sub` matches User ID.
    - `iss` matches OpenTrusty issuer.
    - `aud` matches Client ID.
    - `nonce` matches (if provided).

## Test Case H: Session Isolation (Multi-Tenancy)
- **Preconditions**: Two tenants (A and B) created. User 1 in Tenant A.
- **Actions**: Attempt to access Tenant B's admin data using User 1's session.
- **Expected Results**:
    - Request rejected with `403 Forbidden` or `404 Not Found`.
    - No data leakage between tenants.

## Test Case I: Audit Log Generation
- **Preconditions**: Any state-changing action (e.g., Client Registration).
- **Actions**: Query `audit_logs` table or view "Platform -> Audit Logs" (if implemented).
- **Expected Results**:
    - Entry exists for the action.
    - Includes `actor_id`, `tenant_id`, `action_type`, and `timestamp`.

---
*Status: Beta Validation Baseline*
