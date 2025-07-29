#!/bin/bash

set -e

# Usage: ./update-mod.sh github.com/youruser/yourmodule[@version]
# If @version is omitted, defaults to @latest

if [ -z "$1" ]; then
  echo "Usage: $0 <module-path>[@version]"
  exit 1
fi

# Split path and optional version
FULL_ARG="$1"
MODULE_PATH="${FULL_ARG%%@*}"
VERSION="${FULL_ARG#*@}"

if [ "$MODULE_PATH" = "$VERSION" ]; then
  VERSION="latest"
fi

echo "Updating module: $MODULE_PATH@$VERSION"

# Set environment to bypass the proxy for this module
export GOPRIVATE=$MODULE_PATH
export GONOSUMDB=$MODULE_PATH

# Fetch the specified version directly from GitHub
go get "${MODULE_PATH}@${VERSION}"

# Clean up go.mod and go.sum
go mod tidy

echo "✅ Module updated to $MODULE_PATH@$VERSION"

