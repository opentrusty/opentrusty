# OpenTrusty

<p align="center">
  <img src="docs/assets/logo-full.svg" alt="OpenTrusty - Secure • Open • Identity" width="400"/>
  </br>
  [![CI](https://github.com/opentrusty/opentrusty/actions/workflows/ci.yml/badge.svg)](https://github.com/opentrusty/opentrusty/actions/workflows/ci.yml) | [![Publish API Docs](https://github.com/opentrusty/opentrusty/actions/workflows/docs.yml/badge.svg)](https://github.com/opentrusty/opentrusty/actions/workflows/docs.yml) | [![Release Gate](https://github.com/opentrusty/opentrusty/actions/workflows/release.yml/badge.svg)](https://github.com/opentrusty/opentrusty/actions/workflows/release.yml)
</p>

**OpenTrusty** is a security-first, open-source Identity Provider (IdP) implementing OAuth 2.0 and OpenID Connect (OIDC), designed for auditability, correctness, and long-term maintainability.

OpenTrusty is developed as a **non-profit, community-oriented project** under the Apache License 2.0, with an emphasis on clear security boundaries, explicit assumptions, and responsible open source governance.

---

## Project Status

> ⚠️ **Current Status: Pre-Production / Foundation Stage**

- Core OAuth2 and OIDC protocols are implemented and security-hardened
- Architecture and threat model are complete
- Automated test coverage and deployment automation are in progress

**OpenTrusty is not yet recommended for production use without review.**

---

## Key Capabilities

### OAuth 2.0
- Authorization Code Flow
- PKCE (S256 and plain)
- Refresh Tokens
- Token Revocation (RFC 7009)

### OpenID Connect
- id_token (RS256)
- Discovery (`/.well-known/openid-configuration`)
- JWKS (`/jwks.json`)
- nonce and at_hash hardening

### Multi-Tenancy
- Strict tenant isolation
- No implicit or default tenant fallback

### Security Principles
- Argon2id password hashing (ONLY approved algorithm)
- Database-backed sessions (no JWT for primary sessions)
- HttpOnly, Secure, SameSite cookies
- Prepared statements for all database access
- Explicit threat model and security assumptions

> See `docs/fundamentals/capabilities.md` for a full feature matrix.

---

## Architecture Overview

OpenTrusty follows strict architectural rules:

- **Identity ≠ Authentication**: Identity lifecycle is decoupled from credentials
- **Protocol Isolation**: OAuth2 and OIDC logic are isolated from transport and core domains
- **Fail-Closed by Default**: No implicit defaults for tenant or security context
- **Auditability First**: Security-relevant events are logged explicitly

Detailed documentation:
- `docs/architecture/architecture-rules.md`
- `docs/security/threat-model.md`
- `docs/security/security-assumptions.md`

---

## Getting Started (Development)

### Prerequisites

- Go 1.25+
- Docker & Docker Compose
- Make (recommended)

### Development Setup

```bash
git clone https://github.com/opentrusty/opentrusty.git
cd opentrusty
make dev
make run
````

This will:

* Start PostgreSQL via Docker
* Apply database migrations
* Run the OpenTrusty server locally

---

## Documentation

* **Architecture & Protocols**: `docs/architecture/`
* **Security & Deployment**: `docs/security/`
* **Audit & Conformance**: `docs/audit/`
* **Governance**: `docs/governance/`

Start with: `docs/README.md`

---

## Contributing

OpenTrusty welcomes contributors who value correctness, security, and clarity.

Before contributing:

* Read `docs/architecture/architecture-rules.md`
* Review `SECURITY.md` for vulnerability disclosure
* Follow the project's coding and review standards

---

## License

OpenTrusty is licensed under the Apache License, Version 2.0.

See the [LICENSE](./LICENSE) file for details.

---

## Disclaimer

OpenTrusty is provided "as is" without warranty.
Users are responsible for reviewing and validating the software for their specific security and compliance requirements.