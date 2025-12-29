# Beta Deployment Guide (Single-Node)

This guide describes how to deploy OpenTrusty Beta on a single machine for testing and evaluation purposes.

> [!WARNING]
> **Not Production Ready**
> This deployment is for testing only. It lacks:
> - High Availability
> - Hardened TLS configuration
> - Rate limiting
> - DDoS protection

---

## 1. Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                      Single Node                             │
│                                                              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐   │
│  │   Nginx/     │    │  OpenTrusty  │    │  Control     │   │
│  │   Caddy      │───▶│   Backend    │    │  Panel UI    │   │
│  │  (Reverse    │    │  :8080       │    │  :5173       │   │
│  │   Proxy)     │───▶└──────────────┘    └──────────────┘   │
│  │   :443       │                              ▲             │
│  └──────────────┘──────────────────────────────┘             │
│         │                                                    │
│         ▼                                                    │
│  ┌──────────────┐                                            │
│  │  PostgreSQL  │                                            │
│  │   :5432      │                                            │
│  └──────────────┘                                            │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. Port Mapping

| Service | Internal Port | External Domain (Example) |
|---------|---------------|---------------------------|
| Backend API | 8080 | `api.opentrusty.local` |
| Control Panel | 5173 | `console.opentrusty.local` |
| PostgreSQL | 5432 | (internal only) |

---

## 3. Reverse Proxy Configuration

### Option A: Nginx

```nginx
# /etc/nginx/sites-available/opentrusty

# API Backend
server {
    listen 443 ssl http2;
    server_name api.opentrusty.local;

    ssl_certificate /etc/ssl/certs/opentrusty.crt;
    ssl_certificate_key /etc/ssl/private/opentrusty.key;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# Control Panel UI
server {
    listen 443 ssl http2;
    server_name console.opentrusty.local;

    ssl_certificate /etc/ssl/certs/opentrusty.crt;
    ssl_certificate_key /etc/ssl/private/opentrusty.key;

    location / {
        proxy_pass http://127.0.0.1:5173;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### Option B: Caddy

```caddyfile
# /etc/caddy/Caddyfile

api.opentrusty.local {
    reverse_proxy localhost:8080
}

console.opentrusty.local {
    reverse_proxy localhost:5173
}
```

---

## 4. TLS Expectations

For Beta:
- Self-signed certificates are acceptable for local testing.
- For public-facing deployments, use Let's Encrypt via Caddy or Certbot.

> [!CAUTION]
> **Security Warning**
> Do not run without TLS in any environment where real credentials will be used.

---

## 5. Systemd Service Files (Optional)

### Backend Service

```ini
# /etc/systemd/system/opentrusty.service
[Unit]
Description=OpenTrusty Backend
After=network.target postgresql.service

[Service]
Type=simple
User=opentrusty
WorkingDirectory=/opt/opentrusty
ExecStart=/opt/opentrusty/bin/opentrusty
Restart=on-failure
EnvironmentFile=/opt/opentrusty/.env

[Install]
WantedBy=multi-user.target
```

### UI Service (Production Build)

For production, build the UI and serve with a static server:

```bash
cd opentrusty-control-panel
npm run build
# Serve dist/ with Nginx or Caddy
```

---

## 6. Environment Variables Reference

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `BOOTSTRAP_EMAIL` | Yes (first run) | Initial platform admin email |
| `BOOTSTRAP_PASSWORD` | Yes (first run) | Initial platform admin password |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | OpenTelemetry endpoint |

---

## 7. Verification

After deployment, verify:

1. `curl https://api.opentrusty.local/api/v1/health` returns 200
2. `https://console.opentrusty.local/admin/login` renders the login page
3. Login with bootstrap credentials succeeds
