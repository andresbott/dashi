# Dashi

Go + Vue.js self-hosted dashboard application.

## Documentation

Before implementing, read `docs/agents/architecture.md` and the subsystem doc
for the area you're changing:

- `docs/agents/architecture.md` — layering, package roles, design decisions, known debt (entry point)
- `docs/agents/features.md` — feature status: implemented / partial / deliberately missing
- `docs/agents/widgets.md` — dual widget registries, config contract, how to add a widget
- `docs/agents/viewer.md` — browser stack, assets, two render stacks, JS contract
- `docs/agents/rendering.md` — image/PNG pipeline, display protocol, dithering
- `docs/agents/testing.md` — gates, coverage threshold, lint policy
- `docs/agents/releasing.md` — tagging, goreleaser, embedded-UI trap

`docs/autodoc/` is the older generated doc set — historical record only; where
it conflicts with `docs/agents/` or the code, the code wins. Design specs live
in `docs/superpowers/`, analysis docs in `docs/project/`.

## Commands

- `make test` — Go unit tests
- `make ui-test` — Vue.js tests (vue-tsc + vitest)
- `make lint` — golangci-lint
- `make coverage` — 70% per-package threshold on ./internal/...
- `make verify` — all tests + lint + coverage + benchmark + license
- `make run` — Start Go server (debug mode, no embedded UI)
- `make run-ui` — Build UI + embed + start server
- `cd webui && npm run dev` — Vite dev server (proxies API to editor on :8088)

## Rules

- Never edit `.golangci.yaml` (or any `.golangci.*` config) without explicit user approval. When a linter flags code, fix the code or use a targeted inline `//nolint:<linter> // reason` directive instead. Only change lint config after the user says yes.
- Never hand-edit `app/spa/files/ui/`, `webui/dist/`, or `internal/dashboard/browser/assets/viewer.css` — all are build artifacts. The first two are produced by `make package-ui`; viewer.css is produced by `make viewer-css`.
