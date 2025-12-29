# Local Full-System Beta Topology

This document describes the required local environment for validating OpenTrusty as a Beta-grade system.

## 1. Local Topology

All services are expected to run on `localhost`. For production-like path validation, you may optionally map these to `/etc/hosts`.

| Service | Port | Local URL | Role |
| :--- | :--- | :--- | :--- |
| **PostgreSQL** | 5433 | `localhost:5433` | Persistence Layer |
| **OpenTrusty Core** | 8080 | `localhost:8080` | Auth Plane & Management API Plane |
| **Control Panel** | 5173 | `localhost:5173` | Administrative UI (Vite Dev Server) |
| **Demo Client** | 8081 | `localhost:8081` | External OIDC Demo App |

## 2. Environment Variables

### 2.1. OpenTrusty Core (`opentrusty` repo)
Set these in `.env` or export them before running:
```bash
SERVER_HOST=localhost
SERVER_PORT=8080
DB_HOST=localhost
DB_PORT=5433
DB_USER=opentrusty
DB_PASSWORD=opentrusty_dev_password
DB_NAME=opentrusty
OPENID_KEY_ENCRYPTION_KEY=12345678901234567890123456789012 # 32 bytes
```

### 2.2. Demo Client (`opentrusty-demo-app` repo)
```bash
PORT=8081
CLIENT_ID=<generated_id>
CLIENT_SECRET=<generated_secret>
REDIRECT_URI=http://localhost:8081/callback
AUTH_URL=http://localhost:8080
```

## 3. Startup Order

Services MUST be started in this specific order to ensure dependency readiness:

1.  **PostgreSQL**: `make dev` (starts docker-compose)
2.  **Database Migration**: `go run cmd/server/main.go migrate`
3.  **OpenTrusty Core**: `go run cmd/server/main.go serve all`
4.  **Control Panel**: `npm run dev` (in `opentrusty-control-panel` directory)
5.  **Demo Client**: `go run main.go` (in `opentrusty-demo-app` directory)

---
*Status: Beta Validation Baseline*
