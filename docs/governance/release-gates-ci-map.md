# Release Gates CI Mapping

**Status**: Normative
**Owner**: OpenTrusty Maintainers
**Source**: Derived from [Release Maturity Model](./release-maturity-model.md) and [Release Gates](./release-gates.md)

This document provides a machine-executable mapping of the Release Gates to CI/CD jobs. It is intended to be the source of truth for CI configuration logic.

## 1. Alpha Gates
**Trigger Context**: `push` to feature branches, `tag` matching `v*-alpha.*`
**Strictness**: Low (Experimental)

| Gate ID | Gate Description | CI Job Name | Trigger | Required Inputs | Produced Artifacts | Failure Condition |
|---------|------------------|-------------|---------|-----------------|--------------------|-------------------|
| UT-01 | Unit Tests | `test-unit` | `push`, `tag` | Source Code | `ut-report.json`, `coverage.out` | Any test failure or compilation error |
| DOC-01 | API Documentation Freshness | `check-docs` | `push`, `tag` | Source Code, `swagger.json` | `swagger-validation.log` | `scripts/check-docs.sh` exits non-zero |
| INT-01 | Integration Tests | `test-integration` | `push`, `tag` | Source Code, PostgreSQL | `st-report.json` | **OPTIONAL** (Warning only) |
| E2E-01 | E2E Tests (Docker) | `test-e2e` | `tag` | Docker Image | `e2e-report.json` | **OPTIONAL** (Manual review allowed) |

## 2. Beta Gates
**Trigger Context**: `tag` matching `v*-beta.*`
**Strictness**: Medium (Feature Complete)

| Gate ID | Gate Description | CI Job Name | Trigger | Required Inputs | Produced Artifacts | Failure Condition |
|---------|------------------|-------------|---------|-----------------|--------------------|-------------------|
| UT-01 | Unit Tests | `test-unit` | `tag` | Source Code | `ut-report.json`, `coverage.out` | Any test failure |
| DOC-01 | API Documentation Freshness | `check-docs` | `tag` | Source Code, `swagger.json` | `swagger-validation.log` | Script failure |
| INT-01 | Integration Tests | `test-integration` | `tag` | Source Code, PostgreSQL | `st-report.json` | **REQUIRED** (Any failure blocks release) |
| E2E-01 | E2E Tests (Docker) | `test-e2e` | `tag` | Docker Image | `e2e-report.json` | **REQUIRED** (Any failure blocks release) |
| SYS-01 | Systemd Smoke Test | `test-systemd` | `tag` | Binary Artifact | `systemd-test.log` | **OPTIONAL** (Warning only) |

## 3. Release Candidate (RC) Gates
**Trigger Context**: `tag` matching `v*-rc.*`
**Strictness**: High (Production Ready)

| Gate ID | Gate Description | CI Job Name | Trigger | Required Inputs | Produced Artifacts | Failure Condition |
|---------|------------------|-------------|---------|-----------------|--------------------|-------------------|
| UT-01 | Unit Tests | `test-unit` | `tag` | Source Code | `ut-report.json` | **STRICT** |
| DOC-01 | API Documentation Freshness | `check-docs` | `tag` | Source Code, `swagger.json` | `index.html` (Bundled Docs) | **STRICT** |
| INT-01 | Integration Tests | `test-integration` | `tag` | Source Code, PostgreSQL | `st-report.json` | **STRICT** |
| E2E-01 | E2E Tests (Docker) | `test-e2e` | `tag` | Docker Image | `e2e-report.json` | **STRICT** |
| SYS-01 | Systemd Smoke Test | `test-systemd` | `tag` | Binary Artifact | `systemd-test.log` | **STRICT** |
| SEC-01 | Security Scan | `security-scan` | `tag` | Source Code, Dep Tree | `trivy-report.json` | Critical/High Vulnerabilities found |
| PERF-01 | Performance Regression | `benchmark` | `tag` | Source Code | `bench-results.txt` | **OPTIONAL** (Comparison > 15% regression) |
| REL-01 | Release Checklist | `checklist-init` | `tag` | Template | `issue-checklist` | GitHub Issue not created |

## 4. General Availability (GA) Gates
**Trigger Context**: Promotion Workflow (Manual Dispatch) targeting `v*` (no suffix)
**Strictness**: Absolute (Immutable)

| Gate ID | Gate Description | CI Job Name | Trigger | Required Inputs | Produced Artifacts | Failure Condition |
|---------|------------------|-------------|---------|-----------------|--------------------|-------------------|
| UT-01 | Unit Tests | `test-unit` | `promote` | Source Code | `ut-report.json` | **STRICT** |
| DOC-01 | API Documentation Freshness | `check-docs` | `promote` | Source Code, `swagger.json` | `public-docs-site` | **STRICT** |
| INT-01 | Integration Tests | `test-integration` | `promote` | Source Code, PostgreSQL | `st-report.json` | **STRICT** |
| E2E-01 | E2E Tests (Docker) | `test-e2e` | `promote` | Docker Image | `e2e-report.json` | **STRICT** |
| SYS-01 | Systemd Smoke Test | `test-systemd` | `promote` | Binary Artifact | `systemd-test.log` | **STRICT** |
| SEC-01 | Security Scan | `security-scan` | `promote` | Source Code | `trivy-report.json` | **STRICT** (Any finding blocks) |
| PERF-01 | Performance Regression | `benchmark` | `promote` | Source Code | `bench-results.txt` | **STRICT** (> 5% regression blocks) |
| UPG-01 | Upgrade Path Validation | `upgrade-test` | `promote` | Previous GA, Current GA | `upgrade-log.txt` | Migration failure or data loss |
| REL-01 | Release Checklist | `checklist-verify` | `promote` | Issue ID | `verification-result` | Checklist not marked 100% complete |

## Gate Escalation Rules

1.  **Optional to Required**: `INT-01` (Integration) and `E2E-01` (E2E) transition from **OPTIONAL** in Alpha to **REQUIRED** in Beta.
2.  **Warning to Blocking**: `SYS-01` (Systemd) transitions from **OPTIONAL** (Warning) in Beta to **STRICT** (Blocking) in RC.
3.  **Threshold Tightening**: `PERF-01` (Performance) tightens from "Should Pass" (15% allowance) in RC to **STRICT** (5% allowance) in GA.
4.  **No Bypass**: At RC and GA levels, `SEC-01` (Security) and `REL-01` (Checklist) cannot be bypassed by automated means; they require documented maintainer overrides only if allowed by Governance.
