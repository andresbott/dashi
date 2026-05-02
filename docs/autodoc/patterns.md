# Dashi — Patterns

## Adding a New Widget

Backend (`internal/widgets/{pkg}/`):
1. Create `module.go` implementing `widgets.Module` (embed `widgets.NoopModule`
   for optional methods you don't need).
2. Create `static.go` (+ optional HTML template) with your `NewStaticRenderer`.
3. If you need an interactive API: create `handler.go` with an unexported
   `handler` struct and `newHandler` constructor; include in-package
   `writeJSONError` helper. Implement `Module.RegisterRoutes(r)` to mount
   your routes.
4. If you need warmup: implement `Module.Warmup(ctx, configs)`. Parse
   `configs` ([]json.RawMessage) into your config struct and call the
   relevant data client's warmup method.
5. Add tests: `static_test.go`, `handler_test.go` (if applicable),
   `module_test.go` asserting `Type()`, `Renderer()` non-nil,
   `RegisterRoutes` mounts expected paths, and a compile-time
   `var _ widgets.Module = (*Module)(nil)`.
6. Register in `app/router/main.go`: add one line to the `modules` slice.

Frontend (`webui/src/widgets/{type}/`):
1. Create the folder (name matches the type string exactly, including
   any hyphens).
2. Add `Widget.vue` (and `WidgetConfig.vue` if configurable).
3. If you need API calls: add `api.ts` (axios client wrapper) and
   `composable.ts` (Vue Query wrapper). Types go in `types.ts`.
4. Create `index.ts` exporting a `WidgetModule`:

   ```ts
   import { defineAsyncComponent } from 'vue'
   import type { WidgetModule } from '@/widgets/types'

   const myWidget: WidgetModule = {
       type: 'my-widget',
       component: defineAsyncComponent(() => import('./Widget.vue')),
       configComponent: defineAsyncComponent(() => import('./WidgetConfig.vue')),
       label: 'My Widget',
       icon: 'ti-something',
       description: 'What it does',
   }

   export default myWidget
   ```

5. Register in `webui/src/lib/widgetRegistry.ts`: add one import line and
   append to the `modules` array.

### Caveats
- **Folder naming:** backend package name has no hyphens (Go convention);
  frontend folder name matches the widget type string exactly (including
  hyphens). `Module.Type()` returns the hyphenated type string.
- **Warmup configs** are pre-filtered to this widget's type — you don't
  scan dashboards yourself.
- **Handlers must not import `app/router/handlers`** — in-package
  `writeJSONError` mirrors `handlers.ErrorJSON`.
- **Search-style frontend-only widgets:** implement `Module` with a
  placeholder `Renderer()` that emits an empty div.

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
