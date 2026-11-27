# Widgets — registries, two render stacks, how to add one

Every widget renders in independent pipelines that share only the type string
and the config JSON (see [viewer.md](viewer.md), [rendering.md](rendering.md)).
The backend has two render stacks (browser + image/e-ink); the frontend has a
Vue component registry for the admin/editor panel. The invariant that keeps
them in sync is **not checked by the compiler**: the backend registry keys in
`app/router/main.go` and the frontend registry key in
`webui/src/widgets/*/index.ts` must be the identical string, and all sides must
parse the same config shape. Nothing fails at compile time if they drift, and an
unknown type just renders nothing (no error), so tests enforce it instead:
`internal/widgets/testdata/widget_contract.json` lists every type, and
`app/router/widget_contract_test.go` and `webui/src/lib/widgetRegistry.test.ts`
check their side against it. Compare type strings, not folder names:
`swisstransport` and `pageindicator` differ from their types, and the `weather`
package hosts two.

## The registries

- **Backend** (`internal/widgets/registry.go`): `Registry` holds two maps —
  `renderers` (image stack, for e-ink) and `browser` (browser stack). Both map
  type → `StaticRenderer func(config json.RawMessage, ctx RenderContext)
  (template.HTML, error)`. `Registry.RenderBrowser` falls back to the image
  renderer when a widget has not been ported to the browser stack yet (see
  [viewer.md](viewer.md)). Unknown types render nothing (empty output), **not
  an error** — a misspelled type string fails silently by design.
- **Frontend** (`webui/src/lib/widgetRegistry.ts`): discovers every
  `webui/src/widgets/{folder}/index.ts` with `import.meta.glob`, giving type →
  `{component, configComponent | null, label, icon, description, category,
  noWidgetProp?, settingsSize?}` (the `WidgetModule` type in
  `webui/src/widgets/types.ts`), ordered by label. There is no list to edit.
  `category` is one of the ids in `webui/src/widgets/categories.ts` (Time &
  date, Weather, Information, Navigation, Layout); `getWidgetsByCategory()`
  groups by it, and the widget docs index uses that grouping. A widget with a
  `wide` form is shown with its settings under the preview on its docs page.
  `settingsSize: 'wide'` opens the Widget Settings dialog
  (`components/dashboards/WidgetSettingsDialog.vue`) at 60rem instead of the
  default 28rem, for a form laid out in columns; only `weather` uses it. Its
  form drops to one column below the `md` breakpoint, and the global
  `.p-dialog` max-width keeps the dialog on screen. The dialog's actions sit
  in the PrimeVue footer, so they stay visible while a long form scrolls. Each folder holds
  one module, which is why `weather-compact` has its own folder and imports from
  `weather`. Components are `defineAsyncComponent` — each widget is its own
  chunk, so an `index.ts` must not import a component directly. This registry is
  for the admin/editor SPA only; the viewer is server-rendered (see
  [viewer.md](viewer.md)).

## Widget documentation pages

`/admin/docs/widgets` (`webui/src/views/docs/DocWidgets.vue`) lists the
widgets as cards grouped by `category`; each card links to
`/admin/docs/widgets/:type` (`DocWidget.vue`): summary, then an **Overview**
tab (the widget's real `Widget.vue` next to its real `WidgetConfig.vue`, wired
so form changes re-render the preview) and a **Details** tab (parameter
table).

- **Docs metadata:** `webui/src/widgets/{folder}/docs.ts` exports a
  `WidgetDocs` (`webui/src/lib/widgetDocs.ts`): `summary`, `params` (name,
  type, default, description, optional group), optional `sections`/`notes`,
  and `preview` (sample `config`, optional `height`, `provide`d injections,
  `hideForm` + `presets` for forms that write user data, as markdown's does).
  `configType` names the contract key when a widget shares another's config
  (`weather-compact` → `weather`). `webui/src/lib/widgetDocs.test.ts` fails if
  a widget has no docs, a contract field is undocumented, or a documented
  field does not exist.
- **Mocked data:** data-backed widgets also have `docFixture.ts` with canned
  payloads keyed by TanStack query-key prefix (`['weather', …]`), passed as
  `docs.mocks`. The doc pages provide their own QueryClient
  (`webui/src/lib/docQueryClient.ts`) to the preview and form subtree: it
  answers mocked prefixes from the fixtures (no refetch) and runs the real
  queryFn for anything else (widget themes, icon lookups on the local API).
  Mocks from all widgets are merged, so stack children get data too.
  Fixtures can drift from the real API shape; they are typed against the
  widget's `types.ts`, which catches renamed fields but not new ones.
- **Image** reads its URL resolver from the `DATA_IMAGE_URL` injection (docs
  provide inline sample SVGs); the default is the data API.

## Config validation

Widget config is opaque to the store, but a module can reject its own config
before it is saved by implementing the optional `widgets.ConfigValidator`
(`ValidateConfig(config json.RawMessage) error`, in
`internal/widgets/registry.go`). The `newSharedDeps` loop registers it with
`registry.RegisterValidator`, and the store calls `Registry.ValidateConfig` for
every widget on create, update and zip import; a failure is a 400 naming the
widget and the reason (details in
[architecture.md](architecture.md#security-model--know-the-boundaries)).

- Validate only what is dangerous or unrenderable. A validator that rejects
  config older dashboards already hold makes those dashboards unsavable until
  the user fixes the widget, since every autosave sends the whole dashboard.
- Keep the rule in the widget package and reuse it in the renderer: stored
  config is never re-validated on read, so the renderer must still neutralise
  bad values (bookmark renders `#`).
- Mirror the rule in the editor form when users can type the value, and share
  the cases: bookmark's `internal/widgets/bookmark/testdata/url_cases.json` is
  read by both `bookmark_test.go` and `webui/src/widgets/bookmark/url.test.ts`.
- Containers must delegate: `stack` validates each child through the
  registry, so nesting a widget cannot skip its rule.

Current validators: `bookmark` (URL allowlist: empty, `http://`, `https://` or
a single-`/` path) and `stack`.

### Sink audit (2026-11-21)

Every place widget config reaches an `href`, `src`, `style`, `class` or a
typed `template.*` value was checked. Only bookmark's `url` could carry script
(fixed). The rest are safe: clock's `font` lands in a `style` attribute, where
html/template escapes it or replaces it with `#ZgotmplZ` (and in Vue, a
`:style` object binding); image `fit` is allowlisted and image/markdown names
go through `encodeURIComponent` and the data stores' name checks; markdown
HTML comes from goldmark with its safe defaults (raw HTML dropped, dangerous
link schemes stripped; tests in `internal/data/notes/render_test.go`); search
`engine` only selects a key of the fixed `engines` allowlist
(`internal/widgets/search/engine.go`, unknown → google) and `placeholder` is
attribute-escaped; bookmark `icon` only
feeds a CDN/asset `img src` or a CSS class. Weather's `resolveIconHTML` builds
`template.HTML` with `fmt.Sprintf` from widget-theme manifest values (not
widget config), HTML-escaped but not CSS-escaped. Theme packs are trusted
uploads, but that is the one remaining hand-built HTML in `internal/widgets`.
Re-run the audit when a widget gains a URL- or markup-shaped config field.

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

The editor additionally treats a row as **one 12-column line** (decided
2026-11-06): `clampRowToLine`, `maxResizeWidth` and `fitsInRow` in
`rowLayout.ts` keep every edit (add, resize, in-row move, cross-row drop) from
pushing a widget past column 12, and a drop that does not fit is refused with a
blocked guide. Go's `PlaceRow` is unchanged: a stored row that already
overflows still wraps in the viewer; the editor just never creates one.

Widget drag and drop (rows and stacks, decided 2026-11-27) has three layers:

- **Session** — `webui/src/lib/widgetDrag.ts`, module scope (one drag at a
  time). `DashboardRow.vue` registers each row card (toolbar included) and
  the stack widget (`widgets/stack/Widget.vue`, edit mode only) registers its
  drop zone; the component that started the drag calls `startWidgetDrag` and,
  on release, `endWidgetDrag`, which returns the move or null (Escape, no
  target, blocked row). It publishes one target in `dragState` for the row
  guides and the stack's insert marker.
- **Hit testing** — `lib/widgetDragHit.ts`, pure numbers. A stack's zone is
  its card's whole `.widget-body` and **wins over the row while the pointer
  is inside it**; the card header and the gaps around it target the row.
  Otherwise the nearest row is targeted, with a small snap margin between
  rows. A dragged stack ignores every zone (no stack in a stack).
- **Commit** — every drop, same row included, goes through `applyWidgetMove`
  (`lib/widgetMove.ts`) via the `WIDGET_MOVE` injection, which only
  `DashboardEditView.vue` provides: it assigns the page's rows once per drop
  (one autosave). Without a provider a drop does nothing.

Placement rules: a row widget keeps its width (refused unless it fits); a
widget dropped into a stack is inserted at the hovered index with the stack's
width and no column; a child dragged out onto **any** row (its stack's own
row included) gets `min(6, 12 − used columns)` and is blocked only when the
row is full — the guide shows that shrunken span. A child's stored width is
never trusted (children always render at their stack's width). SortableJS
only detects the drag and paints the floating clone; both lists are passive
(`:sort="false"`, `:move="() => false"`). Keyboard moves across rows or into
and out of stacks are not built (see [features.md](features.md)).

## Current matrix (verified 2026-07-22)

| Type | Static (Go) | Interactive (Vue) | Config UI | Notes |
|---|---|---|---|---|
| weather / weather-compact | yes | yes | yes | shares `weather` package; two renderers |
| bookmark | yes | yes | yes | URL allowlist enforced on save, render and in the form — see [Config validation](#config-validation) |
| clock | yes | yes | yes | |
| battery | yes | yes | no | reads `%` from query param |
| page-indicator | yes | yes | no | `noWidgetProp: true` on frontend |
| market | yes | yes | yes | |
| xkcd | yes | yes | yes | disk-cached comics |
| transport | yes | yes | yes | Swiss transport; `swisstransport` folder on both sides |
| sysinfo | yes | yes | yes | gopsutil, host metrics |
| stack | yes | yes | no | container: renders children via the registry itself (`NewStaticRenderer(registry)`), and validates them through it too |
| markdown | yes | yes | yes | reads `.md` from dashboard `md/` dir via `dashboard.Store` (spec: `docs/superpowers/specs/2026-04-16-markdown-widget-design.md`) |
| image | yes | yes | yes | serves uploaded dashboard assets |
| search | yes | yes | yes | browser: plain GET form, no JS; image: static bar. Engine list shared via `internal/widgets/search/testdata/engines.json` |

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

1. Spike a per-widget iframe preview in the editor before deleting any
   `Widget.vue` (see open questions).

Go-served widget metadata (spec plan 03) is **deferred**: only the editor uses
that metadata, and the tests above close the drift for less work.

### Config contract

`configs` in `internal/widgets/testdata/widget_contract.json` lists each widget's
config fields, keyed by type string. `internal/widgets/contract_test.go` compares
it to the JSON tags of every `<key>Config` struct under `internal/widgets/{package}/`,
and `webui/src/widgets/contract.test.ts` compares it to the `*WidgetConfig`
interface in each `webui/src/widgets/{folder}/types.ts`, so both naming patterns
are required.

- `fields` must match both sides.
- `editorOnly` fields exist only in the editor type. These are known gaps, not
  approved design.
- `noForm` skips the editor side. `stack` uses it, because its `widgets` hold
  child widgets rather than form values.

`weather-compact` has no entry because it shares `weatherConfig` and
`WeatherWidgetConfig`.

### Known mismatches (verified 2026-11-06)

- **sysinfo:** the form offers 10 options, but `internal/widgets/sysinfo/` only
  reads `showMemory`, `showUptime` and `disks`. The other 7 do nothing in the
  viewer. They are listed under `editorOnly` in the
  [contract](#config-contract); removing or implementing them is still open.
- **clock:** with no saved config, `Widget.vue` shows seconds and date, while
  the Go config defaults them to off.
- **stack:** `stack.go` renders children through `registry.Render`, so ported
  children still use their image templates inside a stack.

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
  transport and xkcd need a template and script (search is done); stack
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
5. **Frontend (editor/admin):** `webui/src/widgets/{folder}/` with `index.ts`,
   `Widget.vue`, and `WidgetConfig.vue` + `types.ts` if configurable; add
   `api.ts` + `composable.ts` if it fetches data (TanStack Vue Query). The
   registry finds the folder by itself; `index.ts` must declare the **same type
   string** as the Go module. Add that string to `types` in
   `internal/widgets/testdata/widget_contract.json`, and any config fields to
   `configs` (see [Config contract](#config-contract)). Give `index.ts` a
   `category`.
6. **Documentation:** add `docs.ts` (summary, every config field in `params`,
   a sample `preview.config`) and, if the widget fetches data, `docFixture.ts`
   with canned payloads for its query keys, wired as `docs.mocks` — see
   [Widget documentation pages](#widget-documentation-pages).
   `widgetDocs.test.ts` enforces the field list.
7. **Warmup (if data-backed):** implement `Module.Warmup`, which `newSharedDeps`
   calls with every saved config of that type, so first render is warm. Widgets
   with nothing to warm embed `widgets.NoopModule`.
8. **Validation (if config reaches a URL, markup or script sink):** implement
   `widgets.ConfigValidator` and neutralise the same values in the renderer —
   see [Config validation](#config-validation).

Naming: type string is lowercase-kebab (`weather-compact`); Go package is the
type without hyphens; the webui folder is the type string; Vue components are
PascalCase. The exception is `swisstransport`, the folder on both sides, whose
type stays `transport` so saved dashboards keep working. Browser assets are served
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
- Renderers get data via `widgets.RenderContext` (DashboardID, WidgetTheme,
  ColorMode, Palette, QueryParams, PageIndex, TotalPages) and their injected
  client — they must not do per-request slow fetches; hit the package cache.
