# reduct-grafana

Data Source Grafana Plugin for ReductStore

## 🧑‍💻 Development

### Frontend Development

Start the frontend development server with hot reload:

```bash
npm run dev
```

This builds the frontend with sourcemaps and watches for changes.

### Backend Development & Debugging

Build the backend:

```bash
mage build:backend
```

Start the development environment with debugging:

```bash
DEVELOPMENT=true docker compose up --build
```

This starts:

- Grafana with the plugin installed
- ReductStore database
- Go backend with debugger attached to port `2345`

### Testing with Other Grafana Versions

To test your plugin against a specific Grafana version, run:

```bash
GRAFANA_VERSION=9.5.16 docker compose up --build
```

You can replace `9.5.16` with any other Grafana version as needed.

## 🚀 Running Grafana with ReductStore Plugin

Using Docker:

```bash
TAG=v0.1.1
VERSION="${TAG#v}"
sudo docker run -d -p 3000:3000 --name=grafana \
  -e "GF_PLUGINS_PREINSTALL=reductstore-datasource@@https://github.com/reductstore/reduct-grafana/releases/download/${TAG}/reductstore-datasource-${VERSION}.zip" \
  -e "GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=reductstore-datasource" \
  ubuntu/grafana:11.6-24.04_stable
```

Or build the Docker image yourself:

```bash
docker build -t ghcr.io/reductstore/grafana:${TAG} --build-arg TAG=${TAG} .
```

Push to GitHub registry:

```bash
docker push ghcr.io/reductstore/grafana:${TAG}
```

Test locally:

```bash
docker run --rm -p 3000:3000 --name grafana-test ghcr.io/reductstore/grafana:${TAG}
```

Check installed plugins:

```bash
docker logs grafana | grep -i "installing plugin"
```

Then open your browser at [http://localhost:3000](http://localhost:3000) (admin/admin).

With juju and microk8s:

```bash
juju refresh grafana --resource grafana-image=ghcr.io/reductstore/grafana:dev
```
