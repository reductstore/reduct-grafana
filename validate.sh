#!/usr/bin/env bash
# Build, package, and validate the plugin using Grafana's plugin-validator-cli.
# Usage: ./validate.sh
set -e

PLUGIN_ID="reductstore-datasource"
BUILD_DIR=".build"
ZIP_PATH="$BUILD_DIR/${PLUGIN_ID}.zip"

echo "==> Building frontend"
npm run build

echo "==> Building backend (Linux only)"
mage -v build:linux || true

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
  grafana/plugin-validator-cli \
    -strict \
    -analyzer=osv-scanner \
    -sourceCodeUri file:///src \
    "$ZIP_PATH"

echo "==> Running plugin validator (metadata)"
sudo docker run --pull=always --rm \
  -v "$PWD/$ZIP_PATH":/archive.zip \
  grafana/plugin-validator-cli -analyzer=metadatavalid /archive.zip

echo "==> Validation complete"
