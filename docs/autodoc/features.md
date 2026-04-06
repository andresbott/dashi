# Dashi — Feature Status

## Dashboard Management

| Feature | Status | Notes |
|---------|--------|-------|
| Create dashboard | Implemented | Name, icon, type (interactive/image) |
| List dashboards | Implemented | Grid view with cards |
| Edit dashboard | Implemented | Full editor: pages, rows, widgets, theme, background |
| Delete dashboard | Implemented | With confirmation dialog |
| Dashboard preview | Implemented | Creates "-prev" suffixed copy for image preview |
| Delete previews | Implemented | Bulk delete all preview dashboards |
| Default dashboard | Implemented | Boolean flag per dashboard; root `/` redirects to it |
| Read-only mode | Implemented | Config flag disables all writes, UI hides edit controls |

**Create dashboard** — POST creates a new dashboard with a random 6-char
alphanumeric ID. Folder name derived from snake_case of dashboard name with
collision detection.

**Edit dashboard** — Full editor with page management (add/delete/rename/reorder),
row management (add/delete/reorder, height/width/title), widget management
(add/delete/reorder via drag, type/title/width/config). Local copy edited then
saved via PUT.

**Dashboard preview** — Creates a temporary copy with ID ending in "-prev" so
image dashboards can be previewed before saving. Bulk delete cleans all previews.

## Dashboard Display

| Feature | Status | Notes |
|---------|--------|-------|
| Interactive mode | Implemented | Vue SPA with live widget components |
| Image mode (PNG) | Implemented | Server-side HTML→PNG via litehtml-go |
| Multi-page navigation | Implemented | Tab navigation, ?page=N query param |
| Theme support | Implemented | Font injection, icon resolution |
| Color mode | Implemented | auto/light/dark |
| Accent color | Implemented | CSS variable applied to dashboard |
| Background: none | Implemented | |
| Background: color | Implemented | |
| Background: gradient | Implemented | Gradient editor in UI |
| Background: image | Implemented | From theme or uploaded asset |
| Custom CSS | Implemented | Sidecar custom.css per dashboard |
| Container settings | Implemented | Max-width, alignment, show-boxes |

## Widget Types

| Widget | Interactive | Image (static) | Config UI | Notes |
|--------|-------------|----------------|-----------|-------|
| weather | Implemented | Implemented | Implemented | Full forecast, hourly, details, charts |
| weather-compact | Implemented | Implemented | Implemented | Minimal weather display |
| bookmark | Implemented | Implemented | Implemented | Link with title, subtitle, icon |
| clock | Implemented | Implemented | Implemented | 12/24h, seconds, date |
| battery | Implemented | Implemented | No config | Reads % from query param |
| page-indicator | Implemented | Implemented | No config | Dot indicators for pages |
| market | Implemented | Implemented | Not implemented | Stock/crypto ticker with chart |
| search | Implemented | Not implemented | Implemented | Search bar, configurable engine |

## Asset Management

| Feature | Status | Notes |
|---------|--------|-------|
| Upload assets | Implemented | Per-dashboard, 10MB limit |
| List assets | Implemented | API + UI |
| Delete assets | Implemented | |
| Allowed file types | Implemented | .png, .jpg, .jpeg, .svg, .webp, .css |
| Path traversal protection | Implemented | Rejects ".." in paths |

## Themes

| Feature | Status | Notes |
|---------|--------|-------|
| Default embedded theme | Implemented | Inter font + Tabler icons |
| User themes from disk | Implemented | {dataDir}/themes/{name}/ |
| Font-based icons | Implemented | CSS class + codepoint |
| Image-based icons | Implemented | SVG/PNG files |
| Theme backgrounds | Implemented | Served via API |
| Font serving | Implemented | TTF via API endpoint |

## External Data

| Feature | Status | Notes |
|---------|--------|-------|
| Weather forecast | Implemented | Open-Meteo API, 30-min cache |
| Geocoding | Implemented | Open-Meteo geocoding API |
| Air quality | Implemented | Open-Meteo AQI |
| Market data | Implemented | Yahoo Finance, tiered cache |
| Data warmup | Implemented | Pre-fetch on startup |

## Infrastructure

| Feature | Status | Notes |
|---------|--------|-------|
| Embedded frontend | Implemented | go:embed, single binary |
| Config file generation | Implemented | `dashi generate-config` |
| Env var overrides | Implemented | DASHI_ prefix |
| Observability server | Partial | Prometheus metrics, health TODO |
| Logging | Implemented | slog with custom formatting |
