#!/bin/bash
set -e

# Script to update version numbers in README.md
# Usage: ./scripts/update-readme-version.sh <new-version>
# Example: ./scripts/update-readme-version.sh v0.0.3

if [ -z "$1" ]; then
    echo "Error: Version argument required"
    echo "Usage: $0 <version>"
    echo "Example: $0 v0.0.3"
    exit 1
fi

NEW_VERSION="$1"
README_FILE="README.md"

if [ ! -f "$README_FILE" ]; then
    echo "Error: README.md not found"
    exit 1
fi

echo "Updating README.md to version $NEW_VERSION..."

# Update all version references in download URLs
# This replaces patterns like /releases/download/v0.0.1/ with the new version
# and jenkins_v0.0.1_ with the new version

# Use a temp file for cross-platform compatibility
TMP_FILE=$(mktemp)
cp "$README_FILE" "$TMP_FILE"

# Replace version references
sed "s|/releases/download/v[0-9]\+\.[0-9]\+\.[0-9]\+/|/releases/download/$NEW_VERSION/|g" "$TMP_FILE" | \
sed "s|jenkins_v[0-9]\+\.[0-9]\+\.[0-9]\+_|jenkins_${NEW_VERSION}_|g" > "$README_FILE"

# Clean up
rm -f "$TMP_FILE"

echo "README.md updated successfully to $NEW_VERSION"
