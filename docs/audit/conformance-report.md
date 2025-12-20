# Architecture Conformance Report

**Date**: 2025-12-17
**Scope**: Full Codebase Audit against `docs/architecture-rules.md`

## Executive Summary
The codebase is largely compliant with the newly established architectural rules, particularly in areas of Security, Persistence, and Protocol Compliance. **B-IDENTITY-01** (Identity-Credential Coupling) has been **FIXED**. Remaining critical risk exists regarding Tenant Context propagation.

**Overall Status**: 🟡 **PARTIALLY COMPLIANT**

---

## 1. Identity
| Rule | Status | Notes |
| :--- | :--- | :--- |
| **1.1 Persistence** | 🟢 Satisfied | Identity creation is decoupled from credential management (B-IDENTITY-01). **Identity provisioning is now credential-agnostic. Credential attachment is explicit and optional.** |
| **1.2 Decoupling** | 🟢 Satisfied | `User` struct has no dependency on `OAuthClient`. |
| **1.3 Uniqueness** | 🟢 Satisfied | Enforced via `UNIQUE(tenant_id, email)` index. |
| **1.4 Lifecycle** | 🟢 Satisfied | User deletion is independent of Client deletion. |

### Issues
*No active blockers in Identity domain.*

## 2. Tenant
| Rule | Status | Notes |
| :--- | :--- | :--- |
| **2.1 Isolation** | 🟡 Partial | Repository layer enforces isolation `WHERE tenant_id = $1`. However, Transport layer defaults to `"default"` tenant if header is missing. |
| **2.2 Enforcement** | 🟢 Satisfied | `tenant_id` is a mandatory column in core tables. |
| **2.3 No Cross-Talk** | 🟢 Satisfied | Lookups are scoped by Tenant ID. |

### Issues
- **[RISK]** `internal/transport/http/handlers.go:464`: `getTenantID` helper returns `"default"` if no context is found. This implicitly enables "Single Tenant" mode and risks leaking data if a request is misrouted.
  - **Remediation**: Make `TenantID` mandatory or fail request if missing (except for specific non-tenant endpoints).

## 3. Protocol (OAuth2 / OIDC)
| Rule | Status | Notes |
| :--- | :--- | :--- |
| **3.1 Standardization** | 🟢 Satisfied | Implementation uses standard `authorize` / `token` endpoints. |
| **3.2 Compliance** | 🟢 Satisfied | `redirect_uri` exact matching is enforced. |
| **3.3 Client Secrets** | 🟢 Satisfied | Hashed in storage. |
| **3.4 Token Scope** | 🟢 Satisfied | Access tokens persist `scope`. |

## 4. Authorization
| Rule | Status | Notes |
| :--- | :--- | :--- |
| **4.1 Source of Truth** | 🟢 Satisfied | IdP generates tokens with claims. |
| **4.2 Enforcement** | 🟢 Satisfied | IdP does not contain business logic enforcement. |
| **4.3 Reusability** | 🟢 Satisfied | Roles are scoped to Tenant/Project, not Client. |

## 5. Security
| Rule | Status | Notes |
| :--- | :--- | :--- |
| **5.1 Session Mgmt** | 🟢 Satisfied | Sessions are DB-backed (`sessions` table). No JWTs for sessions. |
| **5.2 Cryptography** | 🟢 Satisfied | Argon2id is used. |
| **5.3 Cookies** | 🟢 Satisfied | Configurable Secure/HttpOnly flags. |
| **5.4 Secrets** | 🟢 Satisfied | Private keys encrypted at rest. |

## 6. Observability
| Rule | Status | Notes |
| :--- | :--- | :--- |
| **6.1 Structured** | 🟢 Satisfied | Uses `slog`. |
| **6.2 Audit Trails** | 🟢 Satisfied | Critical events logged via `audit.Logger`. |
| **6.3 Context** | 🟢 Satisfied | Includes Actor, Tenant, IP. |
| **6.4 Privacy** | 🟢 Satisfied | No clear-text secrets found in logs. |

## 7. Development Practices
| Rule | Status | Notes |
| :--- | :--- | :--- |
| **7.1 Clean Arch** | 🟢 Satisfied | Handlers -> Service -> Repo separation is clear. |
| **7.2 Dependencies** | 🟢 Satisfied | Minimal external deps. |

---

## Action Plan

1. **Fix Tenant Context**: Remove the default `"default"` value in `getTenantID`. Require explicit header or domain mapping.
2. [x] **Refactor Identity Creation**: Split `CreateUser` into `ProvisionIdentity` and `AddPassword` to decouple the password requirement. (FIXED - B-IDENTITY-01)
