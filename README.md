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

## Linting & Testing

```bash
npm run typecheck          # TypeScript type check
npm run lint               # ESLint
npm run lint:fix           # ESLint with auto-fix
npm run test:ci            # Frontend unit tests
npm run e2e                # Playwright e2e tests

mage coverage              # Backend unit tests with coverage
go test -tags=integration ./pkg/... -v -cover  # Backend integration tests
```

## Plugin Validation

For convenience, you can run the validation script which builds the plugin, packages it, and runs the Grafana plugin validator:

```bash
./validate.sh
```
