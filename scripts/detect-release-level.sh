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

# Validate tag format: v{MAJOR}.{MINOR}.{PATCH}[-{MATURITY}.{NUMBER}]
if [[ ! "$TAG_NAME" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-alpha\.[0-9]+|-beta\.[0-9]+|-rc\.[0-9]+)?$ ]]; then
    echo "Error: Invalid tag format: $TAG_NAME" >&2
    echo "Expected: v{MAJOR}.{MINOR}.{PATCH}[-{MATURITY}.{NUMBER}]" >&2
    echo "Examples: v0.1.0-alpha.1, v0.2.0-beta.2, v1.0.0-rc.1, v1.0.0" >&2
    exit 1
fi

# Detect maturity level
if [[ "$TAG_NAME" =~ -alpha\.[0-9]+$ ]]; then
    LEVEL="alpha"
elif [[ "$TAG_NAME" =~ -beta\.[0-9]+$ ]]; then
    LEVEL="beta"
elif [[ "$TAG_NAME" =~ -rc\.[0-9]+$ ]]; then
    LEVEL="rc"
else
    LEVEL="ga"
fi

echo "$LEVEL"
