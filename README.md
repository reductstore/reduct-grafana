# ReductStore Grafana Data Source

This repository contains the Grafana data source plugin for ReductStore.

## Documentation

Usage and configuration details are on the website: [www.reduct.store/docs/next/integrations/grafana](https://www.reduct.store/docs/next/integrations/grafana)

## Development

Start frontend development

```bash
npm run dev
```

Build backend

```bash
mage build:backend
```

Start full development setup

```bash
DEVELOPMENT=true docker compose up --build
```

Test against another Grafana version

```bash
GRAFANA_VERSION=9.5.16 docker compose up --build
```
