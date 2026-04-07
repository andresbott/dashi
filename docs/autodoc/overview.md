# Dashi — Architecture Overview

## Purpose

Dashi is a self-hosted dashboard application (Go backend + Vue.js frontend) that
serves configurable widget dashboards in two modes: **interactive** (live Vue.js
SPA) and **image** (server-side rendered PNG for e-ink displays). All data is
file-based (no database). The frontend is embedded into a single Go binary.

## Package Layout

```
dashi/
  main.go                    Entry point → app/cmd
  app/
    cmd/                     CLI commands (server, generate-config, theme)
    router/                  Gorilla Mux routing, middleware, static dashboard serving
      handlers/              API v0 handlers (dashboards, weather, market, themes)
    spa/                     Embedded SPA serving (go:embed webui dist)
    metainfo/                Build version metadata
  internal/
    dashboard/               Dashboard types, file-based store, ID generation
      image/                 PNG rendering via litehtml-go + fogleman/gg
      static/                HTML rendering via Go templates
    widgets/                 Widget registry (type → StaticRenderer function)
      weather/               Weather widget (static + chart rendering)
      market/                Market widget (static + chart rendering)
      bookmark/              Bookmark link widget
      clock/                 Clock widget
      battery/               Battery widget (reads query param)
      pageindicator/         Page indicator dots widget
    themes/                  Theme store (embedded default + user themes from disk)
    weather/                 Open-Meteo API client + in-memory cache (30-min TTL)
    market/                  Yahoo Finance API client + in-memory cache (tiered TTL)
  webui/                     Vue 3 + Vite + PrimeVue frontend
    src/
      views/dashboards/      DashboardListView, DashboardView, DashboardEditView
      components/dashboards/  Widget display + config components
      composables/           Vue Query composables (useDashboards, useWeather, etc.)
      lib/api/               Axios API client modules
      lib/widgetRegistry.ts  Frontend widget registry (component + config + metadata)
      types/                 TypeScript interfaces
      store/                 Pinia stores (minimal UI state)
      router/                Vue Router (/ → first dashboard or list, /dashboards, /:id, /dashboards/:id/edit)
  data/                      Default data directory (dashboards/, themes/)
```

## Key Flows

### Root Route

```
GET / → beforeEnter guard fetches dashboard list
  → If a dashboard has default=true → redirect to /:defaultDashboardId
  → Else if dashboards exist → redirect to /:firstDashboardId (alphabetical)
  → If no dashboards (or fetch error) → redirect to /dashboards (list/create view)
```

### View Interactive Dashboard

```
GET /:id → StaticDashboardMiddleware (type != "image" → pass through)
  → SPA serves Vue app
  → DashboardView.vue fetches GET /api/v0/dashboards/{id}
  → Renders Vue widget components per registry
  → Each widget fetches live data via composables (weather, market)
```

### Render Image Dashboard (PNG)

```
GET /:id?page=N → StaticDashboardMiddleware (type == "image")
  → buildRenderData() (theme fonts, background as base64 data URI)
  → static.Render() calls each widget's StaticRenderer via registry
  → Each renderer fetches data from cache + outputs HTML fragment
  → Full HTML assembled from Go template
  → image.Render() converts HTML → PNG via litehtml-go
  → Returns Content-Type: image/png
```

### Edit Dashboard

```
DashboardEditView.vue fetches dashboard → local deep copy
  → User edits pages/rows/widgets via UI
  → Save → PUT /api/v0/dashboards/{id}
  → DashboardHandler.Update() → Store.Update() → overwrites dashboard.json
  → Vue Query invalidates cache → UI refreshes
```

### Dashboard CRUD (Backend)

```
POST   /api/v0/dashboards                  → Create (generates 6-char alphanumeric ID)
POST   /api/v0/dashboards/upload           → Import dashboard from zip (Content-Type: application/zip)
GET    /api/v0/dashboards                  → List (returns [{id, name, icon, type}])
GET    /api/v0/dashboards/{id}             → Get full dashboard
GET    /api/v0/dashboards/{id}/download    → Export dashboard as zip
PUT    /api/v0/dashboards/{id}             → Update (overwrites dashboard.json)
DELETE /api/v0/dashboards/{id}             → Delete (removes folder)
DELETE /api/v0/dashboards/previews         → Delete all preview dashboards
POST   /api/v0/dashboards/{id}/assets/{path} → Upload asset (Content-Type: application/octet-stream, 10MB max)
GET    /api/v0/dashboards/{id}/assets      → List assets
GET    /api/v0/dashboards/{id}/assets/{path} → Get asset file
DELETE /api/v0/dashboards/{id}/assets/{path} → Delete asset
```

## Widget System

### Backend: Static Rendering

```go
// Registry maps type string → renderer function
type StaticRenderer func(config json.RawMessage, ctx RenderContext) (template.HTML, error)

// RenderContext provides theme, query params, page info
type RenderContext struct {
    Theme       string
    QueryParams map[string]string
    PageIndex   int
    TotalPages  int
}
```

Registered in `router/main.go`. Each widget parses its own config from
`json.RawMessage`, fetches data from cached clients, renders HTML via template.

### Frontend: Interactive Rendering

```ts
// lib/widgetRegistry.ts maps type string → Vue component + config component
interface WidgetRegistryEntry {
  component: Component           // display component
  configComponent: Component     // config dialog (nullable)
  label: string
  icon: string
  description: string
}
```

Widget config stored as `json.RawMessage` / opaque JSON — each widget type
defines its own schema by convention.

### Registered Widgets

| Type | Backend (static) | Frontend (interactive) | Config UI |
|------|-------------------|----------------------|-----------|
| weather | Yes | Yes | Yes |
| weather-compact | Yes | Yes | Yes |
| bookmark | Yes | Yes | Yes |
| clock | Yes | Yes | Yes |
| battery | Yes | Yes | No |
| page-indicator | Yes | Yes | No |
| market | Yes | Yes | No |
| search | No | Yes | Yes |

## Data Storage

All file-based, no database.

- **Dashboards:** `{dataDir}/dashboards/{snake_name}/dashboard.json`
  - In-memory index (`id → folder`) rebuilt on startup
  - Optional sidecar: `custom.css`, `assets/` directory
- **Themes:** Embedded default + `{dataDir}/themes/{name}/theme.yaml`
  - Fonts (TTF), icons (font or image), backgrounds
- **Caches:** In-memory only (weather 30-min TTL, market tiered TTL)

## External Dependencies

- **Open-Meteo API** — weather forecast + geocoding + air quality (no API key)
- **Yahoo Finance API** — market OHLC data
- **litehtml-go** — HTML-to-layout engine for PNG rendering
- **fogleman/gg** — 2D drawing library for PNG output
- **go-bumbu** — shared HTTP middleware, logging, config libraries (sibling repo)
- **PrimeVue 4** — Vue component library
- **TanStack Vue Query** — data fetching/caching

## Server Architecture

Up to three HTTP servers started via errgroup:
1. **Viewer server** (default `:8087`) — read-only dashboard viewer with GET-only API
2. **Editor server** (default `:8088`) — full-CRUD dashboard editor with read + write APIs
3. **Observability server** (default `:9090`) — metrics/health (disabled by default)

Viewer and editor run on separate ports and can be independently enabled/disabled
via `Server.Viewer.Enabled` and `Server.Editor.Enabled`. At least one must be
enabled. When both are enabled, they share the same underlying stores and caches
(built once via `sharedDeps`).

The **viewer** serves only GET APIs (`attachReadAPIs`) and restricts SPA routes
to dashboard ID paths (no `/dashboards`, `/docs`). Root `/` serves the SPA which
resolves the default dashboard client-side.

The **editor** serves both read and write APIs (`attachReadAPIs` + `attachWriteAPIs`)
and the full SPA including list, edit, and documentation views. Root `/` redirects
to `/dashboards`.

Data warmup goroutines pre-fetch weather/market data for all configured
dashboard locations/symbols at startup.
