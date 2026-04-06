# Widget Rendering Architectures

Dashi supports two fundamentally different rendering modes for dashboards, and every widget
must be implemented for both. Understanding each pipeline is essential before building a new
widget.

| | Interactive (Vue) | Image (server-side PNG) |
|-|-------------------|------------------------|
| **Dashboard type** | `"interactive"` | `"image"` |
| **Rendering happens** | In the browser (Vue 3) | On the Go backend |
| **Output** | Reactive DOM | Static PNG image |
| **Styling** | CSS with Vue scoped styles, PrimeVue tokens | Inline CSS in HTML, rendered by litehtml |
| **Data fetching** | Client-side via TanStack Query composables | Server-side in the Go renderer function |
| **Interactivity** | Full (hover, click, timers, transitions) | None (static snapshot) |
| **Use case** | Browser-based dashboards | E-ink displays, sharing, printing |

---

## Shared Concepts

Both architectures share the same data model and configuration format. Understanding this
shared foundation is important before diving into each pipeline.

### Dashboard Structure

```
Dashboard
  +-- id, name, type ("interactive" | "image"), theme, colorMode
  +-- container: { maxWidth, horizontalAlign, verticalAlign }
  +-- imageConfig: { width, height }         (image dashboards only)
  +-- background: { type, value }
  +-- pages[]
        +-- name
        +-- rows[]
              +-- height, width, title
              +-- widgets[]
                    +-- id, type, title, width (1-12 grid units)
                    +-- config: JSON (opaque, widget-specific)
```

Defined in `internal/dashboard/types.go`.

### Widget Configuration

Widget config is opaque `json.RawMessage` on the backend and a plain object on the frontend.
Each widget defines its own config schema. There is no shared validation layer --- each
renderer parses the same JSON independently. The type string (e.g. `"weather"`, `"bookmark"`)
must be identical in both the Go registry and the Vue registry.

### Grid System

Both renderers use a 12-column grid. A widget with `width: 6` occupies 50% of its row.
Width defaults to 12 (full row) when omitted or zero.

---

## Architecture 1: Interactive (Vue) Rendering

Interactive dashboards render entirely in the browser. The Go backend serves the dashboard
JSON via the API, and Vue components handle all layout, data fetching, and display.

### Request Flow

```
Browser request: GET /{dashboard-id}
  |
  +--> Static middleware checks dashboard.type
  |      type != "image" --> falls through to SPA
  |
  +--> Vue SPA loads
  |      GET /api/v0/dashboards/{id} --> dashboard JSON
  |      GET /api/v0/themes          --> theme metadata (fonts, icons)
  |
  +--> DashboardView.vue
  |      provides: DASHBOARD_THEME, DASHBOARD_ID, ACTIVE_PAGE, TOTAL_PAGES
  |      injects @font-face rules into <head>
  |      renders pages[activePage].rows --> DashboardRow --> DashboardWidget
  |
  +--> DashboardWidget.vue
  |      looks up widget type in widgetRegistry.ts
  |      renders <component :is="entry.component" :widget="widget" />
  |
  +--> Individual widget component (e.g. WeatherWidget.vue)
         parses config from widget.config
         fetches data via composable (e.g. useWeather)
         renders reactive template
```

### Vue Widget Registry

File: `webui/src/lib/widgetRegistry.ts`

```typescript
interface WidgetRegistryEntry {
    component: Component              // display component
    configComponent: Component | null // config editor (used in edit mode)
    label: string                     // human-readable name
    icon: string                      // Tabler icon class
    description: string               // help text
    noWidgetProp?: boolean            // skip passing widget prop (e.g. page-indicator)
}
```

Components are registered as async imports:

```typescript
weather: {
    component: defineAsyncComponent(() => import('@/components/dashboards/WeatherWidget.vue')),
    configComponent: defineAsyncComponent(() => import('@/components/dashboards/WeatherWidgetConfig.vue')),
    label: 'Weather',
    icon: 'ti-sun',
    description: 'Current conditions and forecast',
},
```

### How a Vue Widget Works

Each widget is a standard Vue 3 `<script setup>` component that receives its data through
props and manages its own rendering and data fetching.

**Props:** Every widget receives `widget: Widget` as a prop (unless `noWidgetProp` is set).
The widget's JSON config is accessed via `props.widget.config`.

**Data fetching:** Widgets that need external data use composables backed by TanStack Query.
For example, `WeatherWidget.vue` calls `useWeather(lat, lon)` which fetches from
`/api/v0/widgets/weather?lat=X&lon=Y`. The composable handles caching, loading states,
and error states.

**Dashboard context:** Widgets can inject dashboard-level values:
- `DASHBOARD_THEME` --- theme name (for icon resolution)
- `DASHBOARD_ID` --- dashboard ID (for asset URLs like `dashboard:assets/icon.webp`)
- `ACTIVE_PAGE` / `TOTAL_PAGES` --- pagination context

**Styling:** Uses Vue scoped `<style scoped>` blocks with PrimeVue CSS variables
(e.g. `var(--p-primary-color)`, `var(--p-text-muted-color)`).

### Example: Bookmark Widget (Vue)

`webui/src/components/dashboards/BookmarkWidget.vue`

This is a simple widget with no data fetching:

```vue
<script setup lang="ts">
const props = defineProps<{ widget: Widget }>()
const dashboardId = inject(DASHBOARD_ID, ref(''))

// Parse config from widget JSON
const config = computed<BookmarkWidgetConfig | null>(() => {
    const c = props.widget.config as unknown as BookmarkWidgetConfig
    if (!c.url || !c.title) return null
    return c
})

// Resolve icon source (Tabler class, self-hosted URL, or dashboard asset URL)
const iconInfo = computed(() => {
    const parsed = parseIcon(config.value.icon)
    if (parsed.type === 'selfhst') return { type: 'image', src: getSelfhstIconUrl(parsed.value) }
    if (parsed.type === 'dashboard') return { type: 'image', src: getDashboardIconUrl(dashboardId.value, parsed.value) }
    return { type: 'tabler', class: 'ti ' + parsed.value }
})
</script>

<template>
    <a :href="config.url" target="_blank" class="bookmark-link">
        <img v-if="iconInfo?.type === 'image'" :src="iconInfo.src" />
        <i v-else-if="iconInfo?.type === 'tabler'" :class="iconInfo.class" />
        <div class="bookmark-text">
            <span class="bookmark-title">{{ config.title }}</span>
            <span v-if="config.subtitle" class="bookmark-subtitle">{{ config.subtitle }}</span>
        </div>
    </a>
</template>
```

Key observations:
- Config parsed from `widget.config` via computed property
- Icon resolution handles multiple sources (Tabler icons, external images, dashboard assets)
- No API calls --- bookmarks are purely client-side
- Scoped CSS with PrimeVue variables for theme-aware colors

### Example: Weather Widget (Vue, with data fetching)

`webui/src/components/dashboards/WeatherWidget.vue`

```vue
<script setup lang="ts">
const iconTheme = inject(DASHBOARD_THEME, ref('default'))

// TanStack Query composable --- handles caching, loading, errors
const { data: weather, isLoading, isError } = useWeather(lat, lon)

// Theme data for icon resolution
const { data: themes } = useThemes()
</script>

<template>
    <div v-if="isLoading">Loading...</div>
    <div v-else-if="isError">Error</div>
    <div v-else>
        <WeatherIcon :icon-name="weather.current.icon" :theme-name="iconTheme" :themes="themes" />
        <span>{{ weather.current.temperature }}deg</span>
    </div>
</template>
```

Key observations:
- Data fetched client-side via `useWeather()` composable
- Composable hits `/api/v0/widgets/weather?lat=X&lon=Y` (backend caches API responses)
- Loading/error states handled reactively
- Icons resolved through theme system

---

## Architecture 2: Image (Server-Side PNG) Rendering

Image dashboards are rendered entirely on the Go backend. The server generates an HTML
document from widget renderers, then converts it to a PNG image using the litehtml-go
library.

### Request Flow

```
HTTP request: GET /{dashboard-id}
  |
  +--> Static middleware (middleware_static.go)
  |      store.Get(id) --> dashboard JSON
  |      dashboard.type == "image" --> handle here
  |
  +--> Parse page index from ?page=N query param
  |
  +--> Build RenderData
  |      theme, fonts, colorMode, customCSS, container settings
  |      extract rows from pages[pageIdx]
  |
  +--> For each widget in each row:
  |      registry.Render(widgetType, config, RenderContext)
  |        --> calls the registered StaticRenderer function
  |        --> returns HTML fragment (template.HTML)
  |
  +--> Static Renderer (static/renderer.go)
  |      assembles master.html template with all widget HTML fragments
  |      applies 12-column grid layout (float-based, not flexbox)
  |      injects theme fonts, custom CSS, background CSS
  |      outputs complete HTML document
  |
  +--> Image Renderer (image/renderer.go)
  |      creates litehtml document from HTML
  |      draws background image (scaled to cover) onto RGBA canvas
  |      renders litehtml document onto canvas
  |      encodes as PNG
  |
  +--> Response: image/png
```

### Go Widget Registry

File: `internal/widgets/registry.go`

```go
type StaticRenderer func(config json.RawMessage, ctx RenderContext) (template.HTML, error)

type RenderContext struct {
    Theme       string            // dashboard theme name
    QueryParams map[string]string // URL query parameters
    PageIndex   int               // zero-based current page
    TotalPages  int               // total pages in dashboard
}
```

Registration happens in `app/router/main.go`:

```go
registry := widgets.NewRegistry()
registry.Register("weather", weatherwidget.NewStaticRenderer(weatherClient, themeStore))
registry.Register("bookmark", bookmarkwidget.NewStaticRenderer())
registry.Register("clock", clockwidget.NewStaticRenderer(nil))
registry.Register("market", marketwidget.NewStaticRenderer(marketClient))
// ...
```

### How a Static Widget Renderer Works

A static renderer is a Go function that receives the widget's JSON config and a render
context, and returns an HTML fragment. The fragment is inserted into the master HTML template
alongside other widgets.

The function signature is:

```go
func(config json.RawMessage, ctx RenderContext) (template.HTML, error)
```

The typical pattern:
1. **Parse config** --- unmarshal JSON into a typed struct
2. **Fetch data** --- call external data clients if needed (weather API, market API)
3. **Resolve icons** --- use the theme store to map canonical icon names to font glyphs or image paths
4. **Build template data** --- prepare a struct with all display values
5. **Generate charts/images** --- render any charts as PNG, base64-encode them for embedding
6. **Execute template** --- render an embedded HTML template, return the HTML fragment

### The HTML-to-PNG Pipeline

The image rendering pipeline has two stages:

**Stage 1: HTML Assembly** (`internal/dashboard/static/renderer.go`)

The static renderer builds a complete HTML page using `master.html`. The template:
- Uses a float-based 12-column grid (`.widget-cell { float: left; width: XX% }`)
- Applies theme font family, dark/light color mode, custom CSS
- Background is set via CSS (gradients work, images use data URIs)
- Each widget's HTML fragment is injected into a grid cell

**Stage 2: PNG Rendering** (`internal/dashboard/image/renderer.go`)

The image renderer uses `litehtml-go` to convert the HTML page to a PNG:

```go
func (r *Renderer) Render(html string, width, height int, backgroundImage ...[]byte) ([]byte, error)
```

- Creates a litehtml document from the HTML string
- Registers custom fonts (theme display fonts, icon fonts, embedded Inter/GoMono)
- If a background image is provided, draws it scaled-to-cover on the canvas
- Renders the litehtml document on top
- Encodes the result as PNG

The canvas dimensions come from `dashboard.imageConfig.width` and `dashboard.imageConfig.height`.
If height is 0, it auto-sizes to content height.

### litehtml Constraints

litehtml is a lightweight HTML/CSS renderer. It supports a large subset of CSS but has
important limitations. See `docs/project/litehtml-rendering-reference.md` for the full
reference. Key points:

**Safe to use:**
- Block, inline, inline-block, flexbox layout
- All standard box model properties (margin, padding, border, border-radius)
- Backgrounds: solid colors, linear/radial/conic gradients
- Text: font-family, font-size, font-weight, text-align, line-height, white-space
- Colors: named, hex, rgb(), rgba()
- Generated content (::before, ::after)
- Tables

**Not supported:**
- JavaScript (obviously --- it's a static renderer)
- CSS Grid
- CSS custom properties (variables like `--my-var`)
- CSS animations or transitions
- Outlines
- RTL/bidirectional text

**Critical rule:** Widget HTML must be entirely self-contained. No external resources ---
all images must be base64-encoded data URIs, all styles must be inline or part of the
master template. The renderer has no network access during HTML rendering.

### Example: Bookmark Widget (Go Static Renderer)

`internal/widgets/bookmark/bookmark.go`

This is a minimal renderer with no data fetching:

```go
type bookmarkConfig struct {
    URL      string `json:"url"`
    Icon     string `json:"icon"`
    Title    string `json:"title"`
    Subtitle string `json:"subtitle"`
}

func NewStaticRenderer() func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
    return func(config json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
        var cfg bookmarkConfig
        json.Unmarshal(config, &cfg)

        data := bookmarkData{
            URL:      sanitizeURL(cfg.URL),   // rejects javascript:, data:, vbscript:
            Title:    cfg.Title,
            Subtitle: cfg.Subtitle,
        }

        var buf bytes.Buffer
        tmpl.Execute(&buf, data)
        return template.HTML(buf.String()), nil
    }
}
```

The HTML template (`bookmark.html`) is embedded via `//go:embed`. Key observations:
- Config struct mirrors the same JSON schema used by the Vue component
- URL sanitization prevents XSS (important since the HTML is trusted by litehtml)
- No external dependencies --- the renderer is self-contained
- Returns an HTML fragment, not a full page

### Example: Weather Widget (Go Static Renderer, with data fetching)

`internal/widgets/weather/static.go`

This is a complex renderer with external data and chart generation:

```go
func NewStaticRenderer(client *weather.Client, themeStore *themes.Store) StaticRenderer {
    return func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
        var cfg weatherConfig
        json.Unmarshal(config, &cfg)

        // 1. Fetch live weather data (server-side, cached)
        wd, err := client.GetWeather(cfg.Latitude, cfg.Longitude)

        // 2. Resolve icons via theme store (font glyph or image path)
        currentIconHTML := resolveIconHTML(themeStore, ctx.Theme, wd.Current.Icon)

        // 3. If graph enabled, generate chart as PNG and base64-encode it
        if cfg.ShowGraph {
            chartPNG, _ := chart.Generate(graphPoints, chartOptions)
            data.GraphImage = base64.StdEncoding.EncodeToString(chartPNG)
        }

        // 4. Execute HTML template with all data
        tmpl.Execute(&buf, data)
        return template.HTML(buf.String()), nil
    }
}
```

Key observations:
- External data client injected via constructor (not created per-request)
- Data is cached in-memory by the weather client (24-hour TTL)
- Icons resolve to either font glyphs (`<span style="font-family: icon-font-X">U+XXXX</span>`)
  or image tags (`<img src="/path/to/icon.svg">`)
- Charts are rendered to PNG using Go's image libraries, then base64-encoded into `<img>` tags
- The startup warmup system pre-fetches data so the first render is instant

---

## Side-by-Side: Same Widget, Both Architectures

This section compares how the same widget concept maps to each architecture, to help you
understand what goes where when implementing a new widget.

### Data Fetching

| Concern | Vue (interactive) | Go (image) |
|---------|-------------------|------------|
| Where data is fetched | Client-side in a composable | Server-side in the renderer function |
| Caching | TanStack Query (client-side) | In-memory cache in the Go client |
| Loading states | `isLoading` / `isError` from composable | Not applicable (sync render) |
| Error handling | Show error UI in template | Return error from renderer |
| Data client lifecycle | Created once in composable, shared via query cache | Injected into renderer via constructor |

### Icon Resolution

| Concern | Vue (interactive) | Go (image) |
|---------|-------------------|------------|
| How icons render | `<i class="ti ti-sun">` or `<img src="...">` | Font glyph span or `<img>` with file path |
| Theme icons | Fetched via `/api/v0/themes/{name}/icons/{icon}` | Resolved via `themeStore.ResolveIcon()` |
| Icon fonts | Loaded as web fonts in browser | Registered as custom fonts in litehtml container |
| Dashboard asset icons | `getDashboardIconUrl(id, path)` -> API URL | Not currently supported (only theme icons) |

### Chart Rendering

| Concern | Vue (interactive) | Go (image) |
|---------|-------------------|------------|
| Where charts render | Client-side SVG (computed paths) | Server-side PNG (Go image libraries) |
| Output format | `<svg>` element with `<path>` | Base64-encoded `<img src="data:image/png;base64,...">` |
| Sizing | Responsive (scales with container) | Fixed pixel dimensions |

### Styling

| Concern | Vue (interactive) | Go (image) |
|---------|-------------------|------------|
| CSS approach | Scoped styles + PrimeVue variables | Inline styles + master.html template styles |
| Theme colors | `var(--p-primary-color)` etc. | `{{if .IsDark}}#999{{else}}#666{{end}}` in template |
| Fonts | Loaded as @font-face in `<head>` | Registered on the image renderer |
| Responsiveness | Fluid, adapts to viewport | Fixed dimensions from imageConfig |

---

## Implementing a New Widget

### Step-by-Step Checklist

When adding a new widget, you must implement it in both architectures. The type string
must match exactly between them.

#### 1. Go Static Renderer (image dashboard support)

Create a new package: `internal/widgets/{name}/`

**`static.go`:**
```go
package mywidget

import (
    "bytes"
    "encoding/json"
    "html/template"
    _ "embed"
    "github.com/andresbott/dashi/internal/widgets"
)

//go:embed static.html
var staticHTML string
var tmpl = template.Must(template.New("mywidget").Parse(staticHTML))

type mywidgetConfig struct {
    // Fields matching your widget's JSON config
    Title string `json:"title"`
}

type mywidgetData struct {
    // Fields passed to the HTML template
    Title string
}

func NewStaticRenderer() func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
    return func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
        var cfg mywidgetConfig
        if len(config) > 0 {
            if err := json.Unmarshal(config, &cfg); err != nil {
                return "", fmt.Errorf("mywidget config: %w", err)
            }
        }

        data := mywidgetData{Title: cfg.Title}

        var buf bytes.Buffer
        if err := tmpl.Execute(&buf, data); err != nil {
            return "", fmt.Errorf("mywidget render: %w", err)
        }
        return template.HTML(buf.String()), nil
    }
}
```

**`static.html`:**
```html
<div class="widget-mywidget">
    <span>{{.Title}}</span>
</div>
```

**If your widget needs external data**, inject the client via the constructor:
```go
func NewStaticRenderer(client *myclient.Client) func(...) (...) {
    return func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
        data, err := client.GetData(...)
        // ...
    }
}
```

**Register** in `app/router/main.go`:
```go
registry.Register("mywidget", mywidget.NewStaticRenderer())
```

#### 2. Vue Component (interactive dashboard support)

**`webui/src/components/dashboards/MywidgetWidget.vue`:**
```vue
<script setup lang="ts">
import { computed } from 'vue'
import type { Widget } from '@/types/dashboard'

const props = defineProps<{ widget: Widget }>()

interface MywidgetConfig {
    title: string
}

const config = computed<MywidgetConfig | null>(() => {
    if (!props.widget.config) return null
    return props.widget.config as unknown as MywidgetConfig
})
</script>

<template>
    <div class="mywidget">
        <span v-if="config">{{ config.title }}</span>
        <span v-else>Configure widget</span>
    </div>
</template>

<style scoped>
.mywidget {
    padding: 0.5rem;
}
</style>
```

**Optional config component** (`MywidgetWidgetConfig.vue`):
```vue
<script setup lang="ts">
const props = defineProps<{ config: Record<string, unknown> | null }>()
const emit = defineEmits<{ 'update:config': [config: Record<string, unknown>] }>()
</script>

<template>
    <div class="flex flex-column gap-2">
        <label>Title</label>
        <InputText :model-value="config?.title" @update:model-value="emit('update:config', { ...config, title: $event })" />
    </div>
</template>
```

**Register** in `webui/src/lib/widgetRegistry.ts`:
```typescript
mywidget: {
    component: defineAsyncComponent(() => import('@/components/dashboards/MywidgetWidget.vue')),
    configComponent: defineAsyncComponent(() => import('@/components/dashboards/MywidgetWidgetConfig.vue')),
    label: 'My Widget',
    icon: 'ti-star',
    description: 'Description of the widget',
},
```

#### 3. Test

Write a Go test for the static renderer:
```go
func TestStaticRenderer(t *testing.T) {
    renderer := NewStaticRenderer()
    html, err := renderer(json.RawMessage(`{"title":"Hello"}`), widgets.RenderContext{})
    // assert no error, assert html contains expected content
}
```

#### 4. Verify Visual Parity

After implementing both sides, compare the visual output:

- **Interactive**: run `cd webui && npm run dev`, create a test dashboard with your widget
- **Image**: create an image-type dashboard, request it with `?html=1` query param to see
  the raw HTML, or without to see the PNG. Use `?debug=1` to see widget cell boundaries.

### Common Pitfalls

1. **Type string mismatch** --- the Go registry and Vue registry must use the exact same
   type string. A typo means the widget renders as a placeholder in one mode.

2. **External resources in image mode** --- litehtml cannot fetch URLs. All images must be
   base64-encoded data URIs. All fonts must be registered on the renderer. CSS `url()` for
   background images only works with data URIs.

3. **CSS variables in image mode** --- litehtml does not support CSS custom properties.
   Use literal color values or template conditionals (`{{if .IsDark}}#eee{{else}}#333{{end}}`).

4. **CSS Grid in image mode** --- not supported. Use float-based layout (the master template
   already handles this) or flexbox within your widget HTML.

5. **JavaScript in image mode** --- not possible. Anything that requires timers, event
   handlers, or DOM manipulation only works in interactive mode. The clock widget, for
   example, renders the current server time as static text in image mode.

6. **Config schema divergence** --- both renderers parse the same JSON config independently.
   If you add a field to one side, add it to the other. There is no shared schema validation.

7. **Icon rendering differences** --- in Vue, theme icons load via HTTP. In image mode,
   font-based icons render as glyph spans and image-based icons render as `<img>` tags
   with file paths. Test both.

8. **Chart rendering differences** --- Vue widgets render charts as SVG for crisp scaling.
   Image widgets render charts as PNG bitmaps at fixed resolution. Design your chart
   dimensions with the target `imageConfig` size in mind.

### RenderContext

The `RenderContext` passed to Go static renderers provides dashboard-level information:

- `Theme` --- the dashboard's theme name (use for icon resolution)
- `QueryParams` --- URL query parameters (e.g. `?battery=85` for the battery widget)
- `PageIndex` --- zero-based index of the currently rendered page
- `TotalPages` --- total number of pages (used by the page-indicator widget)

The Vue equivalent is provided via Vue's dependency injection:
- `inject(DASHBOARD_THEME)` --- theme name
- `inject(DASHBOARD_ID)` --- dashboard ID
- `inject(ACTIVE_PAGE)` / `inject(TOTAL_PAGES)` --- pagination

### Data Caching and Warmup

Widgets that fetch external data benefit from the startup warmup system. On server start,
`app/router/main.go` scans all dashboards for widgets of known types and pre-fetches their
data. This means the first image render returns instantly instead of blocking on API calls.

If your widget uses a new data source, add a warmup function:
1. Scan all dashboards for your widget type
2. Extract unique query parameters from configs
3. Call the client's warmup method with all unique parameters

See `warmupWeather()` and `warmupMarket()` in `app/router/main.go` for examples.

---

## HTML Preview Mode

For debugging image dashboards, append `?html=1` to the URL to get the raw HTML output
instead of the PNG. This lets you inspect the HTML that litehtml will render, which is
invaluable when troubleshooting layout or styling issues.

Append `?debug=1` to see colored overlays on widget grid cells, making it easy to verify
width calculations and row boundaries.

---

## File Reference

### Backend (Go)

```
internal/widgets/
    registry.go                   -- Registry type + StaticRenderer interface
    {name}/
        static.go                 -- NewStaticRenderer() constructor
        static.html               -- embedded HTML template (go:embed)
        static_test.go            -- renderer tests

internal/dashboard/
    types.go                      -- Dashboard, Page, Row, Widget structs
    store.go                      -- filesystem persistence + asset management
    static/
        renderer.go               -- assembles widgets into full HTML page
        master.html               -- HTML page template (grid layout, theme styles)
    image/
        renderer.go               -- HTML to PNG via litehtml-go
        container.go              -- litehtml drawing container (fonts, shapes, gradients)

app/router/
    main.go                       -- widget registration + warmup
    middleware_static.go          -- image dashboard request handling
    api_v0.go                     -- API routes
    handlers/dashboards.go        -- dashboard CRUD API
```

### Frontend (Vue/TypeScript)

```
webui/src/
    lib/
        widgetRegistry.ts         -- Vue widget registry
        injectionKeys.ts          -- provide/inject symbols
    components/dashboards/
        DashboardWidget.vue       -- widget wrapper (dynamic component, edit controls)
        DashboardRow.vue          -- row layout
        {Name}Widget.vue          -- widget display component
        {Name}WidgetConfig.vue    -- widget config editor
    composables/
        useWeather.ts             -- weather data fetching (TanStack Query)
        useMarket.ts              -- market data fetching
        useDashboards.ts          -- dashboard CRUD
    views/dashboards/
        DashboardView.vue         -- renders interactive dashboards
        DashboardEditView.vue     -- dashboard editor
```
