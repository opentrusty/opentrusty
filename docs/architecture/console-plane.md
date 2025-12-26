# Console Plane Architecture

## Purpose
The Console (`opentrusty-control-panel`) is a client-side Single Page Application (SPA) that provides a user interface for the OpenTrusty system.

## Architecture
-   **Type**: Static Asset (HTML/JS/CSS).
-   **Deployment**: Served via Nginx/CDN (not embedded in Go binary).
-   **Authentication**: Uses Session Cookies shared via top-level domain.

## Interaction Model

### 1. Login
The Console posts credentials to the Auth Plane.
-   **Target**: `https://auth.opentrusty.org/api/v1/auth/login`
-   **Result**: HttpOnly Cookie set for `.opentrusty.org`.

### 2. Management
The Console calls the Admin Plane for data.
-   **Target**: `https://api.opentrusty.org/api/v1/...`
-   **Auth**: Browser automatically sends the HttpOnly Cookie (CORS `credentials: include`).

## Constraints
-   **No Private Knowledge**: The Console MUST NOT have direct DB access.
-   **Untrusted Client**: The Backend treats the Console as untrusted; all inputs are validated.
