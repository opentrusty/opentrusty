# Audit Log Security Model

**Last Updated:** 2025-12-29

This document defines the security model for audit log visibility within OpenTrusty.

## Purpose

Audit logs provide accountability for all security-sensitive actions within the platform.
This document specifies who can access audit logs and under what conditions.

---

## Visibility Matrix

| Viewer | Log Access | Scope | Notes |
|--------|-----------|-------|-------|
| **Platform Admin** | ✅ All tenants | Read-only | Access is itself audited |
| **Tenant Owner** | ✅ Own tenant | Read-only | Full tenant visibility |
| **Tenant Admin** | ✅ Own tenant | Read-only | Operational visibility |
| **Tenant Member** | ❌ None | N/A | No audit access |

---

## Platform Admin Audit Access

### Rationale

Platform Admin has **intentional** read-only access to all tenant audit logs for:
- Compliance monitoring
- Security incident investigation
- Operational oversight
- Regulatory requirements

### Security Controls

1. **Read-Only Access**
   - Platform Admin CANNOT mutate audit logs
   - Platform Admin CANNOT delete audit entries
   - No write operations are exposed via API

2. **Access Auditing**
   - All Platform Admin access to cross-tenant audit logs generates an audit event
   - Audit event type: `audit_access`
   - Includes: actor ID, action, timestamp

3. **Data Redaction**
   - Sensitive fields (passwords, tokens, secrets) are ALWAYS redacted
   - Redaction is applied uniformly regardless of viewer role
   - Redaction keywords: `password`, `secret`, `token`, `key`, `authorization`, `hash`, `credential`, `private`, `api_key`

### Does NOT Break Tenant Isolation

Platform Admin audit access is a **read-only observer** capability:
- DOES NOT grant write access to tenant data
- DOES NOT allow Platform Admin to modify tenant users
- DOES NOT allow Platform Admin to access tenant secrets
- DOES NOT create an implicit tenant membership

The data mutation boundary remains strictly enforced.

---

## API Endpoints

### Platform Audit Logs

```
GET /api/v1/audit
```

**Parameters:**
- `all_tenants=true` - Include logs from all tenants (Platform Admin only)

**Authorization:** Requires `platform:view_audit` permission

### Tenant Audit Logs

```
GET /api/v1/tenants/{tenantID}/audit
```

**Authorization:** Requires:
- `tenant:view_audit` for the specific tenant, OR
- `platform:view_audit` (Platform Admin)

---

## Audit Event Types

| Event Type | Description | Logged When |
|------------|-------------|-------------|
| `login_success` | Successful authentication | User logs in |
| `login_failed` | Failed authentication | Login attempt fails |
| `user_created` | New user provisioned | User is created |
| `password_changed` | Password updated | User changes password |
| `client_created` | OAuth2 client registered | Client is registered |
| `secret_rotated` | Client secret rotated | Secret is refreshed |
| `tenant_created` | New tenant provisioned | Tenant is created |
| `audit_access` | Audit logs accessed | Platform Admin views all logs |

---

## Invariants

1. Audit logs are immutable once written
2. No user can delete audit entries
3. Sensitive data is redacted before storage
4. All security-relevant actions generate audit events
5. Cross-tenant audit access is restricted to Platform Admin
