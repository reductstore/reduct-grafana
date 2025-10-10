# reduct-grafana

Data Source Grafana Plugin for ReductStore

## Installation

Using Docker:

```bash
sudo docker run -d -p 3000:3000 --name=grafana \
  -e "GF_PLUGINS_PREINSTALL=reductstore-datasource@@https://github.com/reductstore/reduct-grafana/releases/download/v0.1.0/reductstore-datasource-0.1.0.zip" \
  -e "GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=reductstore-datasource" \
  ubuntu/grafana:11.6-24.04_stable
```

Or build the Docker image yourself:

```bash
docker build --no-cache -t ghcr.io/reductstore/grafana:dev .
```

Push to GitHub registry:

```bash
docker push ghcr.io/reductstore/grafana:dev
```

Test locally:
```bash
docker run --rm -p 3000:3000 --name grafana-test ghcr.io/reductstore/grafana:dev
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
