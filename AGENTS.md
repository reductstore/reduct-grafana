# Repository Guidelines

## Project Structure & Module Organization
- `src/`: Frontend datasource plugin code (TypeScript/React), including editors in `src/components/` and plugin metadata in `src/plugin.json`.
- `pkg/`: Backend Go implementation. Core datasource/query logic is in `pkg/plugin/`; shared models are in `pkg/models/`.
- `tests/`: Playwright end-to-end tests (`*.spec.ts`) and helpers.
- `provisioning/`: Local Grafana provisioning files for datasource setup.
- `dist/`: Built plugin artifacts (generated; do not hand-edit).
- `.github/workflows/`: CI/CD pipelines; use these as the source of truth for required checks.

## Build, Test, and Development Commands
- `npm ci`: Install Node dependencies (CI-compatible install).
- `npm run dev`: Frontend watch build for local development.
- `npm run build`: Production frontend bundle.
- `npm run typecheck`: TypeScript checks without emitting files.
- `npm run lint` / `npm run lint:fix`: Lint checks and auto-fixes.
- `npm run test:ci`: Frontend unit tests (non-watch mode).
- `npm run e2e`: Run Playwright tests in `tests/`.
- `mage buildAll`: Build backend binaries.
- `mage coverage`: Run backend unit coverage task.
- `go test -tags=integration ./pkg/... -v -cover`: Backend integration tests.

## Coding Style & Naming Conventions
- Use project linters/formatters: ESLint + Prettier (Grafana config via `.config/`).
- TypeScript/React: 2-space indentation; components in `PascalCase` (for example, `QueryEditor.tsx`), helpers/modules in descriptive lowercase file names.
- Go: follow `gofmt` defaults; keep package names lowercase and concise.

## Testing Guidelines
- Frontend unit tests: Jest files named `*.test.ts` or `*.test.tsx` (co-located under `src/`).
- E2E tests: Playwright specs under `tests/*.spec.ts`.
- Backend tests: `*_unit_test.go` for unit scope, `*_integration_test.go` for integration scope.
- Run the same checks as CI before opening a PR: typecheck, lint, unit tests, build, and relevant backend/e2e tests.

## Commit & Pull Request Guidelines
- Follow existing history style: short imperative subject lines, often with PR reference (for example, `Update Grafana go SDK (#38)`).
- Keep commits focused; avoid mixing frontend, backend, and infra refactors in one commit unless tightly related.
- PRs should include: clear summary, linked issue(s), test evidence (commands run), and UI screenshots/GIFs for editor or query UX changes.
- Ensure CI passes (`.github/workflows/ci.yml`) before requesting review.

## Security & Configuration Tips
- Use Node `>=22` (see `package.json` engines).
- Do not commit secrets (for example, `GRAFANA_ACCESS_POLICY_TOKEN`).
- For local matrix checks, use `GRAFANA_VERSION`/`GRAFANA_IMAGE` with `docker compose` as in CI.
