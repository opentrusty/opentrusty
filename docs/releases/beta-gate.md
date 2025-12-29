# Beta Release Gate

Explicit criteria for OpenTrusty Beta classification.

## Beta Entry Criteria

All conditions below MUST be satisfied before declaring Beta status.

### 1. Local System Validation

| Layer | Requirement | Verification |
|-------|-------------|--------------|
| 1 | All services start and respond | `infra-smoke.md` |
| 2 | Platform admin auto-provisioned | `bootstrap-test.md` |
| 3 | Tenant lifecycle functional | `tenant-lifecycle.md` |
| 4 | OAuth2 client management works | `oauth-client-test.md` |
| 5 | Complete OIDC flow succeeds | `oidc-e2e.md` |
| 6 | Audit logging verified | `audit-verification.md` |
| 7 | Security boundaries enforced | `isolation-negative.md` |

**Command:** `./tests/local/run-all.sh`

**Pass Criteria:** All 7 layers report PASS

### 2. Security Requirements

- [ ] No JWT for primary user sessions
- [ ] Session cookies: HttpOnly, Secure, SameSite=Lax
- [ ] client_secret never in URLs or logs
- [ ] Passwords hashed with Argon2id
- [ ] Cross-tenant access blocked

### 3. Protocol Compliance

- [ ] OIDC Discovery endpoint functional
- [ ] JWKS endpoint returns valid keys
- [ ] Authorization codes single-use
- [ ] redirect_uri exact match enforced
- [ ] PKCE (S256) supported

### 4. Documentation

- [ ] `docs/testing/local-system-test.md` exists
- [ ] Test reports generated in `tests/local/reports/`
- [ ] Known limitations documented

## Beta Classification

```
┌─────────────────────────────────────────────────────────┐
│                                                         │
│   Beta = ALL automated tests pass                       │
│        + Security requirements verified                 │
│        + Can onboard tenant and serve real users        │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## What Beta Means

- **Suitable for:** Controlled testing with real users
- **NOT suitable for:** Production deployment
- **Required before:** External pilot programs
- **Missing for GA:** TLS, production config, external audit

## Verification Command

```bash
cd /path/to/opentrusty
./tests/local/run-all.sh

# Expected output:
# ╔══════════════════════════════════════════════════════════════╗
# ║               🎉 ALL TESTS PASSED - BETA READY 🎉            ║
# ╚══════════════════════════════════════════════════════════════╝
```

---

*Status: Beta Gate Definition*
