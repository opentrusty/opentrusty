# UI E2E Test Matrix

This matrix defines the critical user journeys that MUST execute successfully in a browser environment to verify the system release readiness.

## A. Platform Bootstrap

| ID | Test Case | Steps | Expected Result |
| :--- | :--- | :--- | :--- |
| **UI-01** | **Initial Admin Login** | 1. Navigate to `/admin/`<br>2. Detect Login Page<br>3. Enter Bootstrap Credentials<br>4. Click "Sign in" | Redirects to Dashboard. "Dashboard" and "Tenants" menus visible. |

## B. Tenant Lifecycle

| ID | Test Case | Steps | Expected Result |
| :--- | :--- | :--- | :--- |
| **UI-02** | **Create Tenant** | 1. Navigate to "Tenants"<br>2. Click "Create Tenant"<br>3. Form input: Name "E2E Tenant"<br>4. Submit | Toast Success. New tenant appears in list. |
| **UI-03** | **Tenant Isolation** | 1. Click "Manage" on "E2E Tenant"<br>2. Verify URL contains `/tenants/{id}`<br>3. Verify title matches tenant name | Tenant dashboard loads with context-specific actions. |

## C. Client Lifecycle

| ID | Test Case | Steps | Expected Result |
| :--- | :--- | :--- | :--- |
| **UI-04** | **Register OIDC Client** | 1. In Tenant context, go to "Clients"<br>2. Click "Register Client"<br>3. Name: "Demo App", Redirect URI: `http://localhost:8081/callback`<br>4. Submit | Client created. `client_id` and `client_secret` displayed in modal/page. |

## D. OIDC Login Flow

| ID | Test Case | Steps | Expected Result |
| :--- | :--- | :--- | :--- |
| **UI-05** | **Initiate Auth** | 1. Open Demo App (`:8081`)<br>2. Click "Login"<br>3. Redirected to OpenTrusty | URL matches `http://localhost:8080/oauth2/authorize`. Login page rendered. |
| **UI-06** | **User Authentication** | 1. Enter Tenant User credentials<br>2. Submit | Redirect back to Demo App (`:8081/callback`). |
| **UI-07** | **Token Verification** | 1. Demo App displays "Logged In"<br>2. Inspect ID Token claims | Token contains valid `sub`, `aud` (Client ID), and `iss` (OpenTrusty). |

## E. Observability & Audit

| ID | Test Case | Steps | Expected Result |
| :--- | :--- | :--- | :--- |
| **UI-08** | **Audit Log Verification** | 1. Return to Admin Console<br>2. Navigate to "Audit Logs"<br>3. Filter by "Login" | Entry present for the OIDC login event with correct timestamp and actor. |
