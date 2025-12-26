# Stage 4 Manual Smoke Test

## Objective
Verify the "Minimal Production Skeleton" with split Auth/Admin planes.

## Prerequisites
1.  **Database**: Clean & Bootstrapped.
2.  **Services**:
    -   `serve auth` on Port 8080 (Auth Plane).
    -   `serve admin` on Port 8081 (Admin Plane).
    -   `npm run dev` (Vite) on Port 5173 (Console).

## Test Steps

### 1. Auth Plane Verification
-   **Action**: Open `http://localhost:8080/.well-known/openid-configuration`
-   **Expect**: JSON response with OIDC configuration.

### 2. Admin Plane Verification
-   **Action**: Open `http://localhost:8081/health`
-   **Expect**: `{"status": "pass", ...}`

### 3. Console Login (Crossing Planes)
-   **Action**:
    -   Go to `http://localhost:5173/admin/login`
    -   Enter Bootstrap Credentials (`admin@platform.com` / `<generated>`)
    -   Click "Login"
-   **Mechanism**:
    -   POST `/api/v1/auth/login` -> Proxied to **8080**
    -   Cookie Set (HttpOnly)
    -   Redirect to Dashboard
    -   GET `/api/v1/auth/me` -> Proxied to **8081** (Admin checks session)
-   **Expect**: Successful login, redirection to Dashboard.

### 4. Tenant Management (Admin Plane)
-   **Action**: Click "Tenants" sidebar item.
-   **Mechanism**:
    -   GET `/api/v1/tenants` -> Proxied to **8081**
-   **Expect**: List of tenants (likely empty or just the bootstrap context).

## Pass Criteria
-   All steps complete without 404s or 500s.
-   Console Network tab confirms correct backends are hit (via Proxy).
