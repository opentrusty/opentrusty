#!/bin/bash
set -e

# detect-release-level.sh
# Detects the maturity level from a Git tag and outputs gate requirements.
# Returns: alpha, beta, rc, or ga

TAG_NAME="${1:-$GITHUB_REF_NAME}"

if [[ -z "$TAG_NAME" ]]; then
    echo "Error: No tag name provided" >&2
    exit 1
fi

# Strip 'refs/tags/' prefix if present
TAG_NAME="${TAG_NAME#refs/tags/}"

# Validate tag format: v{MAJOR}.{MINOR}.{PATCH}[_{MATURITY}{NUMBER}]
if [[ ! "$TAG_NAME" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(_alpha[0-9]+|_beta[0-9]+|_rc[0-9]+)?$ ]]; then
    echo "Error: Invalid tag format: $TAG_NAME" >&2
    echo "Expected: v{MAJOR}.{MINOR}.{PATCH}[_{MATURITY}{NUMBER}]" >&2
    echo "Examples: v0.1.0_alpha1, v0.2.0_beta2, v1.0.0_rc1, v1.0.0" >&2
    exit 1
fi

# Detect maturity level
if [[ "$TAG_NAME" =~ _alpha[0-9]+$ ]]; then
    LEVEL="alpha"
elif [[ "$TAG_NAME" =~ _beta[0-9]+$ ]]; then
    LEVEL="beta"
elif [[ "$TAG_NAME" =~ _rc[0-9]+$ ]]; then
    LEVEL="rc"
else
    LEVEL="ga"
fi

echo "$LEVEL"
