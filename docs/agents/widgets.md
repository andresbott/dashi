# Widgets — registries, two render stacks, how to add one

Every widget renders in independent pipelines that share only the type string
and the config JSON (see [viewer.md](viewer.md), [rendering.md](rendering.md)).
The backend has two render stacks (browser + image/e-ink); the frontend has a
Vue component registry for the admin/editor panel. The invariant that keeps
them in sync is **conventional, not enforced**: the backend registry keys in
`app/router/main.go` and the frontend registry key in
`webui/src/lib/widgetRegistry.ts` must be the identical string, and all sides
must parse the same config shape. Nothing will fail at compile time if they
drift — an unknown type just renders nothing (no error). The planned tests that
close this are listed under [Widget boundary](#widget-boundary-decided-2026-11-06).

## The registries

- **Backend** (`internal/widgets/registry.go`): `Registry` holds two maps —
  `renderers` (image stack, for e-ink) and `browser` (browser stack). Both map
  type → `StaticRenderer func(config json.RawMessage, ctx RenderContext)
  (template.HTML, error)`. `Registry.RenderBrowser` falls back to the image
  renderer when a widget has not been ported to the browser stack yet (see
  [viewer.md](viewer.md)). Unknown types render nothing (empty output), **not
  an error** — a misspelled type string fails silently by design.
- **Frontend** (`webui/src/lib/widgetRegistry.ts`): a hand-kept list importing
  each `webui/src/widgets/{name}/index.ts`, giving type → `{component,
  configComponent | null, label, icon, description, noWidgetProp?}`. Components
  are `defineAsyncComponent` — each widget is its own chunk. This registry is
  for the admin/editor SPA only; the viewer is server-rendered (see
  [viewer.md](viewer.md)).

## Row placement

Widgets carry a `Width` (1–12 column span) and an optional `Column` (1-based
start column; `0`/absent = flow after the previous widget, which is what
dashboards created before column support carry). Placement is resolved once by
`dashboard.PlaceRow` (`internal/dashboard/layout.go`), shared by both render
stacks and mirrored in the editor (`webui/src/lib/rowLayout.ts`). It converts
each column into a leading gap and pushes overlaps right, so a widget can sit in
any column and adding widgets can never overlap.
See [rendering.md](rendering.md) for the gap → `margin-left`/`col-offset`
mechanism.

## Current matrix (verified 2026-07-22)

| Type | Static (Go) | Interactive (Vue) | Config UI | Notes |
|---|---|---|---|---|
| weather / weather-compact | yes | yes | yes | shares `weather` package; two renderers |
| bookmark | yes | yes | yes | XSS gap in `href` — see `docs/TODO.md` |
| clock | yes | yes | yes | |
| battery | yes | yes | no | reads `%` from query param |
| page-indicator | yes | yes | no | `noWidgetProp: true` on frontend |
| market | yes | yes | yes | |
| xkcd | yes | yes | yes | disk-cached comics |
| transport | yes | yes | yes | Swiss transport; package `swisstransport` |
| sysinfo | yes | yes | yes | gopsutil, host metrics |
| stack | yes | yes | no | container: renders children via the registry itself (`NewStaticRenderer(registry)`) |
| markdown | yes | yes | yes | reads `.md` from dashboard `md/` dir via `dashboard.Store` (spec: `docs/superpowers/specs/2026-04-16-markdown-widget-design.md`) |
| image | yes | yes | yes | serves uploaded dashboard assets |
| search | **no** | yes | yes | deliberately interactive-only |

## Package layout rule (current)

Client packages and widget packages are separated:

- **Clients + cache:** `internal/providers/{name}/` —
  `internal/providers/weather/`, `.../market/`, `.../swisstransport/`,
  `.../sysinfo/`, `.../xkcd/`
- **Widget packages:** `internal/widgets/{name}/` — module, renderers (image +
  browser), templates, handler

A 2026-05-15 merge (commits `9d795bf`, `0b6e7fc`) folded clients into widget
packages; that has been reverted. Docs written between 2026-05-15 and 2026-08
use the merged-layout paths and are stale. On 2026-08-17 the client packages
moved from `internal/{name}/` into the `internal/providers/` group; only the
import paths changed (package names are unchanged).

## Widget boundary (decided 2026-11-06)

**A widget is self-contained up to the language boundary.** Everything the
server ships lives in the Go widget folder; editor forms stay in `webui/`; tests
keep the two sides in agreement. This overrides D6 and the "move `Config.vue`
and `types.ts` into widget folders" step of plan 03 in
`docs/superpowers/specs/2026-08-17-vanilla-viewer-architecture.md`. Don't
re-propose full colocation unless a trigger below applies.

| Part | Lives in |
|---|---|
| Module, image and browser renderers, templates, CSS, viewer JS, handler, tests | `internal/widgets/{name}/` |
| Editor metadata (`index.ts`), `WidgetConfig.vue`, `types.ts`, editor API helpers | `webui/src/widgets/{name}/` |
| Data client and cache | `internal/providers/{name}/` |

### Why

- `go:embed` forces server assets next to the Go code. Nothing forces the forms
  there: only Vite consumes them.
- Colocation fixes navigation, not drift. Type strings match on both sides; the
  mismatches are in config fields (see below), and moving files doesn't catch
  those.
- The folder still wouldn't be self-contained. All 11 forms need PrimeVue from
  `webui/node_modules`, and clock, bookmark and weather-compact import shared
  SPA code through `@/`, so `internal/` would depend on the SPA.
- Moving forms is technically possible (tried on a scratch copy) but needs
  permanent Vite, TypeScript and Vitest config.
- Dashi has no third-party widgets, which is what drives colocation in most
  comparable projects. WordPress colocates even its core blocks; Grafana and
  Perses keep languages apart inside each plugin; Home Assistant splits backend
  and frontend across repos.

A generic schema-driven form was also rejected: by a size and import estimate,
5 of the 11 forms (about 67% of the lines) would need a custom escape hatch.

### Planned, not built yet

1. Replace the hand-kept list in `widgetRegistry.ts` with `import.meta.glob`
   over `webui/src/widgets/*/index.ts`.
2. Add a test that Go module types match the frontend types. Compare type
   strings, not folder names: `swisstransport` and `pageindicator` differ from
   their types, and the `weather` package hosts two.
3. Add a test that form fields and Go config fields match, then fix the known
   mismatches below.
4. Spike a per-widget iframe preview in the editor before deleting any
   `Widget.vue` (see open questions).

Go-served widget metadata (spec plan 03) is **deferred**: only the editor uses
that metadata, and the tests above close the drift for less work.

### Known mismatches (verified 2026-11-06)

- **sysinfo:** the form offers 10 options, but `internal/widgets/sysinfo/` only
  reads `showMemory`, `showUptime` and `disks`. The other 7 do nothing in the
  viewer.
- **clock:** with no saved config, `Widget.vue` shows seconds and date, while
  the Go config defaults them to off.
- **search:** the Go renderer outputs a blank placeholder, so it shows nothing
  in the viewer.
- **stack:** `stack.go` renders children through `registry.Render`, so ported
  children still use their image templates inside a stack.
- **markdown:** `AdminNotes.vue` and `useAdminNotes.ts` import
  `widgets/markdown/api`, so admin code depends on a widget. Move it to
  `lib/api`, as the widget boundary implies.

### Open questions

- **Can `Widget.vue` be deleted?** Spec plan 03 replaces the 14 files with a
  server-rendered preview, but two things stand in the way:
  - Stack's drag and child editing live in its `Widget.vue` and the spec doesn't
    cover them. They need a home, so a Vue registry stays regardless.
  - Viewer scripts start once per page, so a re-mounted clock probably stops
    ticking, and `viewer.css` may collide with the editor's styles (both
    inferred, not run). An iframe per widget should solve both, at the cost of
    height syncing and drag handling. If the spike fails, keep the Vue versions
    as editor-only previews.
- **Which widgets need a browser port?** Weather, weather-compact, market,
  transport and xkcd need a template and script; search needs a template; stack
  needs a small change. Battery, page-indicator, image, bookmark and markdown
  could stay on the image-template fallback, with plainer styling. Not decided.

### What would reopen this

- Installable third-party widgets (favours colocation).
- One-widget changes that keep missing one side despite the field-parity test
  (favours colocation).
- A small set of field types that shrinks custom forms to about two (favours
  schema-driven forms).

## Adding a widget — the full checklist

1. **Widget package:** `internal/widgets/{name}/` with `module.go`
   (`widgets.Module` wiring), `image.go` + `image.html` (e-ink stack), and
   optionally `browser.go` + `browser.html` + `{name}.css` + `{name}.js`
   (browser stack). Tests: `image_test.go`, `browser_test.go` (if ported),
   `module_test.go`. If the widget ships CSS, add a scoping test enforcing all
   selectors descend from `.dashi-{type}` (pattern:
   `internal/widgets/sysinfo/browser_test.go:83-92`).
2. **Client package (if needed):** `internal/providers/{name}/` with client + cache +
   `client_test.go`. Inject `*http.Client` so tests don't hit real APIs.
3. **Register:** in `app/router/main.go` (`newSharedDeps`), add the module to
   the `modules` slice. The loop registers both the image renderer
   (`registry.Register`) and, if the module implements
   `widgets.BrowserRenderable`, the browser renderer (`registry.RegisterBrowser`).
4. **API endpoint (if interactive):** handler in
   `internal/widgets/{name}/handler.go`, mounted by the module's
   `RegisterRoutes`, which `attachReadAPIs` (`app/router/api_v0.go`) calls for
   every module. Those routes are read-only because both servers mount them;
   write routes go in `attachWriteAPIs` only — never expose writes on the
   viewer.
5. **Frontend (editor/admin):** `webui/src/widgets/{type}/` with `index.ts`,
   `Widget.vue`, and `WidgetConfig.vue` + `types.ts` if configurable; add
   `api.ts` + `composable.ts` if it fetches data (TanStack Vue Query). Register
   in `webui/src/lib/widgetRegistry.ts` with the **same type string**.
6. **Warmup (if data-backed):** implement `Module.Warmup`, which `newSharedDeps`
   calls with every saved config of that type, so first render is warm. Widgets
   with nothing to warm embed `widgets.NoopModule`.

Naming: type string is lowercase-kebab (`weather-compact`); Go package is the
type without hyphens; Vue components are PascalCase. Browser assets are served
at `/_dashi/widgets/{type}.js` — see [viewer.md](viewer.md).

## Image-renderer constraints (break these and PNGs render wrong)

- Output must be **self-contained HTML**: inline styles, base64 data-URI
  images. litehtml fetches nothing over the network and ignores external CSS.
- CSS support is litehtml's subset: **no CSS Grid**
  (`docs/project/litehtml-rendering-reference.md:268`), **no CSS `var()` custom
  properties** (`:270`), no animations or transitions (`:269`). **Flexbox is
  supported** (`:143` — comprehensive: `display:flex`, direction, wrap,
  `justify-content`, `align-items`/`self`/`content`, `order`, `flex-basis/grow/
  shrink`, `margin:auto` centring) and is the recommended layout tool alongside
  floats; the limitations list only begins at `:193`. Test against
  `internal/dashboard/image` rendering, not a browser.
  This is why `RenderContext.Palette` passes concrete hex values into the image
  stack while the browser stack gets CSS variables — see [viewer.md](viewer.md).
- `template.HTML` returns are exempt from gosec G203 in `internal/widgets/`
  (`.golangci.yaml`) because templates are trusted — do not pass user input
  unescaped into that HTML.
- Renderers get data via `widgets.RenderContext` (DashboardID, Theme,
  ColorMode, Palette, QueryParams, PageIndex, TotalPages) and their injected
  client — they must not do per-request slow fetches; hit the package cache.
