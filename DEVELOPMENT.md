# Developing Dashi

## Overview

Dashi is a Go server with a Vue 3 editor, shipped as one binary with the editor
UI embedded.

- `main.go`, `app/`: CLI commands (Cobra), routers, handlers and wiring.
- `internal/`: dashboards, widgets, themes, data providers and the rendering stacks.
- `lib/einkimage/`: the e-ink dithering library.
- `webui/`: the Vue editor (PrimeVue, TanStack Query, Vite) and the viewer's Tailwind stylesheet.
- `ha-addon/`, `repository.yaml`: the Home Assistant add-on.

Architecture, widget, rendering and testing notes live in
[docs/agents/](docs/agents/architecture.md); start with `architecture.md`.

## Prerequisites

- Go 1.25.7 (from `go.mod`) and a C compiler: CGO is required for litehtml-go.
- Node.js 24 (the version CI uses).
- [golangci-lint](https://github.com/golangci/golangci-lint) for `make lint`.
- [go-licence-detector](https://github.com/elastic/go-licence-detector) for `make license-check`.
- [GoReleaser](https://goreleaser.com) v2 for `make build`.

## Common tasks

```sh
make help           # list all targets
make test           # Go unit tests
make ui-test        # editor type check (vue-tsc) and unit tests (vitest)
make lint           # Go linter
make ui-lint        # ESLint for webui/
make coverage       # every package under internal/ must reach 70% coverage
make verify         # full check: all of the above plus benchmarks and license check
```

Run `make verify` before opening a PR. Testing conventions and lint policy:
[docs/agents/testing.md](docs/agents/testing.md).

## Run locally

```sh
make run            # start the server in debug mode
make run-ui         # rebuild the editor UI and viewer stylesheet, then start the server
```

The editor listens on <http://localhost:8088> and the viewer on
<http://localhost:8087>. Both use built-in defaults and store data in `./data`
(gitignored, starts empty). Add a `config.yaml` (`go run main.go config`) or a
`.env` (copy `.env.example`) to change settings.

`make run` serves the editor UI last packaged into `app/spa/files/ui/`. For
editor work, run the Vite dev server with hot reload next to `make run`; it
proxies API calls to the editor on `:8088`:

```sh
cd webui && npm install && npm run dev
```

After changing a viewer template (`page.html` or a widget's `browser.html`),
rebuild the viewer stylesheet with `make viewer-css`. The embedded UI and
`internal/dashboard/browser/assets/viewer.css` are generated: never edit them
by hand.

## Build

```sh
make build          # binary for the current OS/arch, editor UI included
make clean          # remove dist/
```

`make build` runs `make package-ui` and `make viewer-css`, then a GoReleaser
snapshot build into `dist/`.

## Release

1. Bump `version:` in `ha-addon/config.yaml` to the new version, without the `v`.
2. Run `make verify`.
3. On a clean `main`, run:

   ```sh
   make tag version="v1.2.3"
   ```

`make tag` refuses uncommitted changes to tracked files, any branch but
`main`, or a tag that doesn't match the add-on version. Pushing the tag starts the Release workflow, which
publishes Linux, Windows and macOS binaries and a `.deb` to GitHub Releases,
and the add-on images to `ghcr.io/andresbott/dashi-{amd64,aarch64}`. Details
and traps: [docs/agents/releasing.md](docs/agents/releasing.md).
