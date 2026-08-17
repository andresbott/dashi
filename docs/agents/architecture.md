# Architecture — layering, package roles, design decisions

Dashi (module `github.com/andresbott/dashi`) is a self-hosted dashboard
**application**: Go backend + embedded Vue 3 SPA for the admin/editor panel.
The dashboard viewer is server-rendered HTML (two output stacks: browser and
image/PNG for e-ink/ESP32). All persistence is flat files; there is no database.

Agent docs: [features.md](features.md) · [widgets.md](widgets.md) ·
[viewer.md](viewer.md) · [rendering.md](rendering.md) · [testing.md](testing.md) ·
[releasing.md](releasing.md)

## Layering

    app/cmd (Cobra CLI: start, config, theme, version)
        |
    app/router (viewer + editor mux routers, middleware)
        |            \
    app/router/handlers (API v0)   app/spa (go:embed'd Vue build for editor)
        |
    internal/dashboard (types + file store)   internal/themes   internal/widgets/*
        |
    internal/dashboard/browser (HTML for browsers)  internal/dashboard/static (HTML)
        |                                              |
    widgets.BrowserRenderable                      internal/dashboard/image (PNG via litehtml)
        |
    lib/einkimage (standalone dithering library)

- **`internal/dashboard/types.go` is the domain core**: `Dashboard` → `Pages` →
  `Rows` → `Widgets`, widget `Config` is opaque `json.RawMessage`. `Auth` is a
  per-dashboard `auth.json` sidecar (bcrypt hash).
- **`internal/dashboard.Store` is pure file persistence**: one folder per
  dashboard under `{dataDir}/dashboards/{snake_name}/` holding `dashboard.json`
  plus optional sidecars (`custom.css`, `auth.json`, `assets/`, `md/`). An
  in-memory `id → folder` index is built at startup. IDs are random 6-char
  lowercase alphanumerics; folder names are snake_case of the dashboard name —
  independent of the ID, so renames don't move files.
- **`internal/widgets` is the backend widget registry** — see
  [widgets.md](widgets.md). Registration happens in `app/router/main.go`
  (`newSharedDeps`).
- **`lib/` vs `internal/`**: `lib/einkimage` is a self-contained, reusable
  library (own README, no dashi imports). Only promote code to `lib/` when it
  has zero application coupling.

## Two servers, one binary (decided 2026-04-07, PR #12)

`app/cmd/server.go` starts up to three HTTP servers via errgroup:

| Server | Default | Role |
|---|---|---|
| viewer | `:8087` | read-only: `attachReadAPIs` (GET only) + dashboard-ID SPA routes |
| editor | `:8088` | full CRUD: `attachReadAPIs` + `attachWriteAPIs` + full SPA |
| observability | `:9090`, **disabled** | placeholder — handler is still `nil` (TODO in server.go) |

Both share one `sharedDeps` (stores, clients, renderers) built once in
`app/router/main.go` — **do not construct a second `dashboard.Store`**; the
in-memory index would go stale across instances. The read/write split lives in
`app/router/api_v0.go`: never mount `attachWriteAPIs` on the viewer. At least
one of viewer/editor must be enabled (validated in `app/cmd/config.go`).

## Security model — know the boundaries

- **Per-dashboard basic auth** (2026-04-08, PR #14): `NewDashboardAuthMiddleware`
  in `app/router/middleware_auth.go` guards GET requests **on the viewer only**,
  for both `/{id}` and `/api/v0/dashboards/{id}...` paths. Credentials are
  bcrypt-checked with constant-time username compare. The **editor is
  intentionally unprotected** — it is meant to be bound to a trusted interface;
  it only hosts the auth CRUD endpoints (`/dashboards/{id}/auth`).
- **Widget config is a pass-through**: the backend does not validate widget
  config JSON. A stored-XSS gap via the bookmark widget's `href` is catalogued
  in `docs/TODO.md` with chosen fix directions — read it before touching config
  handling; do not invent a different validation scheme.
- **Asset uploads**: 10MB cap, extension allowlist (`.png .jpg .jpeg .svg
  .webp .css .md`), path-traversal rejection — all enforced in
  `internal/dashboard/store.go` (`validateAssetPath`, `isAllowedAssetExt`).
  Keep validation in the store, not in handlers.
- The `Auth` block in `app/cmd/config.go` (Enabled/DefaultUser) is **config
  scaffolding only** — nothing consumes it yet. Don't document it as a feature.

## Configuration

Loaded via `go-bumbu/config` in order (last wins): built-in defaults → `.env`
→ `config.yaml` → env vars prefixed `DASHI_`. `dashi config` writes a
commented default `config.yaml` (refuses to overwrite; `-o` to change path).
Note: the README Quick Start still says `dashi generate-config > config.yaml` —
that command name and stdout redirection are stale; the code is the truth.

## Key decisions (why the code is shaped this way)

- **Files, no DB** (inception): backup = copy a folder; dashboards are
  hand-editable and diffable. Don't introduce a database for new features.
- **Opaque widget config** (inception): each widget owns its schema on both
  sides; no central validation layer. Tradeoff accepted — see security note.
- **Widget clients and widgets are separated** (current): clients live at
  `internal/{name}/` (`internal/weather/`, `internal/market/`,
  `internal/swisstransport/`, `internal/sysinfo/`, `internal/xkcd/`); widget
  packages at `internal/widgets/{name}/` hold module, renderers, templates and
  handler. A 2026-05-15 merge (commits `9d795bf`/`0b6e7fc`) folded clients
  into widget packages; that has been reverted. Older docs predating the
  revert are stale on this.
- **Embedded SPA** (inception): `make package-ui` copies `webui/dist/` into
  `app/spa/files/ui/` for `go:embed`. Those files are build artifacts — never
  hand-edit them.
- **Startup warmup**: `warmupWeather`/`warmupMarket` (app/router/main.go) scan
  all dashboard configs and pre-fetch data so first render is warm. A new
  data-backed widget that should be warm at boot needs its own warmup pass.

## Historical docs — what wins

`docs/autodoc/` (overview/features/decisions/patterns) predates the 2026-05
refactors and the xkcd/transport/sysinfo/stack/markdown/image widgets: treat it
as historical record; **this `docs/agents/` set and the code win**.
`docs/project/` holds analysis docs (widget rendering architectures, refactor
ideas) and `docs/superpowers/` holds dated specs/plans — good for the *why*,
not for current paths.

## Known debt

- Root `TODO.md`: HA-addon migration, enhanced dithering, theme rework, e-ink
  preview dialog, migrate to user-data dir. Check here before redesigning any
  of these — the direction is chosen. Note: "widgets as isolated components"
  and "remove the preview feature" have shipped; they remain in the file as
  historical record.
- `docs/TODO.md`: the bookmark-widget stored-XSS item (see security model).
- Observability server: enabled flag exists, handler is `nil`.
- **Browser-stack widget port is mid-flight**: clock and sysinfo are ported;
  the remaining twelve render through their image renderers when viewed in a
  browser. See [viewer.md](viewer.md).
