# Integration Contract

This document defines the integration responsibilities between OpenTrusty components.

## Component Interaction Matrix

```
┌─────────────────────────────────────────────────────────┐
│                     Browser                             │
└────────────┬──────────────────────┬────────────────────┘
             │                      │
             ▼                      ▼
┌────────────────────┐   ┌────────────────────┐
│  console.*         │   │  auth.*            │
│  Control Panel     │   │  Authentication    │
│  (SPA)             │   │  (Core Binary)     │
└─────────┬──────────┘   └──────────┬─────────┘
          │                         │
          │ REST API                │ Session
          ▼                         │ Validation
┌────────────────────┐              │
│  api.*             │◄─────────────┘
│  Management API    │
│  (Core Binary)     │
└────────────────────┘
```

## Responsibility Boundaries

### Auth Plane (`auth.*`) Responsibilities
- Accept user credentials
- Issue session cookies
- Render login/consent pages
- Issue OAuth2/OIDC tokens
- Validate tokens for resource servers

### API Plane (`api.*`) Responsibilities
- Validate session cookies from auth
- Enforce RBAC authorization
- Execute admin operations (CRUD)
- Return JSON responses
- Never render HTML

### Console Plane (`console.*`) Responsibilities
- Render admin UI
- Call Management API
- Handle 401/403 gracefully
- Never store secrets
- Never bypass API authorization

## Explicit Forbidden Cross-Overs

| Cross-Over | Status | Rationale |
|------------|--------|-----------|
| Console → Auth endpoints directly | ❌ FORBIDDEN | Console uses session cookies, not OAuth flows |
| Auth → Admin UX rendering | ❌ FORBIDDEN | Auth only renders login pages |
| API → HTML rendering | ❌ FORBIDDEN | API is JSON-only |
| Console → Direct DB access | ❌ FORBIDDEN | All data via API |
| Auth → Console assets | ❌ FORBIDDEN | No SPA in core binary |

## Session Flow

1. **Admin authenticates** via `auth.*` login page (Port 8080)
2. **Auth issues** HttpOnly session cookie (Scoped to `.opentrusty.org`)
3. **Browser loads** Control Panel from `console.*`
4. **Console calls** `api.*` (Port 8081) with cookie attached
5. **Load Balancer / Proxy** routes requests based on path:
    - `/api/v1/auth/*` -> Auth Plane (Port 8080)
    - `/oauth2/*` -> Auth Plane (Port 8080)
    - `/.well-known/*` -> Auth Plane (Port 8080)
    - `/api/v1/*` (Management) -> Admin Plane (Port 8081)
6. **API validates** session via shared session store
7. **API enforces** RBAC and returns data

## Contract Violations

Any of the following constitutes a contract violation:

1. Core binary serving SPA assets
2. Console implementing auth logic
3. API returning HTML
4. Auth exposing admin CRUD endpoints
5. Console storing tokens in browser storage
