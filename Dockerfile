FROM docker.io/ubuntu/grafana:9-22.04

USER root

# tools for plugin install
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl unzip && \
    rm -rf /var/lib/apt/lists/*

# Create Pebble/Grafana data paths and set ownership
RUN mkdir -p /data/plugins /data/log && \
    chown -R 472:472 /data

# ReductStore plugin version and ID
ARG TAG=v1.0.1
ARG PLUGIN_ID=reductstore-datasource

# Install ReductStore plugin into /data/plugins
RUN VERSION="${TAG#v}" && \
    REDUCT_PLUGIN_URL="https://github.com/reductstore/reduct-grafana/releases/download/${TAG}/${PLUGIN_ID}-${VERSION}.zip" && \
    echo "Installing plugin from ${REDUCT_PLUGIN_URL}" && \
    curl -fsSL "${REDUCT_PLUGIN_URL}" -o /tmp/reduct.zip && \
    unzip -q /tmp/reduct.zip -d /data/plugins && \
    rm /tmp/reduct.zip && \
    chown -R 472:472 /data/plugins

# Allow unsigned (until plugin is signed)
ENV GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS="reductstore-datasource"

# Drop privileges
USER 472:472

EXPOSE 3000
