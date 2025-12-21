# OpenTrusty Operations Manual

OpenTrusty follows a **binary-first philosophy**. The primary distribution artifact is a single, statically-linked Linux binary with minimal external dependencies besides a Postgres database.

## 1. Binary Deployment (Standard)

Binary deployment is the recommended way to run OpenTrusty for maximum performance and minimum overhead.

### Building
```bash
make build
```
The binary will be located in `./bin/opentrusty`.

### Running
```bash
./bin/opentrusty
```
OpenTrusty requires environment variables for configuration (see Section 3).

## End-to-End Testing
OpenTrusty includes a suite of E2E tests that simulate real-world identity workflows against a live environment.

### Prerequisites
- Docker and Docker Compose installed.
- Environment is running: `make dev-up`.

### Running Tests
To run the E2E tests, execute:
```bash
make test-e2e
```

The tests cover:
1. **Platform Admin**: Registration, tenant creation, and OAuth2 client registration.
2. **Tenant Admin**: User provisioning and management.
3. **End User OIDC Flow**: Full authorization code flow with ID Token validation.

---

## 2. Docker Deployment (Optional)

Docker is provided for containerized environments, testing, and simplified dependency management.

### Building the Image
```bash
docker build -t opentrusty:latest -f deploy/docker/Dockerfile .
```

### Running with Docker
```bash
docker run -p 8080:8080 \
  -e DB_URL=postgres://user:pass@host:5432/db \
  opentrusty:latest
```

### Local Development & E2E Testing (Compose)
For developers, a Docker Compose setup is provided to spin up OpenTrusty along with a PostgreSQL instance.

**Important**: This setup is intended ONLY for local development and testing.

```bash
# Start the environment
docker compose -f deploy/docker/docker-compose.yml up -d

# Stop and clean up (remove volumes)
docker compose -f deploy/docker/docker-compose.yml down -v
```

### Systemd Deployment (Production)
For bare-metal or VM deployments, OpenTrusty is managed via a `systemd` service.

Detailed instructions for binary installation, user creation, and service management are located in the specialized guide:
👉 **[Systemd Deployment Guide](https://github.com/opentrusty/opentrusty/blob/main/deploy/systemd/README.md)**

---

## 3. Configuration

OpenTrusty is configured exclusively via environment variables. There are no default secrets.

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_URL` | Postgres connection string | `postgres://opentrusty:password@localhost:5432/opentrusty` |
| `PORT` | Listening port (default: 8080) | `8080` |
| `ISSUER` | OIDC Issuer URL | `https://auth.example.com` |

---

## 4. Maintenance

### Database Migrations
OpenTrusty automatically applies migrations on startup if they are available in the expected directory.

### Upgrading
For binary deployments, simply replace the binary and restart the service. OpenTrusty is designed for seamless schema upgrades.
