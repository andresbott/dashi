# Features — status and where each lives

Check here before adding a capability — the gap may already be catalogued.
Statuses verified against the code on 2026-07-22.

## Dashboard management

| Feature | Status | Where |
|---|---|---|
| CRUD + list | Implemented | `app/router/handlers/dashboards.go`, `internal/dashboard/store.go` |
| Zip export/import | Implemented | `Store.ExportZip`/`Store.ImportZip`; routes in `app/router/api_v0.go` |
| Default dashboard flag | Implemented | `Dashboard.Default`; viewer resolves root `/` server-side and redirects (`app/router/main.go:246-260`), falling back to first by name when none marked default |
| Preview dashboards (`-prev` suffix) | **Removed** (earlier on this branch) | Feature and all related symbols (`DeletePreviews`, `previewSuffix`, `isPreviewID`) deleted — historical record only |
| Per-dashboard basic auth | Implemented (viewer only) | `middleware_auth.go`; auth CRUD on editor API; settings UI `DashboardSettingsView.vue` |
| Asset upload/list/delete | Implemented | store asset methods; 10MB cap, extension allowlist incl. `.md` |
| Custom CSS sidecar | Implemented | `Store.GetCustomCSS` (`custom.css` per dashboard) |
| Multi-page + tab navigation | Implemented | `Page` type; `?page=N` for image mode |

## Widgets

Full matrix and how to add one: [widgets.md](widgets.md). Registered types
(backend `app/router/main.go` + frontend `webui/src/lib/widgetRegistry.ts`):
weather, weather-compact, bookmark, clock, battery, page-indicator, market,
xkcd, transport (Swiss transport), sysinfo, stack, markdown, image — plus
frontend-only **search** (no static renderer: an input box is meaningless in a
PNG — DIY-by-design gap, do not "fix" it).

## Rendering

| Feature | Status | Where |
|---|---|---|
| Interactive viewer (server-rendered HTML) | Implemented | Viewer is server-rendered (see [viewer.md](viewer.md)); editor is Vue SPA in `webui/`, embedded via `app/spa` |
| Image dashboards → PNG | Implemented | `middleware_static.go` → `internal/dashboard/static` (HTML) → `internal/dashboard/image` (litehtml → PNG) |
| Display protocol (headers/query: format, width, height, rotation, action) | Implemented | `parseDisplayHeaders` in `middleware_static.go`; formats: `png`, `png-bw`, `png-spectra6`, packed `bw`, `spectra6` |
| E-ink dithering | Implemented (2026-05-01, PR #18) | `lib/einkimage` (library) + `internal/dashboard/image/dither.go` (integration); "enhanced dither" listed as debt in TODO.md |
| Swipe navigation for displays | Implemented | `X-Action: swipe_left/right` → redirect with new `?page=` |
| Image dashboards viewed in browser | Implemented | Image-type dashboards with no display headers render through the browser stack (see [viewer.md](viewer.md)); `serveImageHTMLPreview` was removed |
| Themes (fonts, icons, backgrounds) | Implemented | `internal/themes`; embedded default (Inter + Tabler); user themes `{dataDir}/themes/{name}/theme.yaml`; legacy manifest format still parsed |
| Theme bootstrap CLI | Implemented | `dashi theme create [name] --type image|font` |

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
- **Widget config validation / XSS fix**: catalogued in `docs/TODO.md` with fix
  options (frontend URL validation, backend per-type validation, or both).
- **Market widget config UI exists** (`MarketWidgetConfig.vue`) — older
  autodoc saying "Not implemented" is stale.
- Root `TODO.md` backlog: HA addon, enhanced dither, isolated widget
  components, remove previews, recreate themes, e-ink preview dialog,
  user-data dir migration.
