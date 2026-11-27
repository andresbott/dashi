# Architecture — layering, package roles, design decisions

Dashi (module `github.com/andresbott/dashi`) is a self-hosted dashboard
**application**: Go backend + embedded Vue 3 SPA for the admin/editor panel.
The dashboard viewer is server-rendered HTML (two output stacks: browser and
image/PNG for e-ink/ESP32). All persistence is flat files; there is no database.

Agent docs: [features.md](features.md) · [widgets.md](widgets.md) ·
[viewer.md](viewer.md) · [rendering.md](rendering.md) · [testing.md](testing.md) ·
[releasing.md](releasing.md)

## Layering

    app/cmd (Cobra CLI: start, config, widget-theme, version)
        |
    app/router (viewer + editor mux routers, middleware)
        |            \
    app/router/handlers (API v0)   app/spa (go:embed'd Vue build for editor)
        |
    internal/dashboard (types + file store)   internal/themes   internal/widgetthemes
    internal/widgets/*
        |                                                               |
        |                              internal/providers/* (data-source clients + cache)
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

Because the two live on different ports, the SPA cannot guess where the public
viewer is: `GET /api/v0/info` (`app/router/handlers/info.go`) advertises both
servers as `{"viewer":{"enabled":…,"url":…},"editor":{"enabled":…,"url":…}}`.
The viewer URL is `Server.Viewer.PublicUrl` when set (used verbatim), otherwise
the request host with the viewer port swapped in — which covers both `vite dev`
and a plain two-port deployment. Only http/https ever come out of it: the value
lands in an SPA `href`, so a forged `Host`/`X-Forwarded-*` must not be able to
smuggle another scheme in. The SPA loads it once at init
(`webui/src/lib/serverInfo.ts`) and falls back to same-origin `/{id}` paths.

The editor URL is `Server.Editor.PublicUrl` and is **never derived** from the
request: behind an ingress (HA add-on: viewer on a host port, editor under
`<HA URL>/hassio/ingress/<slug>`) the request cannot reveal it. It is empty
when unset or when the editor is disabled (`router.publicEditor` in
`app/router/main.go`). The same value is handed to the dashboard middleware
for the viewer's Edit button — see [viewer.md](viewer.md#edit-button).

## Security model — know the boundaries

- **Per-dashboard basic auth** (2026-04-08, PR #14): `NewDashboardAuthMiddleware`
  in `app/router/middleware_auth.go` guards GET requests **on the viewer only**,
  for both `/{id}` and `/api/v0/dashboards/{id}...` paths. Credentials are
  bcrypt-checked with constant-time username compare. The **editor is
  intentionally unprotected** — it is meant to be bound to a trusted interface;
  it only hosts the auth CRUD endpoints (`/dashboards/{id}/auth`). The
  optional viewer Edit button links to `Server.Editor.PublicUrl`, i.e. it
  publishes the editor's address to every viewer of that dashboard — so that
  URL should only point at an editor behind an authenticating proxy (HA
  ingress is one).
- **Widget config is opaque, with opt-in per-widget validation** (decided
  2026-11-21, closing the bookmark stored-XSS item): a widget module may
  implement `widgets.ConfigValidator`; `newSharedDeps` registers it on the
  `widgets.Registry`, and hands the registry to the store via
  `dashboard.Store.SetWidgetValidator`. `Store.Create`, `Update` and
  `ImportZip` validate every widget before touching disk and return
  `dashboard.ErrInvalidWidgetConfig` (wrapping the widget ID, page and reason),
  which the handlers turn into a **400**; the editor's save toast shows the
  reason. The store stays decoupled: it only knows the small `WidgetValidator`
  interface. Types without a validator, and unknown types, pass through. Only
  `bookmark` and `stack` (validates children, recursively) have rules today.
  This is defence in depth, three layers for bookmark URLs:
  1. **Save**: `bookmark.validateURL` — allowed are empty, `http://`,
     `https://` (case-insensitive) or a single-`/` path; rejected are every
     other scheme, `//host`, `/\host`, scheme-less text and any ASCII control
     character. The case table `internal/widgets/bookmark/testdata/url_cases.json`
     is shared by the Go and vitest tests.
  2. **Render**: the bookmark template gets `"#"` for any stored URL failing
     the rule (dashboards saved before the rule, or edited on disk, are not
     re-validated on read), and html/template's URL filter (`#ZgotmplZ`) sits
     under that.
  3. **Editor**: `webui/src/widgets/bookmark/url.ts` mirrors the rule; the
     form shows an inline error and emits nothing while the URL fails it, and
     the editor preview drops the `href`.

  Validation is a write gate, not a read guarantee: a renderer must keep
  treating stored config as untrusted. Rules belong in the widget package —
  don't add a central validation layer. Writes that go around the API (the
  store's own default-flag rewrite, hand edits) are not validated.
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

Public URLs (`Server.Viewer.PublicUrl`, `Server.Editor.PublicUrl`) are
normalized **once at config load** (`normalizePublicURLs` in
`app/cmd/config.go`): trimmed, trailing slash stripped, and anything that is
not an absolute http/https URL dropped with a `warn` message that
`app/cmd/server.go` logs at warn level. Consumers (info handler, edit link)
use the values verbatim — don't re-validate downstream, and don't add a public
URL without adding it to that list.
Note: the README Quick Start still says `dashi generate-config > config.yaml` —
that command name and stdout redirection are stale; the code is the truth.

## Key decisions (why the code is shaped this way)

- **Files, no DB** (inception): backup = copy a folder; dashboards are
  hand-editable and diffable. Don't introduce a database for new features.
- **Opaque widget config** (inception): each widget owns its schema on both
  sides; no central validation layer. Since 2026-11-21 a widget can opt into
  save-time validation of its own config — see the security note.
- **Widget clients and widgets are separated** (current): clients live at
  `internal/providers/{name}/` (`internal/providers/weather/`,
  `.../market/`, `.../swisstransport/`, `.../sysinfo/`, `.../xkcd/`); widget
  packages at `internal/widgets/{name}/` hold module, renderers, templates and
  handler. A 2026-05-15 merge (commits `9d795bf`/`0b6e7fc`) folded clients
  into widget packages; that has been reverted. Since 2026-08-17 the clients
  are grouped under `internal/providers/` (they used to sit directly at
  `internal/{name}/`). Older docs predating these moves are stale on this.
- **Widgets are self-contained up to the language boundary** (2026-11-06):
  server-shipped parts live in `internal/widgets/{name}/`, editor forms stay in
  `webui/src/widgets/{name}/`, and planned tests keep them in agreement. This
  overrides D6 of the 2026-08-17 viewer spec; see
  [widgets.md](widgets.md#widget-boundary-decided-2026-11-06).
- **Embedded SPA** (inception): `make package-ui` copies `webui/dist/` into
  `app/spa/files/ui/` for `go:embed`. Those files are build artifacts — never
  hand-edit them.
- **Startup warmup**: `newSharedDeps` (app/router/main.go) collects every saved
  config of each widget type and passes it to that module's `Warmup`, which
  pre-fetches data so first render is warm. A new data-backed widget that should
  be warm at boot implements `Warmup`.

## Historical docs — what wins

`docs/autodoc/` (overview/features/decisions/patterns) predates the 2026-05
refactors and the xkcd/transport/sysinfo/stack/markdown/image widgets: treat it
as historical record; **this `docs/agents/` set and the code win**.
`docs/project/` holds analysis docs (widget rendering architectures, refactor
ideas) and `docs/superpowers/` holds dated specs/plans — good for the *why*,
not for current paths.

## Known debt

- Root `TODO.md`: HA-addon migration, migrate to user-data dir. Check here
  before redesigning any of these — the direction is chosen. Note: "widgets
  as isolated components", "remove the preview feature", "enhanced dither"
  and the "e-ink preview dialog" have shipped (2026-11-21 for the last two,
  see [rendering.md](rendering.md); the dialog has since become the
  `/dashboards/:id/eink-preview` page); they remain in the file as
  historical record.
- Error bodies on both servers are double-wrapped: the go-bumbu middleware
  (`JsonErrors: true`) takes the handler's `{"error":"..."}` body as the
  message string of its own `{"error": ...}`. The SPA's `apiErrorMessage`
  (`webui/src/lib/api/client.ts`) unwraps both shapes; older SPA call sites
  that read `response.data.error` directly get the inner JSON as text.
- Observability server: enabled flag exists, handler is `nil`.
- **Browser-stack widget port is mid-flight**: clock and sysinfo are ported;
  the remaining twelve render through their image renderers when viewed in a
  browser. See [viewer.md](viewer.md). Whether all twelve need porting is open;
  see [widgets.md](widgets.md#open-questions).
- **Widget config drift**: a [contract test](widgets.md#config-contract) now
  fails when form and Go config fields disagree. Sysinfo's unused form options
  remain a recorded exception, and clock's defaults still differ, which that test
  doesn't check; see
  [widgets.md](widgets.md#known-mismatches-verified-2026-11-06).
