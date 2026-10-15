#!/usr/bin/env bash
# Build, package, and validate the plugin using Grafana's plugin-validator-cli.
# Usage: ./validate.sh
set -e

PLUGIN_ID="reductstore-datasource"
BUILD_DIR=".build"
ZIP_PATH="$BUILD_DIR/${PLUGIN_ID}.zip"
VALIDATOR_IMAGE="docker.io/grafana/plugin-validator-cli@sha256:404b0b3af8db3d27d2b481737c8b2444ba48395860fc5a5197d274553180778b"


echo "==> Installing locked frontend dependencies"
npm ci

echo "==> Building frontend"
npm run build

echo "==> Verifying Go modules"
GOFLAGS=-mod=readonly go mod download
GOFLAGS=-mod=readonly go mod verify

echo "==> Building backend (Linux only)"
GOFLAGS=-mod=readonly go run github.com/magefile/mage -v build:linux

echo "==> Preparing build dir"
mkdir -p "$BUILD_DIR"
rm -rf "$BUILD_DIR/$PLUGIN_ID" "$ZIP_PATH"

cp -r dist "$BUILD_DIR/$PLUGIN_ID"

echo "==> Creating ZIP: $ZIP_PATH"
(cd "$BUILD_DIR" && zip -qr "${PLUGIN_ID}.zip" "$PLUGIN_ID")

echo "==> Running plugin validator (osv-scanner)"
sudo docker run --pull=always --rm \
  -v "$PWD:/src" \
  -w /src \
  "$VALIDATOR_IMAGE" \
    -strict \
    -analyzer=osv-scanner \
    -sourceCodeUri file:///src \
    "$ZIP_PATH"

echo "==> Running plugin validator (metadata)"
sudo docker run --pull=always --rm \
  -v "$PWD/$ZIP_PATH":/archive.zip \
  "$VALIDATOR_IMAGE" -analyzer=metadatavalid /archive.zip

echo "==> Validation complete"
