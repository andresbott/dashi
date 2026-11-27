# Features — status and where each lives

Check here before adding a capability — the gap may already be catalogued.
Statuses verified against the code on 2026-07-22.

## Dashboard management

| Feature | Status | Where |
|---|---|---|
| CRUD + list | Implemented | `app/router/handlers/dashboards.go`, `internal/dashboard/store.go`; widget configs are checked on save (bookmark URLs, see Not implemented / partial) |
| Zip export/import | Implemented | `Store.ExportZip`/`Store.ImportZip`; routes in `app/router/api_v0.go` |
| Default dashboard flag | Implemented | `Dashboard.Default`; viewer resolves root `/` server-side and redirects (`app/router/main.go:246-260`), falling back to first by name when none marked default |
| Preview dashboards (`-prev` suffix) | **Removed** (earlier on this branch) | Feature and all related symbols (`DeletePreviews`, `previewSuffix`, `isPreviewID`) deleted — historical record only |
| Per-dashboard basic auth | Implemented (viewer only) | `middleware_auth.go`; auth CRUD on editor API; settings UI `DashboardSettingsView.vue` |
| Asset upload/list/delete | Implemented | store asset methods; 10MB cap, extension allowlist incl. `.md` |
| Custom CSS sidecar | Implemented | `Store.GetCustomCSS` (`custom.css` per dashboard) |
| Multi-page + tab navigation | Implemented | `Page` type; `?page=N` for image mode |
| Editor widget drag and drop (rows + stacks) | Implemented (2026-11-27; stacks before that only reordered their own children) | Grip-drag a widget within a row, to another row, into a stack, out of a stack onto any row, or between stacks. Session `webui/src/lib/widgetDrag.ts`, hit testing `lib/widgetDragHit.ts`, every drop applied by `applyWidgetMove` (`lib/widgetMove.ts`) through the `WIDGET_MOVE` injection that `DashboardEditView.vue` provides. Keyboard: arrow keys nudge a row widget by one column only — no keyboard path across rows or into/out of a stack (see Not implemented / partial). Rules: [widgets.md](widgets.md#row-placement) |
| Edit button on the rendered dashboard | Implemented (2026-11-27) | `Dashboard.ShowEditButton` (`showEditButton`, off by default); the "Show Edit button" toggle has shipped in the General section of dashboard settings (`DashboardSettingsView.vue`), interactive dashboards only. Needs `Server.Editor.PublicUrl` (HA add-on: `editor_public_url`) — never derived, so no URL means no button. Link built in `editLink` (`app/router/middleware_static.go`), drawn by `page.html`; never on image dashboards or PNG output. Whether HA ingress forwards the `/dashboards/<id>/edit` sub-path is unverified (else it lands on the editor home). See [viewer.md](viewer.md#edit-button) |

## Widgets

Full matrix and how to add one: [widgets.md](widgets.md). Registered types
(backend `app/router/main.go` + frontend `webui/src/lib/widgetRegistry.ts`):
weather, weather-compact, bookmark, clock, battery, page-indicator, market,
xkcd, transport (Swiss transport), sysinfo, stack, markdown, image, search.
**search** is a plain GET form in the browser viewer (no JS, new tab) and a
static picture of the bar in PNGs.

## Rendering

| Feature | Status | Where |
|---|---|---|
| Interactive viewer (server-rendered HTML) | Implemented | Viewer is server-rendered (see [viewer.md](viewer.md)); editor is Vue SPA in `webui/`, embedded via `app/spa` |
| Image dashboards → PNG | Implemented | `middleware_static.go` → `internal/dashboard/static` (HTML) → `internal/dashboard/image` (litehtml → PNG) |
| Display protocol (headers/query: format, width, height, rotation, action) | Implemented | `parseDisplayHeaders` in `middleware_static.go`; formats: `png`, `png-bw`, `png-spectra6`, packed `bw`, `spectra6`; width/height 1–4096 (`maxDisplaySide`, 400 otherwise) |
| E-ink dithering | Implemented (2026-05-01, PR #18) | `lib/einkimage` (library) + `internal/dashboard/image/dither.go` (integration) |
| Per-dashboard dither mode ("enhanced dither") | Implemented (2026-11-21) | `Dashboard.Dither` (`dither` in dashboard.json): absent = plain Floyd-Steinberg (unchanged output), `auto` = `einkimage.SuggestProcessingOptions` per render, or a preset name; validated on save (`internal/dashboard/dither.go`, 400 on unknown); "Dither" select in `DashboardSettingsView.vue` (image dashboards only). See [rendering.md](rendering.md#dither-modes-per-dashboard-2026-11-21) |
| E-ink preview page (editor + dashboards list) | Implemented (2026-11-21; the first version was a dialog, replaced by this page) | SPA route `/dashboards/:id/eink-preview` (`views/dashboards/DashboardEinkPreviewView.vue`): image left/centre, controls panel on the right (stacked below md) — device preset or custom W×H, format (`png-bw`/`png-spectra6`/`png`), rotation, page, Actual pixels, Refresh. All settings live in the query string (`preset`, `width`/`height` for custom, `format`, `rotation`, `page` 0-based, `actual=1`, `from=edit\|list`; `router.replace`, invalid values fall back to defaults), so reload/bookmark keep them; Back returns via history or to the `from` entry point. 90°/180°/270° renders are shown turned upright with a caption; the panel's PNG is unchanged. Entry points, image dashboards only: the "E-ink preview" button in `DashboardEditView.vue` (flushes pending autosave, opens on the active page) and a tablet icon in each image row of `/admin/dashboards`; an interactive dashboard opened by URL shows a message. Fetches the server PNG from the **editor's own** `GET /{id}?format=…` (no new server route, no viewer auth); nothing saved. Presets and query/rotation helpers in `webui/src/lib/einkPreview.ts`. See [rendering.md](rendering.md#viewer-and-editor-notes) |
| Swipe navigation for displays | Implemented | `X-Action: swipe_left/right` → redirect with new `?page=` |
| Image dashboards viewed in browser | Implemented | Image-type dashboards with no display headers render through the browser stack (see [viewer.md](viewer.md)); `serveImageHTMLPreview` was removed |
| Themes (background + light/dark palette) | Implemented | `internal/themes`; `{dataDir}/data/themes/{name}/theme.json` + `assets/`; builtin read-only `default` and `ember` (registry in `internal/themes/builtin.go`; `GET /themes/{id}` flags them `builtin`); copy (`POST /themes/{id}/copy`); dashboards select via `themeId`; editor `AdminThemeEdit.vue` with the background section in `webui/src/components/admin/ThemeBackgroundEditor.vue` (see [viewer.md](viewer.md#editing-a-background)) |
| Mobile layout (~375px phones) | Implemented (2026-11-21) | Viewer: below 640px every row stacks into one full-width column (CSS variables in `page.html` + rules in `webui/viewer/viewer.css`). Editor SPA: below 768px the admin sidebar is a hamburger drawer, dialogs are capped to the viewport, list tables drop secondary columns, the drag grid keeps 12 columns and scrolls sideways. Breakpoints and details: [viewer.md](viewer.md#responsive-layout-mobile) |
| Admin sidebar grouping | Implemented | `AdminLayout.vue`: **Theming** = Themes, Widget themes; **Content** = user data widgets show (Notes, Images). There is no background-images section: each theme holds its own images, managed in the theme editor |
| Widget themes (fonts + icons) | Implemented | `internal/widgetthemes`; embedded default (Inter + Tabler); user packs `{dataDir}/data/widget_themes/{name}/widget_theme.yaml`; zip upload/download; uploads and deletes take effect without a restart in both stacks (PNG fonts re-synced via `Store.OnChange`, see [rendering.md](rendering.md#fonts)); hand-placed or CLI-created packs still need a restart; dashboards select via `widgetTheme`; old `theme.yaml` packs are rejected on upload |
| Widget theme bootstrap CLI | Implemented | `dashi widget-theme create [name] --type image|font` (`app/cmd/widget_theme.go`) |

No migration: old `data/backgrounds`, `data/shared-background-images`,
`{dataDir}/themes` packs, theme image refs using the removed `shared:` scheme
and the dashboard keys `theme`, `backgroundId`, `accentColor` are ignored; dangling references
fall back to `default`. Details: [viewer.md](viewer.md#theming-two-entities-css-variables-and-concrete-values).

## Data sources (all in-memory cached, warmup at startup for weather/market)

| Source | TTL | Where |
|---|---|---|
| Weather (Open-Meteo, no key) | 30 min | `internal/providers/weather/` (client + cache) |
| Market (Yahoo Finance) | tiered by range: 15 min intraday … 24 h | `internal/providers/market/cache.go` |
| XKCD | 1 h + disk cache dir `{dataDir}/cache/xkcd` | `internal/providers/xkcd/` (client + cache) |
| Swiss transport (transport.opendata.ch) | 30 s | `internal/providers/swisstransport/` (client + cache) |
| Sysinfo (gopsutil, local) | 30 s | `internal/providers/sysinfo/` (client + cache) |

## Not implemented / partial (direction already chosen — don't re-design)

- **Observability server**: config + port exist, handler is `nil`
  (`app/cmd/server.go` TODO). Prometheus histogram middleware is already wired
  into both routers; only the metrics endpoint is missing.
- **App-level auth** (`Auth` config block): scaffolding only, nothing consumes
  it. Per-dashboard basic auth is the only working auth.
- **Widget config validation**: partial by design. The mechanism is in place
  (optional `widgets.ConfigValidator`, run by the dashboard store on create,
  update and zip import; a rejection answers 400), but only **bookmark** (URL
  allowlist) and **stack** (delegates to its children) declare rules. Other
  widgets stay pass-through; add a validator when a widget puts config into a
  dangerous sink. See [architecture.md](architecture.md#security-model--know-the-boundaries).
- **Keyboard moves into or out of a stack** (known debt, deferred
  2026-11-27): only the pointer can drag a widget into a stack, out of one,
  or between stacks; a stack child's grip has no arrow-key handling. Rows
  have no keyboard cross-row move either, so both are left for one keyboard
  design (e.g. a "Move to…" menu on the grip) that drives the same
  `WIDGET_MOVE` / `applyWidgetMove` path.
- **Market widget config UI exists** (`MarketWidgetConfig.vue`) — older
  autodoc saying "Not implemented" is stale.
- Root `TODO.md` backlog: HA addon, isolated widget components, remove
  previews, user-data dir migration. Enhanced dither and the e-ink preview
  (shipped as a dialog, now a page) have shipped (see Rendering); the file
  still lists them.
