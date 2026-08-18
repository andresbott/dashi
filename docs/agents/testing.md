# Testing — gates, thresholds, conventions

Run these before calling any work done. See also [releasing.md](releasing.md)
for the tag-time gate.

## The gates

| Command | What it runs |
|---|---|
| `make test` | `go test ./... -cover` — fast Go suite |
| `make ui-test` | `cd webui && npm test` = `vue-tsc --noEmit` **then** vitest — type errors fail the suite before any test runs |
| `make lint` | `golangci-lint run` (config: `.golangci.yaml`) |
| `make coverage` | per-package gate: **every** package under `./internal/...` must be ≥ **70%** — not a total figure; one weak package fails the build |
| `make benchmark` | `go test -bench=. ./...` |
| `make license-check` | go-licence-detector against `allowedLicenses.json` / `overrideLicenses.json` |
| `make verify` | all of the above — the full gate |

- Coverage below 70% in a package: **write the tests, never lower
  `COVERAGE_THRESHOLD`**. Note `lib/` is currently outside the coverage gate
  (it only lists `./internal/...`) but einkimage keeps a `_test.go` per source
  file anyway — follow that convention there.
- CI (`.github/workflows/`): test + coverage + ui-test on push/PR,
  golangci-lint, license-check. Local `make verify` is a superset.

## Lint policy

`.golangci.yaml` enables standard linters + `nolintlint`, `gocyclo` (≥20),
`nestif` (≥5), `gosec`, `dupl`. Fix the code, don't silence the tool. A
`//nolint` must name the linter and carry an explanation (`nolintlint`
enforces this). Deliberate, path-scoped exclusions already exist — G203 for
trusted `template.HTML` in widgets/static rendering, G304 for the file-based
stores — so if gosec fires on new code in those areas, check whether your code
actually belongs under the existing exclusion before adding anything.
**Never edit `.golangci.yaml` without explicit user approval** (rule in
CLAUDE.md).

## Conventions

- Go: table-driven tests, `_test.go` next to the source; each widget package
  tests its image renderer output (`image_test.go` for ported widgets,
  `static_test.go` for unported ones) and, if data-backed, its client with a
  stubbed HTTP server (`client_test.go` or `handler_test.go`).
- Handler auth behavior has dedicated tests: `dashboards_auth_test.go`,
  `middleware_auth_test.go`, `store_auth_test.go` — extend these when touching
  the auth path.
- Frontend: vitest + `@vue/test-utils`, jsdom; tests colocated
  (`*.test.ts` in `composables/` and `lib/api/`). TypeScript strictness is a
  test gate via `vue-tsc`.
- Clients accept an injectable `*http.Client` (and xkcd a `nowFn`) precisely so
  tests don't hit real APIs — keep that pattern for new clients.
