# OpenTrusty PR Review Checklist

Use this checklist to review Pull Requests. If any answer is "NO", the PR should be rejected or requested for changes.

## 1. Identity
- [ ] **1.1** Does this change treat Identity as a persistent, unique entity separate from credentials?
- [ ] **1.2** Is the Identity decoupled from any specific Application (`OAuthClient`)?
- [ ] **1.3** Are Identity identifiers unique within the Tenant scope?
- [ ] **1.4** Is the Identity lifecycle independent of Application lifecycle (e.g. deleting an app doesn't delete users)?

## 2. Tenant
- [ ] **2.1** Is the Tenant boundary respected (all data scoped to Tenant)?
- [ ] **2.2** Is multi-tenancy enforced at the persistence layer (e.g. `tenant_id` column)?
- [ ] **2.3** Is cross-tenant data leakage strictly prevented?

## 3. Protocol (OAuth2 / OIDC)
- [ ] **3.1** Does the authentication flow use standard OAuth2/OIDC mechanisms (no custom login APIs)?
- [ ] **3.2** Does the implementation strictly adhere to RFC specs (e.g. exact `redirect_uri` match)?
- [ ] **3.3** Are `client_secrets` handled securely (hashed storage, not passed in URL)?
- [ ] **3.4** Are Access Tokens properly scoped to Tenant and Permissions?

## 4. Authorization
- [ ] **4.1** Is OpenTrusty acting as the Source of Truth for role assignments?
- [ ] **4.2** Is the enforcement logic (Is this Role allowed?) delegated to the Resource Server?
- [ ] **4.3** Are roles reusable across multiple applications (not hard-coded to one client)?

## 5. Security
- [ ] **5.1** Are primary user sessions stateful (DB-backed) and NOT stateless JWTs?
- [ ] **5.2** Is password hashing done using **Argon2id** (no weak algorithms)?
- [ ] **5.3** Are authentication cookies marked `HttpOnly`, `Secure`, and `SameSite=Lax`?
- [ ] **5.4** Are all secrets encrypted at rest or hashed, and never committed to git?

## 6. Observability
- [ ] **6.1** Are all logs structured (JSON)?
- [ ] **6.2** Are security-critical events (Login, Token, Roles) emitted to the Audit Log?
- [ ] **6.3** Do audit logs include Context (Who, Where, When, What)?
- [ ] **6.4** Is PII and secret data redacted from logs?

## 7. Development Practices
- [ ] **7.1** Is business logic isolated in the Domain layer (not in HTTP handlers)?
- [ ] **7.2** Are external dependencies justified and minimal?
