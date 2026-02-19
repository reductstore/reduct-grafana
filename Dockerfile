FROM docker.io/ubuntu/grafana:12-24.04

USER root

# Create Pebble/Grafana data paths and set ownership
RUN mkdir -p /data/plugins /data/log && \
    chown -R 472:472 /data

# Copy plugin built in the release workflow artifact.
COPY dist/reductstore-datasource /data/plugins/reductstore-datasource
RUN chown -R 472:472 /data/plugins

# Allow unsigned (until plugin is signed)
ENV GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS="reductstore-datasource"

# Drop privileges
USER 472:472

EXPOSE 3000
