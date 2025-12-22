# Release Checklist Process

This document describes how to use the release checklist system to ensure all requirements from the [Release Maturity Model](../governance/release-maturity-model.md) are met before publishing a release.

## Overview

Every OpenTrusty release MUST have a completed release checklist validated by CI before the release can be published. The checklist includes both automated checks (verified by CI) and manual verification requirements that must be signed off by maintainers.

## Creating a Release Checklist

### 1. Copy the Template

```bash
# Replace {VERSION} with your release tag (e.g., v0.1.0_alpha1)
cp .github/RELEASE_CHECKLIST_TEMPLATE.md .github/releases/{VERSION}-checklist.md
```

### 2. Fill in Release Information

Edit the copied file and replace all template placeholders:

- `{VERSION}`: The exact git tag (e.g., v0.1.0_alpha1)
- `{LEVEL}`: Maturity level (alpha, beta, rc, or ga)
- `{DATE}`: Target release date
- `{MANAGER}`: Name of the release manager

### 3. Complete the Checklist

Work through each section and check off items as they are completed:

**Automated Checks**: These will be verified by CI and will automatically fail if not passing

**Documentation Requirements**: Ensure all required documentation exists for your maturity level
- Alpha: API docs + known issues
- Beta: + migration guide + limitations
- RC: + threat model review + deployment guides
- GA: + production best practices + monitoring guide

**Manual Verification Requirements**: These require human review and testing

**Governance Approvals**: Verify compliance with project governance

### 4. Obtain Maintainer Sign-offs

- **Alpha/Beta**: One maintainer sign-off recommended
- **RC/GA**: Two maintainer sign-offs REQUIRED

Replace the underscores with actual names and dates:
```markdown
**Maintainer 1**: John Doe       Date: 2025-12-22  
**Maintainer 2**: Jane Smith     Date: 2025-12-22
```

### 5. Create Release Notes

Create a release notes file at `docs/releases/{VERSION}.md` with:
- Summary of changes
- Breaking changes (if any)
- Known issues
- Migration guide (if applicable)
- Upgrade instructions

### 6. Commit the Checklist

```bash
git add .github/releases/{VERSION}-checklist.md docs/releases/{VERSION}.md
git commit -m "chore: Add release checklist for {VERSION}"
git push origin main
```

### 7. Tag and Push the Release

```bash
git tag {VERSION}
git push origin {VERSION}
```

The CI will automatically:
1. Detect the maturity level from the tag
2. Validate the release checklist exists and is complete
3. Run all required test gates for the maturity level
4. Generate and publish documentation (if all gates pass)

## Validation

The release checklist is validated by `scripts/validate-release-checklist.sh` which checks:

✅ **Checklist file exists** at `.github/releases/{VERSION}-checklist.md`  
✅ **Format is valid** (title, maturity level, etc.)  
✅ **Template has been filled out** (no `{PLACEHOLDERS}` remain)  
✅ **Maintainer sign-offs present** (for RC/GA)  
✅ **Release notes exist** at `docs/releases/{VERSION}.md` (for RC/GA)  
⚠️ **Warns if many items unchecked** (fails for RC/GA)

## Failure Scenarios

### Checklist Not Found

```
❌ Release checklist not found: .github/releases/v1.0.0-checklist.md

REQUIRED: Create a release checklist before publishing this release.
Steps:
  1. cp .github/RELEASE_CHECKLIST_TEMPLATE.md .github/releases/v1.0.0-checklist.md
  2. Fill in the checklist with version, date, and manager info
  3. Check off all completed items
  4. Commit and push the checklist
```

**Resolution**: Create and commit the checklist before pushing the tag.

### Incomplete Checklist (RC/GA)

```
⚠️  WARNING: 8 items remain unchecked
❌ RC and GA releases require all applicable items to be checked
```

**Resolution**: Complete all required items or downgrade to a lower maturity level.

### Missing Maintainer Sign-offs (RC/GA)

```
❌ Missing Maintainer 2 sign-off (required for rc)
```

**Resolution**: Obtain required maintainer approvals and update the checklist.

### Missing Release Notes (RC/GA)

```
⚠️  WARNING: Release notes not found at docs/releases/v1.0.0.md
❌ Release notes are required for rc releases
```

**Resolution**: Create release notes at the expected location.

## Example Workflow

```bash
# 1. Create checklist from template
cp .github/RELEASE_CHECKLIST_TEMPLATE.md .github/releases/v0.1.0_beta1-checklist.md

# 2. Edit the checklist
vim .github/releases/v0.1.0_beta1-checklist.md
# - Replace {VERSION} with v0.1.0_beta1
# - Replace {LEVEL} with beta
# - Check off completed items
# - Add maintainer sign-off

# 3. Create release notes
vim docs/releases/v0.1.0_beta1.md

# 4. Commit
git add .github/releases/v0.1.0_beta1-checklist.md docs/releases/v0.1.0_beta1.md
git commit -m "chore: Release checklist for v0.1.0_beta1"
git push origin main

# 5. Tag and release
git tag v0.1.0_beta1
git push origin v0.1.0_beta1

# CI will now validate the checklist and run all required gates
```

## Enforcement

The release checklist validation is enforced as a **required gate** in the Release Gate CI workflow. The release CANNOT proceed if:

- The checklist file is missing
- The checklist is malformed or unmodified from the template
- Required items are unchecked (for RC/GA)
- Maintainer sign-offs are missing (for RC/GA)
- Release notes are missing (for RC/GA)

This ensures that every release meets the documented standards in the [Release Maturity Model](../governance/release-maturity-model.md).

## References

- [Release Maturity Model](../governance/release-maturity-model.md) - Normative requirements per maturity level
- [Release Gates](../governance/release-gates.md) - Automated gate definitions
- [Governance](../governance/GOVERNANCE.md) - Project governance and decision-making process
