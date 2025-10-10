#!/usr/bin/env bash
set -e

export GF_PLUGINS_PREINSTALL="${GF_PLUGINS_PREINSTALL:-reductstore-datasource@@https://github.com/reductstore/reduct-grafana/releases/download/v0.1.0/reductstore-datasource-0.1.0.zip}"
export GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS="${GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS:-reductstore-datasource}"

echo "🧩 Preinstalling plugins: $GF_PLUGINS_PREINSTALL"
echo "🔓 Allowing unsigned plugins: $GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS"
echo "🚀 Starting Grafana..."

exec /run.sh
