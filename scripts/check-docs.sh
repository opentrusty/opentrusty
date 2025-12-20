#!/bin/bash
set -e

# check-docs.sh
# Verifies that the generated API documentation is up-to-date with the code.

echo "Checking API documentation freshness..."

# Ensure we are in the project root
cd "$(dirname "$0")/.."

# Run generation
export PATH=$PATH:$(go env GOPATH)/bin
make docs-gen

# Check for changes in docs/api
if git diff --quiet docs/api; then
    echo "✅ API documentation is up-to-date."
    exit 0
else
    echo "❌ API documentation is stale!"
    echo "The following files have changed after running 'make docs-gen':"
    git diff --name-only docs/api
    echo ""
    echo "Please run 'make docs-gen' locally and commit the changes."
    exit 1
fi
