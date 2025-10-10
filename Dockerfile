FROM docker.io/ubuntu/grafana:11.6-24.04_stable

USER root

# tools for plugin install
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl unzip && \
    rm -rf /var/lib/apt/lists/*

# Create Pebble/Grafana data paths and set ownership
RUN mkdir -p /data/plugins /data/log && \
    chown -R 472:472 /data

# Install ReductStore plugin into /data/plugins
ARG REDUCT_PLUGIN_URL="https://github.com/reductstore/reduct-grafana/releases/download/v0.1.0/reductstore-datasource-0.1.0.zip"
RUN curl -fsSL "$REDUCT_PLUGIN_URL" -o /tmp/reduct.zip && \
    unzip -q /tmp/reduct.zip -d /data/plugins && \
    rm /tmp/reduct.zip && \
    chown -R 472:472 /data/plugins

# Allow unsigned (until plugin is signed)
ENV GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS="reductstore-datasource"

# (Optional) Serve under a sub-path in front of Traefik:
ENV GF_SERVER_ROOT_URL="/cos-robotics-model-grafana/"
ENV GF_SERVER_SERVE_FROM_SUB_PATH="true"

# Drop privileges
USER 472:472

EXPOSE 3000
