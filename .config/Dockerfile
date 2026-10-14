ARG grafana_version=12.4.3
ARG grafana_image=grafana-enterprise
ARG grafana_digest=sha256:c7516cc195e0c6fed177006156312cae0a90867274afe9373a45229ed20a7cf7

FROM docker.io/grafana/${grafana_image}:${grafana_version}@${grafana_digest}

ARG anonymous_auth_enabled=true
ARG development=false
ARG TARGETARCH
ARG GO_ARCH=${TARGETARCH:-amd64}
ARG GO_VERSION=1.26.5
ARG DELVE_VERSION=v1.27.2
ARG MAGE_VERSION=v1.17.2


ENV DEV="${development}"

# Make it as simple as possible to access the grafana instance for development purposes
# Do NOT enable these settings in a public facing / production grafana instance
ENV GF_AUTH_ANONYMOUS_ORG_ROLE="Admin"
ENV GF_AUTH_ANONYMOUS_ENABLED="${anonymous_auth_enabled}"
ENV GF_AUTH_BASIC_ENABLED="false"
# Set development mode so plugins can be loaded without the need to sign
ENV GF_DEFAULT_APP_MODE="development"


LABEL maintainer="Grafana Labs <hello@grafana.com>"

ENV GF_PATHS_HOME="/usr/share/grafana"
WORKDIR $GF_PATHS_HOME

USER root

# Installing supervisor and inotify-tools
RUN if [ "${development}" = "true" ]; then \
    if grep -i -q alpine /etc/issue; then \
    apk add supervisor inotify-tools git; \
    elif grep -i -q ubuntu /etc/issue; then \
    DEBIAN_FRONTEND=noninteractive && \
    apt-get update && \
    apt-get install -y supervisor inotify-tools git && \
    rm -rf /var/lib/apt/lists/*; \
    else \
    echo 'ERROR: Unsupported base image' && /bin/false; \
    fi \
    fi

COPY supervisord/supervisord.conf /etc/supervisor.d/supervisord.ini
COPY supervisord/supervisord.conf /etc/supervisor/conf.d/supervisord.conf


RUN if [ "${development}" = "true" ]; then \
    case "${GO_ARCH}" in \
      amd64) GO_SHA256=5c2c3b16caefa1d968a94c1daca04a7ca301a496d9b086e17ad77bb81393f053 ;; \
      arm64) GO_SHA256=fe4789e92b1f33358680864bbe8704289e7bb5fc207d80623c308935bd696d49 ;; \
      arm) GO_ARCH=armv6l; GO_SHA256=6dae9edab81c13bccf962dec15f1fd2ec26c14a6821b4d2c92dab4130c289d7a ;; \
      *) echo "ERROR: Unsupported Go architecture: ${GO_ARCH}" >&2; exit 1 ;; \
    esac && \
    curl --fail --location --silent --show-error --output /tmp/go.tar.gz "https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" && \
    echo "${GO_SHA256}  /tmp/go.tar.gz" | sha256sum -c - && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf /tmp/go.tar.gz && \
    rm -f /tmp/go.tar.gz && \
    /usr/local/go/bin/go install github.com/go-delve/delve/cmd/dlv@${DELVE_VERSION} && \
    /usr/local/go/bin/go install github.com/magefile/mage@${MAGE_VERSION}; \
    fi

# Inject livereload script into grafana index.html
RUN sed -i 's|</body>|<script src="http://localhost:35729/livereload.js"></script></body>|g' /usr/share/grafana/public/views/index.html


COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
