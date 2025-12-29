# Manual Beta Validation Guide

This guide is for a human operator with **zero project context** to validate the OpenTrusty Beta.

---

## 1. Prerequisites

### 1.1 Software Requirements
*   Go 1.23+
*   Node.js 20+ and npm 10+
*   PostgreSQL 16+
*   Docker & Docker Compose (optional, for easy DB setup)

### 1.2 Clone Repositories
```bash
# Core Backend
git clone https://github.com/opentrusty/opentrusty.git
# Control Panel UI
git clone https://github.com/opentrusty/opentrusty-control-panel.git
```

---

## 2. Environment Setup

### 2.1 Start PostgreSQL

**Option A: Docker Compose (Recommended)**
```bash
cd opentrusty
docker-compose up -d
```
*Expected:* `postgres` container running on port `5432`.

**Option B: Native PostgreSQL**
```bash
createdb opentrusty
```

### 2.2 Configure Environment

Create `.env` in the `opentrusty` directory:
```env
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/opentrusty?sslmode=disable

# OTEL (optional, leave empty if not using)
OTEL_EXPORTER_OTLP_ENDPOINT=

# Bootstrap Admin (required for first run)
BOOTSTRAP_EMAIL=admin@example.com
BOOTSTRAP_PASSWORD=adminadmin
```

---

## 3. Service Startup

Start services in **this order**:

### Step 1: Build and Run Core Backend
```bash
cd opentrusty
make build
./bin/opentrusty
```
*Expected Output:*
```
{"level":"INFO","msg":"http_server_started","addr":":8080"}
```
*Listening on:* `http://localhost:8080`

### Step 2: Start Control Panel UI
```bash
cd opentrusty-control-panel
npm install
npm run dev
```
*Expected Output:*
```
  VITE v6.x.x  ready in xxx ms
  ➜  Local:   http://localhost:5173/
```
*Listening on:* `http://localhost:5173`

> [!NOTE]
> **OIDC Demo Client**
> A dedicated demo client application is not required for Beta. The OIDC flow can be verified by registering a client and observing the client credentials.

---

## 4. Health Checks

| Service | URL | Expected |
|---------|-----|----------|
| Backend API | `http://localhost:8080/api/v1/health` | `200 OK` (or 404 if not implemented) |
| Control Panel | `http://localhost:5173/admin/login` | Login page renders |

---

## 5. Golden Path: End-to-End Scenario

### 5.1 Platform Admin Login
1.  Open `http://localhost:5173/admin/login`
2.  Enter `BOOTSTRAP_EMAIL` and `BOOTSTRAP_PASSWORD` from `.env`
3.  Click **Login**
4.  **Expected:** Redirect to `/admin/platform/overview`

### 5.2 Create a Tenant
1.  Navigate to **Tenants** in the sidebar.
2.  Click **Create Tenant**.
3.  Enter a name (e.g., "Acme Corp").
4.  Enter a display name (e.g., "Acme Corporation").
5.  Click **Create**.
6.  **Expected:** Tenant appears in the list.

### 5.3 Provision a Tenant User
1.  Click on the newly created tenant.
2.  Navigate to **Users**.
3.  Click **Provision User**.
4.  Enter email (e.g., `user@acme.com`) and password.
5.  Select role: `tenant_admin`.
6.  Click **Create**.
7.  **Expected:** User appears in the list.

### 5.4 Register an OAuth Client
1.  Log out and log in as the new tenant user (`user@acme.com`).
2.  Navigate to **OAuth Clients** > **Register Client**.
3.  Enter:
    *   Name: "My App"
    *   Type: Web Application
    *   Redirect URI: `https://example.com/callback`
4.  Click **Next** through the wizard.
5.  **Expected:** `client_id` and `client_secret` are displayed.

> [!CAUTION]
> **Copy the `client_secret` immediately. It will not be shown again.**

### 5.5 Verify Audit Logs
1.  Navigate to **Audit Logs**.
2.  **Expected:** Entries for `client_created`, `login_success` are visible.

### 5.6 Initiate OIDC Flow (Manual)

> [!WARNING]
> **Not Available in Beta**
> Interactive OIDC authentication flow is **blocked** because the backend does not have a login page on the auth plane (`/oauth2/authorize` returns 401 JSON).
>
> **Enabled In:** Future release after Auth Plane UI implementation.

---

## 6. Cleanup

```bash
# Stop services (Ctrl+C in each terminal)

# Remove database (if using Docker)
cd opentrusty
docker-compose down -v
```

---

## 7. Troubleshooting

| Issue | Resolution |
|-------|------------|
| `EADDRINUSE` on port 5173 | Kill existing Vite process or change port in `vite.config.ts` |
| `connection refused` to DB | Ensure PostgreSQL is running and `DATABASE_URL` is correct |
| Login fails with 401 | Verify `BOOTSTRAP_EMAIL`/`BOOTSTRAP_PASSWORD` in `.env` |
