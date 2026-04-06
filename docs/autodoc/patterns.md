# Dashi — Patterns

## Adding a New Widget

Files to modify:

**Backend (static/image rendering):**
1. `internal/widgets/{name}/` — new package
   - `static.go` — `NewStaticRenderer(deps) widgets.StaticRenderer` function
   - Embedded HTML template (optional, or inline)
   - `static_test.go` — test the renderer
2. `app/router/main.go` — register: `registry.Register("{name}", {name}widget.NewStaticRenderer(...))`

**Frontend (interactive rendering):**
3. `webui/src/components/dashboards/{Name}Widget.vue` — display component
   - Props: `widget: Widget` (access config via `JSON.parse(widget.config)`)
4. `webui/src/components/dashboards/{Name}WidgetConfig.vue` — config editor (optional)
   - Props: `modelValue: object`, emits `update:modelValue`
5. `webui/src/lib/widgetRegistry.ts` — register entry with component, configComponent, label, icon, description

Naming: Widget type is lowercase kebab (e.g., `weather-compact`). Package name
is the type without hyphens. Vue components are PascalCase (e.g., `WeatherCompactWidget.vue`).

### Caveats
- Backend and frontend registries must use the same type string
- Config is opaque JSON — define your own struct/interface, no shared schema
- If the widget needs external data, inject the client via the renderer constructor (backend) or use a composable (frontend)
- Image-mode widgets must produce self-contained HTML (inline styles, base64 images) — no external resources

## Adding a New API Endpoint

Files to modify:
1. `app/router/handlers/{name}.go` — handler struct with methods
   - Constructor: `New{Name}Handler(deps) *{Name}Handler`
   - Methods return `http.HandlerFunc`
2. `app/router/api_v0.go` — register routes on the subrouter
3. `app/router/main.go` — instantiate handler, pass to `apiV0Routes()`
4. `app/cmd/server.go` — instantiate dependencies if new (client, store, etc.)

Naming: Handler files match the resource name. Routes follow REST conventions
under `/api/v0/`.

### Caveats
- Read-only mode: write endpoints must check `readOnly` flag and return 403
- Error handling uses go-bumbu HTTP error middleware — return errors via the pattern used in existing handlers

## Adding a New External Data Source

Files to modify:
1. `internal/{name}/client.go` — API client with in-memory cache
   - Cache struct with TTL, mutex-protected map
   - Public method to fetch data (checks cache first)
   - `WarmupX()` method for startup pre-fetch
2. `internal/{name}/types.go` — response types
3. `internal/{name}/client_test.go` — tests
4. `app/cmd/server.go` — instantiate client, add warmup goroutine
5. `app/router/main.go` — pass client to handler/widget constructors
6. `app/router/handlers/{name}.go` — API handler (if data exposed directly)

Naming: Package name matches the data domain (e.g., `weather`, `market`).

### Caveats
- Cache is in-memory only — lost on restart, rebuilt by warmup
- Warmup iterates all configured dashboards to find relevant widget configs
- Client should handle HTTP errors gracefully (the cache returns stale data on failure)

## Adding a New Theme

No code changes required. Create directory structure:

```
{dataDir}/themes/{themeName}/
  theme.yaml          — manifest (name, description, fonts, icons)
  fonts/              — TTF font files
  backgrounds/        — background images (optional)
  widgets/weather/icons/  — weather condition icons (optional)
```

Theme manifest format — see `internal/themes/defaults/theme.yaml` for reference.

### Caveats
- Theme name in `theme.yaml` must match directory name
- Font files must be TTF format
- Icons can be font-based (single TTF + codepoint map) or image-based (individual files)
- Icon canonical names must match the set used by widgets (see weather codes in `internal/weather/codes.go`)
