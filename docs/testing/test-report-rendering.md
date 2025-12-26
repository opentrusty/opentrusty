# Test Report Rendering

**Status**: Normative
**Owner**: OpenTrusty Maintainers

This document defines the rules for rendering JSON Test Reports (defined in `test-report-schema.md`) into human-readable Markdown artifacts. These artifacts are published to `docs.opentrusty.org` as part of the release documentation.

## 1. Rendering Philosophy

-   **Technical**: Reports are for engineers and auditors. Avoid marketing language.
-   **Neutral**: State facts (Pass/Fail). Do not obscure failures.
-   **Transparent**: Include environment and commit details to ensure reproducibility.
-   **Non-Commercial**: No external branding or "powered by" messages.

## 2. Standard Markdown Layout

All rendered reports MUST follow this structure:

### 2.1 Header Section
-   **Title**: `[Test Type] Report: [Suite Name]`
-   **Date**: `YYYY-MM-DD HH:MM UTC`
-   **Version**: `[Git Tag]` (or Commit SHA if no tag)

### 2.2 Executive Summary (Table)
A high-level summary table providing immediate status visibility.

| Metric | Value |
| :--- | :--- |
| **Status** | ✅ PASS / ❌ FAIL |
| **Total Tests** | `[total_tests]` |
| **Pass Rate** | `[calculated %]` |
| **Duration** | `[duration_ms] ms` |
| **Commit** | `[git_commit]` |

### 2.3 Environment Metadata (Collapsible)
A `<details>` section containing the environment specifics to aid debugging without cluttering the view.

```markdown
<details>
<summary>Environment Details</summary>

- **OS**: `[os]`
- **Arch**: `[arch]`
- **Runtime**: `[runtime]`
- **CI Job**: [Link]([ci_job_url])

</details>
```

### 2.4 Failures Section
**Condition**: Only render if `failures` array is non-empty.

| Test Name | Category | Reason |
| :--- | :--- | :--- |
| `[test_name]` | `[category]` | `[reason]` |
| ... | ... | ... |

*Note: If > 50 failures, truncate the table and provide a link to the raw JSON artifact.*

## 3. Artifact Naming Conventions (Rendered)

Rendered Markdown files MUST correspond to their JSON source files:

| Source JSON | Rendered Markdown |
| :--- | :--- |
| `test-report-unit.json` | `test-report-unit.md` |
| `test-report-integration.json` | `test-report-integration.md` |
| `test-report-system.json` | `test-report-system.md` |
| `test-report-e2e.json` | `test-report-e2e.md` |

## 4. Versioning & Publication

-   Reports are immutable once published for a GA release.
-   Reports are stored in the documentation history under `versions/[tag]/testing/`.
-   The "latest" documentation site generally points to the report for the latest GA release.

## 5. Rendering Tooling

Implementations of this rendering logic (e.g., `scripts/generate-test-report.sh`) MUST:
1.  Parse the JSON strictly.
2.  Escape Markdown characters in user-provided strings (test names, errors) to prevent formatting brokenness.
3.  Calculate the "Pass Rate" as `(passed / total_tests) * 100` rounded to 2 decimal places.
4.  Determine overall **Status**:
    -   **PASS**: `failed == 0` AND `total_tests > 0`
    -   **FAIL**: `failed > 0` OR `total_tests == 0` (No tests run is a failure)
