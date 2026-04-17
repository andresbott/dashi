# Dashi

Go + Vue.js self-hosted dashboard application.

## Documentation

Project documentation lives in `docs/autodoc/`:

- `overview.md` — Architecture, package layout, key flows, widget system, external dependencies
- `features.md` — Feature status table (implemented / partial / not implemented)
- `decisions.md` — Non-obvious architecture decisions and rationale
- `patterns.md` — Recipes for adding widgets, API endpoints, data sources, themes

## Commands

- `make test` — Go unit tests
- `make ui-test` — Vue.js tests (vitest)
- `make lint` — golangci-lint
- `make coverage` — Check 70% coverage threshold
- `make verify` — All tests + lint + coverage + benchmark + license
- `make run` — Start Go server (debug mode, no embedded UI)
- `make run-ui` — Build UI + embed + start server
- `cd webui && npm run dev` — Vite dev server (proxies API to :8087)

## Rules

- Never edit `.golangci.yaml` (or any `.golangci.*` config) without explicit user approval. When a linter flags code, fix the code or use a targeted inline `//nolint:<linter> // reason` directive instead. Only change lint config after the user says yes.
