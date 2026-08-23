# Widgets — registries, two render stacks, how to add one

Every widget renders in independent pipelines that share only the type string
and the config JSON (see [viewer.md](viewer.md), [rendering.md](rendering.md)).
The backend has two render stacks (browser + image/e-ink); the frontend has a
Vue component registry for the admin/editor panel. The invariant that keeps
them in sync is **conventional, not enforced**: the backend registry keys in
`app/router/main.go` and the frontend registry key in
`webui/src/lib/widgetRegistry.ts` must be the identical string, and all sides
must parse the same config shape. Nothing will fail at compile time if they
drift — an unknown type just renders nothing (no error).

## The registries

- **Backend** (`internal/widgets/registry.go`): `Registry` holds two maps —
  `renderers` (image stack, for e-ink) and `browser` (browser stack). Both map
  type → `StaticRenderer func(config json.RawMessage, ctx RenderContext)
  (template.HTML, error)`. `Registry.RenderBrowser` falls back to the image
  renderer when a widget has not been ported to the browser stack yet (see
  [viewer.md](viewer.md)). Unknown types render nothing (empty output), **not
  an error** — a misspelled type string fails silently by design.
- **Frontend** (`webui/src/lib/widgetRegistry.ts`): type → `{component,
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

## Adding a widget — the full checklist

1. **Widget package:** `internal/widgets/{name}/` with `module.go`
   (`widgets.Module` wiring), `image.go` + `image.html` (e-ink stack), and
   optionally `browser.go` + `browser.html` + `{name}.css` + `{name}.js`
   (browser stack). Tests: `image_test.go`, `browser_test.go` (if ported),
   `module_test.go`. If the widget ships CSS, add a scoping test enforcing all
   selectors descend from `.dashi-{type}` (pattern:
   `internal/widgets/sysinfo/browser_test.go:83-92`).
2. **Client package (if needed):** `internal/{name}/` with client + cache +
   `client_test.go`. Inject `*http.Client` so tests don't hit real APIs.
3. **Register:** in `app/router/main.go` (`newSharedDeps`), add the module to
   the `modules` slice. The loop registers both the image renderer
   (`registry.Register`) and, if the module implements
   `widgets.BrowserRenderable`, the browser renderer (`registry.RegisterBrowser`).
4. **API endpoint (if interactive):** handler in
   `app/router/handlers/{name}.go`, route in `attachReadAPIs`
   (`app/router/api_v0.go`). Write routes go in `attachWriteAPIs` only — never
   expose writes on the viewer.
5. **Frontend (editor/admin):** `webui/src/components/dashboards/{Name}Widget.vue`
   (+ `{Name}WidgetConfig.vue` if configurable), TS types in
   `webui/src/types/`, composable in `webui/src/composables/` if it fetches
   data (TanStack Vue Query). Register in `webui/src/lib/widgetRegistry.ts`
   with the **same type string**.
6. **Warmup (if data-backed):** add a warmup pass in `app/router/main.go`
   (pattern: `warmupWeather`/`warmupMarket`), so first render is warm.

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
