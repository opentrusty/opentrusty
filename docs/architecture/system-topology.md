# System Topology

This document defines the deployment topology and trust domains for OpenTrusty.

## Trust Domains

OpenTrusty operates across three distinct trust domains, each with different security characteristics.

```mermaid
graph TB
    subgraph "External Clients"
        Browser["Browser (End User)"]
        AdminBrowser["Browser (Admin)"]
        App["Application (RP)"]
    end

    subgraph "Console Domain (console.*)"
        direction TB
        UI["Control Panel UI<br/>(Static SPA)"]
    end

    subgraph "Core Binary"
        subgraph "Auth Domain (auth.*)"
            direction TB
            AuthEndpoints["OIDC/OAuth2 Endpoints"]
            LoginPages["Login/Consent Pages<br/>(Server-Rendered)"]
        end

        subgraph "API Domain (api.*)"
            direction TB
            AdminAPI["Management API"]
        end

        subgraph "Shared Domain Core"
            direction TB
            Identity["Identity Service"]
            Tenant["Tenant Service"]
            Session["Session Service"]
            RBAC["Authorization Service"]
        end
    end

    subgraph "Persistence"
        DB[(PostgreSQL)]
    end

    Browser -->|"OAuth2 Flow"| AuthEndpoints
    Browser -->|"Login Form"| LoginPages
    App -->|"Token Exchange"| AuthEndpoints

    AdminBrowser -->|"Session Cookie"| UI
    UI -->|"REST API"| AdminAPI

    AuthEndpoints --> Identity
    AuthEndpoints --> Session
    LoginPages --> Identity
    LoginPages --> Session

    AdminAPI --> Identity
    AdminAPI --> Tenant
    AdminAPI --> RBAC

    Identity --> DB
    Tenant --> DB
    Session --> DB
    RBAC --> DB
```

## Domain Responsibilities

| Domain | Subdomain | Trust Level | Responsibilities |
|--------|-----------|-------------|------------------|
| **Auth** | `auth.*` | Highest | OIDC/OAuth2 protocol, credential verification, session cookies, server-rendered login pages |
| **API** | `api.*` | High | Tenant/user/client management, RBAC enforcement, admin operations |
| **Console** | `console.*` | Untrusted | Human-facing admin UI, API consumer, zero business logic |

## Security Isolation

```mermaid
flowchart LR
    subgraph "Untrusted Zone"
        Console["console.*<br/>Control Panel"]
    end

    subgraph "Trusted Zone (Core Binary)"
        API["api.*<br/>Management API"]
        Auth["auth.*<br/>Authentication"]
    end

    Console -->|"HTTP + Session Cookie"| API
    API -->|"Enforces Authorization"| Auth

    style Console fill:#ffcccc,stroke:#cc0000
    style API fill:#ccffcc,stroke:#00cc00
    style Auth fill:#ccccff,stroke:#0000cc
```

### Key Invariants

1. **Console is Untrusted**: The Control Panel UI cannot bypass API authorization
2. **Shared Session Cookies**: Browser sessions are HttpOnly cookies issued by Auth domain
3. **No Direct DB Access**: Console never touches the database directly
4. **Separate Deployment**: Console is a separate artifact, not embedded in core binary

## Browser Interaction Flows

### End-User Authentication Flow

```mermaid
sequenceDiagram
    participant U as End User
    participant A as auth.*
    participant App as Application

    U->>A: GET /oauth2/authorize
    A->>U: Redirect to Login Page
    U->>A: POST /auth/login (credentials)
    A->>A: Validate, Create Session
    A->>U: Redirect with Authorization Code
    U->>App: Code Exchange
    App->>A: POST /oauth2/token
    A->>App: Access Token + ID Token
```

### Admin Console Flow

```mermaid
sequenceDiagram
    participant A as Admin Browser
    participant C as console.*
    participant API as api.*
    participant Auth as auth.*

    A->>Auth: Login via auth.*
    Auth->>A: Session Cookie (HttpOnly)
    A->>C: Load Control Panel
    C->>API: GET /api/tenants (Cookie attached)
    API->>API: Validate Session, Check RBAC
    API->>C: Tenant List
    C->>A: Render UI
```

## Deployment Topology

| Component | Artifact | Repository |
|-----------|----------|------------|
| Auth + API | Single Go binary | `opentrusty` |
| Console | Static SPA | `opentrusty-control-panel` |
| Database | PostgreSQL | External dependency |

**The core binary exposes two entrypoints:**
- `auth.*` — Authentication Plane (OIDC/OAuth2)
- `api.*` — Management API Plane (Admin REST)

**The Control Panel is deployed separately:**
- Static files served via CDN, Nginx, or similar
- Never bundled into the core binary
