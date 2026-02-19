FROM docker.io/grafana/grafana:12.0.2

USER root

# Create plugin path and set ownership
RUN mkdir -p /var/lib/grafana/plugins && \
    chown -R 472:472 /var/lib/grafana

# Copy plugin files (supports both local `dist/*` and CI artifact `dist/reductstore-datasource/*` layouts).
COPY dist/ /tmp/reductstore-datasource/
RUN if [ -d /tmp/reductstore-datasource/reductstore-datasource ]; then \
      cp -R /tmp/reductstore-datasource/reductstore-datasource /var/lib/grafana/plugins/reductstore-datasource; \
    else \
      cp -R /tmp/reductstore-datasource /var/lib/grafana/plugins/reductstore-datasource; \
    fi && \
    rm -rf /tmp/reductstore-datasource && \
    find /var/lib/grafana/plugins/reductstore-datasource -type f \( -name 'gpx_reductstore_linux_*' -o -name 'reductstore-datasource_linux_*' \) -exec chmod 0755 {} + && \
    chown -R 472:472 /var/lib/grafana/plugins

# Allow unsigned (until plugin is signed)
ENV GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS="reductstore-datasource"

# Drop privileges
USER 472:472

EXPOSE 3000
