# Changelog

## [Unreleased]

## [1.0.1] - 2026-02-09

### Changed

- Replaced local Monaco completion provider with `@reductstore/reduct-query-monaco` package, [PR-33](https://github.com/reductstore/reduct-grafana/pull/33)
- Update reduct-go to v1.18.0, [PR-35](https://github.com/reductstore/reduct-grafana/pull/35)

## [1.0.0] - 2025-12-11

### Added

- Redesigned condition editor that supports completion suggestions, [PR-28](https://github.com/reductstore/reduct-grafana/pull/28)
- Add support for Grafana `$__interval` macro, [PR-30](https://github.com/reductstore/reduct-grafana/pull/30)

## [0.1.1] - 2025-10-22

### Fixed

- Fix compatibility issue on older Grafana versions without Combobox, [PR-18](https://github.com/reductstore/reduct-grafana/pull/18)

## [0.1.0] - 2025-10-01

### Added

- Added initial configuration setup to connect to ReductStore, [PR-6](https://github.com/reductstore/reduct-grafana/pull/6)
- Query labels from entries in ReductStore, [PR-7](https://github.com/reductstore/reduct-grafana/pull/7)
- Add support for when condition in queries, [PR-13](https://github.com/reductstore/reduct-grafana/pull/13)
- Add support for JSON content in queries, [PR-14](https://github.com/reductstore/reduct-grafana/pull/14)

### Fixed

- Fix label type conversion in queries [PR-8](https://github.com/reductstore/reduct-grafana/pull/8)
