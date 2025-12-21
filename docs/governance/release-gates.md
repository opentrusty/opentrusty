# Release Gates

This document defines the formal Release Gates that must be satisfied before any version of OpenTrusty is published. These gates ensure that every release meets our standards for security, stability, and trustworthiness.

## 1. API Documentation Trustworthiness (Critical Gate)

API documentation is not merely a developer convenience; it is a **core security artifact**. Accurate documentation is essential for security auditing, correct implementation of protocol flows (OIDC/OAuth2), and preventing integrator error.

### Policy
- A release MUST NOT be published if its public API documentation is incomplete, semantically inaccurate, or out of sync with the code.
- Stale or unreliable documentation is considered a security vulnerability and is a hard blocker for any release.

### Enforcement Mechanism

#### Automated Checks (CI-Enforced)
- **Freshness Check**: The `scripts/check-docs.sh` script must pass in CI. This verifies that the committed `swagger.json` perfectly matches the specification generated from the current source code.
- **Spec Integrity**: The generated OpenAPI specification must be syntactically valid (OpenAPI 3.1).

#### Human-Enforced Checks (Reviewer-Enforced)
- **Semantics**: Reviewers must verify that OpenAPI annotations accurately describe the protocol behavior (e.g., correct HTTP status codes, accurate parameter descriptions).
- **Security Context**: Security schemes and sensitive examples must be reviewed to ensure they align with the project's security assumptions (e.g., no hardcoded secret examples).
- **PR Checklist**: The mandatory documentation item in `docs/contributing/pr-checklist.md` must be checked for every pull request that modifies the API surface.

## 2. Decision Logic
- **Fail-Fast**: If any release gate check fails, the release process is automatically halted. Any change that causes UT, Docker E2E, or systemd smoke tests to fail MUST NOT be released.
- **No Bypass**: Bypassing a release gate requires a formal proposal and consensus from at least two maintainers, documented in a public issue.
- **No Partial Bypass**: Partial bypassing of release gates is not allowed. If a release gate fails, the entire release process must be halted.

## 3. Governance
These gates are governed by the maintainers of OpenTrusty. Any modifications to the release gates must be proposed and approved according to the [GOVERNANCE.md](../governance/GOVERNANCE.md) guidelines.
