# Domain Routing

This document defines the routing rules for OpenTrusty's domains.

## Standard Domain Model

OpenTrusty relies on **Host-Based Routing** to separate concerns.

| Domain | Service | Function |
|--------|---------|----------|
| `auth.example.com` | Auth Service | User Login, OIDC, OAuth2 |
| `api.example.com` | API Service | Management API (JSON) |
| `console.example.com` | Control Panel | Static Admin UI (SPA) |

## Reverse Proxy Configuration

A reverse proxy (Nginx, Caddy, AWS ALB, Cloudflare) is REQUIRED to handle TLS and routing.

### Nginx Example

```nginx
# Auth Service
server {
    listen 443 ssl http2;
    server_name auth.example.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

# API Service
server {
    listen 443 ssl http2;
    server_name api.example.com;

    location / {
        proxy_pass http://127.0.0.1:8081;
        proxy_set_header Host $host;
    }
}
```

## Cookie Scoping

The **Session Cookie** is critical for the Admin Console flow.

- **Issuer**: `auth.example.com`
- **Domain Attribute**: `.example.com` (Wildcard) OR `api.example.com` (if explicitly scoped)

For the Control Panel (`console.example.com`) to validly call the API (`api.example.com`):

1.  User logs in at `auth.example.com`.
2.  Auth issues `session_id` cookie scoped to `.example.com` (Root Domain).
3.  Browser sends this cookie to ANY subdomain of `example.com`, including `api.example.com`.
4.  API validates the cookie.

> **Constraint**: `auth`, `api`, and `console` MUST share a common parent domain if using Root Domain cookies.

## CORS Architecture

The `api.example.com` service must allow CORS from `console.example.com`.

- **Access-Control-Allow-Origin**: `https://console.example.com`
- **Access-Control-Allow-Credentials**: `true` (Essential for cookies)
