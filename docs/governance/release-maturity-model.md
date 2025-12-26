# Release Maturity Model

**Status**: Normative  
**Owner**: OpenTrusty Maintainers  
**Last Updated**: 2025-12-22

This document defines the formal Release Maturity Model for OpenTrusty. All releases MUST conform to the requirements specified for their declared maturity level.

---

## 1. Maturity Level Definitions

### 1.1 Alpha

**Purpose**: Experimental releases for early feedback and exploration of new features.

**Audience**: Developers, early adopters, and contributors willing to accept breaking changes.

**Stability Guarantees**:
- **Breaking Changes**: Allowed between alpha releases
- **API Stability**: Not guaranteed
- **Schema Stability**: Not guaranteed (migrations may be non-reversible)
- **Data Compatibility**: Not guaranteed between alpha versions

**Production Use**: **NOT RECOMMENDED**

**Exploratory Status**: Alpha releases are exploratory and for development purposes only. They are not production commitments and do not require a formal release checklist.

### 1.2 Beta

**Purpose**: Feature-complete releases undergoing stabilization and community validation.

**Audience**: Early adopters, integration partners, and pilot deployments in non-critical environments.

**Stability Guarantees**:
- **Breaking Changes**: Discouraged but may occur with documented migration paths
- **API Stability**: API surface is frozen; only additive changes allowed
- **Schema Stability**: Database migrations must be reversible
- **Data Compatibility**: Upgrade path from previous beta versions must exist

**Production Use**: **USE WITH CAUTION** (Non-critical workloads only)

**Exploratory Status**: Beta releases are exploratory and intended for stabilizing features. While more stable than Alpha, they are not production commitments and do not require a formal release checklist.

### 1.3 Release Candidate (RC)

**Purpose**: Production-ready candidates undergoing final validation before GA. RCs are **Candidate Releases** and must be explicitly promoted to GA.

**Audience**: Production pilots, integration partners, and security auditors.

**Stability Guarantees**:
- **Breaking Changes**: Prohibited without exceptional justification and maintainer consensus
- **API Stability**: Fully frozen; only critical bug fixes allowed
- **Schema Stability**: Database migrations must be fully reversible
- **Data Compatibility**: Full upgrade compatibility with previous RC and the upcoming GA

**Production Use**: **ACCEPTABLE** (With contingency planning)

**Checklist Requirement**: RC releases are production candidates and **MUST** have a verified and completed release checklist validated by CI.

### 1.4 General Availability (GA)

**Purpose**: Production-grade releases with long-term support guarantees. GAs are **Promoted Releases**, created only after a successful RC validation period.

**Audience**: Production deployments, enterprise users, and security-conscious organizations.

**Stability Guarantees**:
- **Breaking Changes**: Prohibited within the same major version
- **API Stability**: Guaranteed; follows semantic versioning strictly
- **Schema Stability**: All migrations reversible; downgrade paths documented
- **Data Compatibility**: Full backward and forward compatibility within major version

**Production Use**: **RECOMMENDED**

**Checklist Requirement**: GA releases are production-grade and **MUST** have a fully verified release checklist inherited from a successful RC promotion.
**Immutability**: Once a version is promoted to GA (e.g., `v1.0.0`), it is immutable. Any fixes must be released as a new patch version (e.g., `v1.0.1`).

---

## 2. Required Test Gates Per Maturity Level

### 2.1 Alpha Requirements (Minimal Gates)

| Gate | Status | Enforcement |
|------|--------|-------------|
| Unit Tests | MUST pass | CI-enforced |
| API Documentation Freshness | MUST pass | CI-enforced |
| Integration Tests | SHOULD pass | CI-warning only |
| E2E Tests (Docker) | MAY fail | Manual review |
| **Release Checklist** | **NOT REQUIRED** | **Skipped in CI** |

**Failure Tolerance**: Alpha releases MAY be published with known test failures if documented in release notes.

### 2.2 Beta Requirements (Minimal Gates)

| Gate | Status | Enforcement |
|------|--------|-------------|
| Unit Tests | MUST pass | CI-enforced |
| Integration Tests | MUST pass | CI-enforced |
| E2E Tests (Docker) | MUST pass | CI-enforced |
| API Documentation Freshness | MUST pass | CI-enforced |
| Systemd Smoke Test | SHOULD pass | CI-warning |
| **Release Checklist** | **NOT REQUIRED** | **Skipped in CI** |

**Failure Tolerance**: Beta releases MUST NOT be published with critical test failures. Non-critical failures require explicit acknowledgment in release notes.

**Architectural Invariant**:
- **No Embedded UI**: From Beta onwards, the auth binary MUST NOT embed the Admin UI.
- **Separate Processes**: Deployment instructions MUST document multi-process setup.

### 2.3 RC Requirements (Full Gates + Checklist)

| Gate | Status | Enforcement |
|------|--------|-------------|
| Unit Tests | MUST pass | CI-enforced |
| Integration Tests | MUST pass | CI-enforced |
| E2E Tests (Docker) | MUST pass | CI-enforced |
| Systemd Smoke Test | MUST pass | CI-enforced |
| API Documentation Freshness | MUST pass | CI-enforced |
| Security Scan | MUST pass | Manual review required |
| Performance Regression | SHOULD pass | Benchmark comparison |
| **Release Checklist** | **MUST be completed** | **CI-enforced (rc.* tags)** |

**Failure Tolerance**: RC releases MUST NOT be published with ANY test failures.

### 2.4 GA Requirements (Full Gates + Checklist)

| Gate | Status | Enforcement |
|------|--------|-------------|
| Unit Tests | MUST pass | CI-enforced |
| Integration Tests | MUST pass | CI-enforced |
| E2E Tests (Docker) | MUST pass | CI-enforced |
| Systemd Smoke Test | MUST pass | CI-enforced |
| API Documentation Freshness | MUST pass | CI-enforced |
| Security Scan | MUST pass | Mandatory maintainer review |
| Performance Regression | MUST pass | Automated threshold checks |
| Upgrade Path Validation | MUST pass | Manual verification from previous GA |
| **Release Checklist** | **MUST be inherited** | **Verified during promotion** |

**Failure Tolerance**: GA releases MUST NOT be published with ANY test failures or unresolved security findings.

---

## 3. Documentation Requirements

### 3.1 Alpha

- **MUST**: API documentation generated and published
- **MUST**: Known issues documented in release notes
- **SHOULD**: Migration guide if schema changes introduced

### 3.2 Beta

- **MUST**: Complete API documentation with examples
- **MUST**: Migration guide for breaking changes
- **MUST**: Known limitations and workarounds documented
- **SHOULD**: Security assumptions documented

### 3.3 RC

- **MUST**: All Beta requirements
- **MUST**: Security assumptions and threat model reviewed
- **MUST**: Deployment guide with systemd instructions
- **MUST**: Rollback procedures documented

### 3.4 GA

- **MUST**: All RC requirements
- **MUST**: Production deployment best practices
- **MUST**: Monitoring and observability guide
- **MUST**: Incident response procedures
- **MUST**: Support and maintenance policy

---

## 4. Promotion Process

Promotion is the explicit act of graduating a Release Candidate to a General Availability release.

1. **Validation Period**: An RC must undergo a validation period (Beta -> RC: >2 weeks; RC -> GA: >4 weeks) as defined in Section 6.
2. **Review**: Maintainers review the release checklist and test results.
3. **Explicit Promotion**: The release manager triggers the **Promote Release** GitHub Action, providing the `candidate_tag` and the target `release_version`. This workflow verifies previous gates, creates the GA tag (e.g., `v1.0.0`) pointing to the exact same commit, and publishes the official release.
4. **Final Gate**: The GA tag triggers the final production release gate, including full compliance and upgrade path validation.

---

## 5. Tag Naming Conventions

### 5.1 Format Rules

All release tags MUST follow this format, adhering to Semantic Versioning 2.0.0:

```
v{MAJOR}.{MINOR}.{PATCH}[-{MATURITY}.{NUMBER}]
```

**Examples**:
- Alpha: `v0.1.0-alpha.1`, `v0.1.0-alpha.2`
- Beta: `v0.2.0-beta.1`, `v0.2.0-beta.2`
- RC: `v1.0.0-rc.1`, `v1.0.0-rc.2`
- GA: `v1.0.0`, `v1.0.1`, `v1.1.0`

### 5.2 Semantic Versioning

- **MAJOR**: Breaking changes to public API or data model
- **MINOR**: New features; backward-compatible additions
- **PATCH**: Bug fixes; no new features

### 5.3 Pre-Release Identifier Rules

- **Alpha**: `-alpha.{N}` where N starts at 1 and increments sequentially
- **Beta**: `-beta.{N}` where N starts at 1 and increments sequentially
- **RC**: `-rc.{N}` where N starts at 1 and increments sequentially
- **GA**: No suffix

**IMPORTANT**: Version numbers MUST NOT be reused. If a tag is published and later found defective, increment the pre-release number or patch version.

### 5.4 Version Progression

Valid progression examples:
- `v0.1.0-alpha.1` → `v0.1.0-alpha.2` → `v0.1.0-beta.1` → `v0.1.0-rc.1` → `v0.1.0`
- `v1.0.0` → `v1.0.1` (patch) → `v1.1.0` (minor) → `v2.0.0` (major)

**PROHIBITED**:
- Skipping maturity levels for the same version (e.g., `alpha.1` → `rc.1` without beta)
- Downgrading maturity (e.g., `beta.1` → `alpha.2`)
- Reusing version identifiers

---

## 5. Release Failure Conditions

A release MUST be considered **FAILED** and MUST NOT be published if any of the following conditions occur:

### 5.1 Critical Failures (Immediate Block)

1. **Test Gate Failure**: Any MUST-pass test gate fails for the declared maturity level
2. **API Documentation Staleness**: `scripts/check-docs.sh` fails (spec out of sync with code)
3. **Build Failure**: Binary compilation fails on any supported platform
4. **Migration Failure**: Database migrations fail to apply or rollback cleanly
5. **Security Vulnerability**: Critical CVE identified in dependencies or code

### 6.2 Policy Violations (Maintainer Review Required)

6. **Governance Bypass**: Release gate bypass attempted without documented consensus
7. **Tag Convention Violation**: Tag name does not follow Section 5 conventions
8. **License Compliance**: Missing or incorrect license headers in new files
9. **Breaking Change in Patch**: Non-additive changes in PATCH version increment
10. **Documentation Gap**: Required documentation (per Section 3) missing or incomplete

### 5.3 Recovery Procedures

**If a release is published and later found to violate these conditions:**

1. **Immediate Communication**: Publish security advisory or incident report
2. **Tag Deprecation**: Mark the tag as deprecated in GitHub releases
3. **Corrective Release**: Publish a new release (incremented version) that addresses the issue
4. **Root Cause Analysis**: Document failure in `docs/_internal/incidents/`
5. **Process Improvement**: Update this document or CI/CD to prevent recurrence

**DO NOT**:
- Delete or force-push to replace tags (violates immutability principle)
- Silently republish artifacts under the same tag

---

## 6. Maturity Level Graduation Criteria

### 6.1 Alpha → Beta

- At least 2 alpha releases published
- Core features functionally complete
- No known critical bugs in core flows (OAuth2/OIDC)
- All MUST-pass test gates for Beta achievable

### 6.2 Beta → RC

- At least 1 beta release with >2 weeks of community testing
- No known security vulnerabilities
- API surface stable (no further changes planned)
- Migration path tested from at least one previous beta

### 6.3 RC → GA

- At least 1 RC release with >4 weeks of production pilot testing
- Zero critical or high-severity bugs
- Security audit completed (or waived by 2+ maintainers with justification)
- Full documentation suite complete
- Support and maintenance plan defined

---

## 7. Governance and Amendments

This document is normative and changes require:

1. **Proposal**: Public issue with rationale for change
2. **Consensus**: Approval from at least 2 maintainers
3. **Documentation**: Update to this document via pull request
4. **Announcement**: Notification in project communication channels

**Effective Date**: Changes take effect immediately upon merge to `main` branch.

---

## 9. References

- [Semantic Versioning 2.0.0](https://semver.org/)
- [Release Gates](./release-gates.md)
- [Project Governance](./GOVERNANCE.md)
- [Architecture Rules](../architecture/architecture-rules.md)
