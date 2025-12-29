# Manual Beta Execution Guide

This guide is for a technical reviewer to manually validate the OpenTrusty Beta system.

## 1. Local Service Startup

To run the full OpenTrusty suite locally, follow these steps in separate terminal tabs:

### 1.1. Prerequisites & Env Vars
OpenTrusty requires specific environment variables. Create a `.env` file or export them:
```bash
export DB_HOST=localhost
export DB_PORT=5433
export DB_USER=opentrusty
export DB_PASSWORD=opentrusty_dev_password
export DB_NAME=opentrusty
export OPENID_KEY_ENCRYPTION_KEY=12345678901234567890123456789012
export OT_BOOTSTRAP_ADMIN_EMAIL=admin@platform.local
```

### 1.2. Start Services
1.  **PostgreSQL**: `make dev` (starts a Docker container on port 5433).
2.  **OpenTrusty Core**:
    ```bash
    # Run migrations first
    go run cmd/server/main.go migrate
    # Start the server (Auth + Admin planes)
    go run cmd/server/main.go serve all
    ```
3.  **Control Panel**:
    ```bash
    cd ../opentrusty-control-panel
    npm run dev
    ```
4.  **Demo Client**:
    ```bash
    cd ../opentrusty-demo-app
    # Ensure CLIENT_ID and CLIENT_SECRET are set after Step 3
    go run main.go
    ```

## 2. Getting the Initial Password

When you run `serve all` for the first time on a fresh database, OpenTrusty auto-generates a platform admin password.

### Option A: From Terminal Output
Watch the `stdout` when launching the server. Look for:
`Password: <random_string>`

### Option B: From Logs
If the server is running in the background or you missed the output:
```bash
grep -a "^Password:" opentrusty.log | head -1 | sed 's/Password: //'
```
*(Note: If you are using the automated test suite, check `tests/local/reports/opentrusty.log` instead.)*

## 3. Environment Readiness
Run the automated health check to confirm everything is up:
```bash
bash scripts/local-beta-smoke.sh
```
Ensure it outputs: `✅ All services are healthy and reachable.`

## 4. Beta Validation Sequence

### Step 1: Platform Setup & Login
1.  Open Chrome/Firefox and navigate to `http://localhost:5173/admin/`.
2.  Log in using the bootstrap credentials (configured via `BOOTSTRAP_ADMIN_EMAIL` and `BOOTSTRAP_ADMIN_PASSWORD`).
3.  **Validation**: You should see the "Platform Overview" dashboard.

### Step 2: Create a Tenant
1.  Go to **Platform -> Tenants**.
2.  Create a new tenant named `Beta Org`.
3.  Copy the generated `Tenant ID`.
4.  **Validation**: The new tenant should appear in the sidebar or tenant switcher.

### Step 3: Register an OIDC Client
1.  Select the `Beta Org` tenant.
2.  Go to **Tenant -> Clients**.
3.  Click **Register Client**.
4.  **Name**: `Demo App`
    **Type**: `Web Application (Auth Code + PKCE)`
    **Redirect URI**: `http://localhost:8081/callback`
5.  Submit. **IMPORTANT**: Securely copy the `Client ID` and `Client Secret` shown on the screen.
6.  **Validation**: Client is listed.

### Step 4: Provision a User
1.  Go to **Tenant -> Users**.
2.  Click **Create User**.
3.  **Email**: `tester@example.com`
    **Password**: `password123`
4.  **Validation**: User is created.

### Step 5: External App Authentication
1.  Stop the `opentrusty-demo-app` if running.
2.  Export the credentials:
    ```bash
    export CLIENT_ID="<your_client_id>"
    export CLIENT_SECRET="<your_client_secret>"
    export REDIRECT_URI="http://localhost:8081/callback"
    export AUTH_URL="http://localhost:8080"
    ```
3.  Start the app: `go run main.go`.
4.  Open `http://localhost:8081`.
5.  Click **Login with OpenTrusty**.
6.  Enter `tester@example.com` / `password123`.
7.  **Validation**: You should reach the "Login Successful!" page with a valid ID Token.

---
*Status: Beta Validation Baseline*
