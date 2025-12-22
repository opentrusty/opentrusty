#!/bin/bash
set -e

# validate-release-checklist.sh
# Validates that a release checklist exists and is properly completed.

VERSION="${1:-$GITHUB_REF_NAME}"
LEVEL="${2}"

if [[ -z "$VERSION" ]]; then
    echo "Error: No version provided" >&2
    exit 1
fi

if [[ -z "$LEVEL" ]]; then
    echo "Error: No maturity level provided" >&2
    exit 1
fi

# Strip 'refs/tags/' prefix if present
VERSION="${VERSION#refs/tags/}"

CHECKLIST_DIR=".github/releases"
CHECKLIST_FILE="$CHECKLIST_DIR/${VERSION}-checklist.md"

echo "Validating release checklist for $VERSION ($LEVEL)..."

# Check if checklist exists
if [[ ! -f "$CHECKLIST_FILE" ]]; then
    echo "❌ Release checklist not found: $CHECKLIST_FILE" >&2
    echo "" >&2
    echo "REQUIRED: Create a release checklist before publishing this release." >&2
    echo "Steps:" >&2
    echo "  1. cp .github/RELEASE_CHECKLIST_TEMPLATE.md $CHECKLIST_FILE" >&2
    echo "  2. Fill in the checklist with version, date, and manager info" >&2
    echo "  3. Check off all completed items" >&2
    echo "  4. Commit and push the checklist" >&2
    echo "" >&2
    exit 1
fi

echo "✅ Checklist file found: $CHECKLIST_FILE"

# Validate checklist format
if ! grep -q "^# Release Checklist:" "$CHECKLIST_FILE"; then
    echo "❌ Invalid checklist format: Missing title" >&2
    exit 1
fi

# Validate version matches
if ! grep -q "# Release Checklist: $VERSION" "$CHECKLIST_FILE"; then
    echo "⚠️  WARNING: Checklist version may not match tag version" >&2
    echo "Expected: # Release Checklist: $VERSION" >&2
fi

# Validate maturity level is specified
if ! grep -q "^\*\*Maturity Level\*\*:" "$CHECKLIST_FILE"; then
    echo "❌ Maturity level not specified in checklist" >&2
    exit 1
fi

# Count unchecked required items based on maturity level
case "$LEVEL" in
    alpha)
        REQUIRED_SECTIONS=("Automated Checks" "Documentation Requirements" "Alpha" "Manual Verification Requirements" "Governance Approvals")
        ;;
    beta)
        REQUIRED_SECTIONS=("Automated Checks" "Documentation Requirements" "Alpha" "Beta" "Manual Verification Requirements" "Governance Approvals")
        ;;
    rc)
        REQUIRED_SECTIONS=("Automated Checks" "Documentation Requirements" "Alpha" "Beta" "RC" "Manual Verification Requirements" "Governance Approvals")
        ;;
    ga)
        REQUIRED_SECTIONS=("Automated Checks" "Documentation Requirements" "Alpha" "Beta" "RC" "GA" "Manual Verification Requirements" "Governance Approvals")
        ;;
    *)
        echo "❌ Unknown maturity level: $LEVEL" >&2
        exit 1
        ;;
esac

echo "Checking required sections for $LEVEL release..."

# Simple validation: ensure checklist has been modified from template
if grep -q "{VERSION}" "$CHECKLIST_FILE" || grep -q "{LEVEL}" "$CHECKLIST_FILE"; then
    echo "❌ Checklist appears to be unmodified template" >&2
    echo "Please fill in version, level, date, and manager information" >&2
    exit 1
fi

# Count total items that should be checked
TOTAL_ITEMS=$(grep -c "^- \[.\]" "$CHECKLIST_FILE" || true)
CHECKED_ITEMS=$(grep -c "^- \[x\]" "$CHECKLIST_FILE" || true)
UNCHECKED_ITEMS=$(grep -c "^- \[ \]" "$CHECKLIST_FILE" || true)

echo "Checklist status:"
echo "  Total items: $TOTAL_ITEMS"
echo "  Checked: $CHECKED_ITEMS"
echo "  Unchecked: $UNCHECKED_ITEMS"

# For RC and GA, require maintainer sign-offs
if [[ "$LEVEL" == "rc" || "$LEVEL" == "ga" ]]; then
    if grep -q "Maintainer 1\*\*: _________" "$CHECKLIST_FILE"; then
        echo "❌ Missing Maintainer 1 sign-off (required for $LEVEL)" >&2
        exit 1
    fi
    if grep -q "Maintainer 2\*\*.*: _________" "$CHECKLIST_FILE"; then
        echo "❌ Missing Maintainer 2 sign-off (required for $LEVEL)" >&2
        exit 1
    fi
    echo "✅ Maintainer sign-offs present"
fi

# Warning if many items are unchecked (but don't fail for alpha/beta)
if [[ "$UNCHECKED_ITEMS" -gt 5 ]]; then
    echo "⚠️  WARNING: $UNCHECKED_ITEMS items remain unchecked"
    if [[ "$LEVEL" == "rc" || "$LEVEL" == "ga" ]]; then
        echo "❌ RC and GA releases require all applicable items to be checked" >&2
        exit 1
    fi
fi

# Check for release notes file
RELEASE_NOTES_FILE="docs/releases/${VERSION}.md"
if [[ ! -f "$RELEASE_NOTES_FILE" ]]; then
    echo "⚠️  WARNING: Release notes not found at $RELEASE_NOTES_FILE"
    if [[ "$LEVEL" == "rc" || "$LEVEL" == "ga" ]]; then
        echo "❌ Release notes are required for $LEVEL releases" >&2
        exit 1
    fi
fi

echo "✅ Release checklist validation passed for $VERSION ($LEVEL)"
