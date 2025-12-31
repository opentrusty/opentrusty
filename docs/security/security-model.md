# Security Model

OpenTrusty is designed with a defense-in-depth approach to identity and access management. This document defines the core security posture and boundaries.

## 1. Multi-Tenancy Invariants

Tenants are logically isolated units of administrative and identity data.
- **Data Isolation**: All resources (Users, Clients, Sessions) are bonded to a `tenant_id`.
- **Authorization Context**: Permission checks ALWAYS require a scope context (Platform, Tenant, or Client).
- **Cross-Tenant Prevention**: A bearer token or session issued for Tenant A can NEVER be used to access resources in Tenant B.

## 2. Administrative Least-Privilege

We distinguish between "Platform Operator" and "Tenant Administrator".

- **Platform Admin**: Manages the infrastructure, tenant lifecycle, and system configuration. Does NOT have default access to tenant data.
- **Tenant Owner/Admin**: Manages the identity and policy within their specific tenant boundary.

## 3. Scoped Audit Access

To prevent unrestricted profiling of tenant activities, Platform Admin access to tenant audit logs is **controlled, scoped, and fully audited**.
- No "Platform-wide" audit view exists.
- Access requires an **Explicit Declaration** of intent (Reason, Window, Tenant).
- The declaration itself is a high-visibility audit event.

## 4. Cryptographic Standards

- **Passwords**: Argon2id (RFC 9106) with unique salt and adaptive parameters.
- **OIDC/OAuth2**: Defaulting to RS256. Enforcing PKCE for public clients.
- **Tokens**: UUID-based authorization codes (short-lived) and session IDs.

## 5. Audit Immutability

Audit logs are defined as append-only. The platform provides no mechanisms for the deletion or suppression of recorded security events.
