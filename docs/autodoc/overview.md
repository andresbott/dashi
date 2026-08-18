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
    data/                    Shared user-data (cross-dashboard)
      data.go                FsStore primitive + validateName + sentinels
      notes/                 Markdown notes (string API, goldmark render)
      images/                Images (bytes + mime)
      backgrounds/           Background images (bytes + mime)
    widgets/                 Widget Module interface + CollectConfigs helper
      scanner.go             CollectConfigs helper + DashboardLister interface
      weather/               Weather widget module (module.go, static.go, handler.go, tests)
      market/                Market widget module (module.go, static.go, handler.go, tests)
      bookmark/              Bookmark widget module (module.go, static.go, tests)
      clock/                 Clock widget module (module.go, static.go, tests)
      battery/               Battery widget module (module.go, static.go, tests)
      pageindicator/         Page indicator widget module (module.go, static.go, tests)
      markdown/              Markdown widget module (module.go, static.go, tests)
      search/                Search widget module (module.go, placeholder static.go)
    themes/                  Theme store (embedded default + user themes from disk)
    weather/                 Open-Meteo API client + in-memory cache (30-min TTL)
    market/                  Yahoo Finance API client + in-memory cache (tiered TTL)
  webui/                     Vue 3 + Vite + PrimeVue frontend
    src/
      views/dashboards/      DashboardView, DashboardEditView, DashboardSettingsView
      views/admin/           AdminLayout shell + AdminDashboards, AdminNotes, AdminImages, AdminBackgrounds
      widgets/               Self-contained widget modules
        <type>/              One folder per widget type (name matches type string exactly)
          Widget.vue         Display component
          WidgetConfig.vue   Configuration component (if applicable)
          index.ts           WidgetModule export
          types.ts           TypeScript types (optional)
          api.ts             Axios API client (optional)
          composable.ts      Vue Query composable (optional)
        types.ts             WidgetModule interface
      components/dashboards/  Dashboard-level components (WidgetContainer, etc.)
      components/admin/      DataTableView (shared images/backgrounds table)
      composables/           Dashboard-level composables (useDashboards, useAdminNotes, useDataItems, ...)
      lib/api/               API clients (dashboard.ts, themes.ts, data.ts)
      lib/widgetRegistry.ts  Module-import registry (aggregates widgets/*/index.ts)
      types/                 Dashboard-level TypeScript interfaces
      store/                 Pinia stores (minimal UI state)
      router/                Vue Router (/, /admin/*, /:id, /dashboards/:id/edit, /dashboards/:id/settings, /docs)
  data/                      Default data directory (dashboards/, themes/)
```

## Key Flows

### Root Route

```
Viewer `/` → beforeEnter guard fetches dashboard list
  → If a dashboard has default=true → redirect to /:defaultDashboardId
  → Else if dashboards exist → redirect to /:firstDashboardId (alphabetical)
  → If no dashboards (or fetch error) → redirect to /admin (list/create view)

Editor `/` (backend redirect) → /admin → /admin/dashboards
```

### Admin Section (editor only)

```
GET /admin                    → redirect to /admin/dashboards
GET /admin/dashboards         → AdminDashboards (list/create/import/download/edit/settings/delete)
GET /admin/notes              → AdminNotes (list/create/edit/delete; shared markdown files)
GET /admin/images             → AdminImages (list/upload/delete; shared DataTableView, kind=images)
GET /admin/backgrounds        → AdminBackgrounds (list/upload/delete; shared DataTableView, kind=backgrounds)
```

`AdminLayout.vue` hosts a sticky sidebar nav + `<router-view>`. Images and
Backgrounds share `components/admin/DataTableView.vue`, a component
parameterized by `kind: 'images' | 'backgrounds'`. All three data kinds
(notes/images/backgrounds) go through `lib/api/data.ts`, with a
`useDataItems(kind)` composable for images/backgrounds and a
`useAdminNotes()` composable for notes. The topbar title click
navigates to `/admin`.

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
POST   /api/v0/dashboards/{id}/assets/{path} → Upload asset (Content-Type: application/octet-stream, 10MB max)
GET    /api/v0/dashboards/{id}/assets      → List assets
GET    /api/v0/dashboards/{id}/assets/{path} → Get asset file
DELETE /api/v0/dashboards/{id}/assets/{path} → Delete asset
```

### Shared Data Layer (`/api/v0/data/*`)

```
GET    /api/v0/data/notes                → list of Items
GET    /api/v0/data/notes/{name}         → {"html": ...} rendered HTML
GET    /api/v0/data/notes/{name}/raw     → raw markdown (text/plain)
POST   /api/v0/data/notes/{name}         → save (application/octet-stream, 10MB max)
DELETE /api/v0/data/notes/{name}

GET    /api/v0/data/images               → list of Items
GET    /api/v0/data/images/{name}        → image bytes with correct mime
POST   /api/v0/data/images/{name}
DELETE /api/v0/data/images/{name}

GET    /api/v0/data/backgrounds          → list of Items
GET    /api/v0/data/backgrounds/{name}
POST   /api/v0/data/backgrounds/{name}
DELETE /api/v0/data/backgrounds/{name}
```

GET endpoints are available on viewer + editor; POST/DELETE only on editor.
Content is shared across all dashboards — not scoped to any one dashboard.
Not included in dashboard export/import zips.

## Widget System

### Backend: Module Interface

Each widget implements the `widgets.Module` interface:

```go
type Module interface {
    Type() string                                          // Widget type string (e.g., "weather")
    Renderer() StaticRenderer                              // Returns renderer function for HTML/PNG mode
    RegisterRoutes(r *mux.Router)                          // Optional: mount interactive API routes
    Warmup(ctx context.Context, configs []json.RawMessage) // Optional: pre-fetch data at startup
}

type StaticRenderer func(config json.RawMessage, ctx RenderContext) (template.HTML, error)
```

The `widgets.NoopModule` type can be embedded to provide no-op implementations
of optional methods (`RegisterRoutes`, `Warmup`). Modules are registered as a
slice in `app/router/main.go`. The `widgets.CollectConfigs` helper scans
dashboards and returns configs for a given widget type (used by warmup).

### Frontend: WidgetModule Export

Each widget folder exports a `WidgetModule` from `index.ts`:

```ts
interface WidgetModule {
  type: string                   // Widget type string (matches backend Module.Type())
  component: Component           // Display component (asyncComponent)
  configComponent: Component     // Config dialog (asyncComponent, nullable)
  label: string
  icon: string
  description: string
}
```

The `lib/widgetRegistry.ts` file imports all widget modules and exposes lookup
functions by type. Widget config is stored as `json.RawMessage` (Go) / opaque
JSON (TS) — each widget defines its own schema.

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
| image | Yes | Yes | Yes |
| markdown | Yes | Yes | Yes |
| search | No | Yes | Yes |

> `image` and `markdown` widgets read from the shared data layer (`/api/v0/data/*`) instead of owning their own routes.

## Data Storage

All file-based, no database.

- **Dashboards:** `{dataDir}/dashboards/{snake_name}/dashboard.json`
  - In-memory index (`id → folder`) rebuilt on startup
  - Optional sidecar: `custom.css`, `assets/` directory
- **Shared data:** `{dataDir}/data/{notes,images,backgrounds}/`
  - Flat per-kind namespace, accessible from any dashboard.
  - Not included in dashboard export/import zips.
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
to dashboard ID paths (no `/admin`, `/docs`). Root `/` serves the SPA which
resolves the default dashboard client-side.

The **editor** serves both read and write APIs (`attachReadAPIs` + `attachWriteAPIs`)
and the full SPA including the `/admin` section and documentation views. Root `/`
redirects to `/admin`.

Data warmup goroutines pre-fetch weather/market data for all configured
dashboard locations/symbols at startup.
