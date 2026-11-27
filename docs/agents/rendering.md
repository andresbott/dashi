# Rendering — the image pipeline and the display protocol

The dashboard viewer is server-rendered HTML; see [viewer.md](viewer.md) for
the browser stack. The complexity is in **image dashboards** (`type: "image"`),
which are rendered server-side into PNGs for e-ink and ESP32 displays. This
file covers that pipeline and its invariants. For widget-level rules see
[widgets.md](widgets.md).

## Pipeline (all in `app/router/middleware_static.go`)

    GET /{id}                      (single path segment, GET only)
      → StaticDashboardMiddleware: store.Get(id)
      → type != "image" → fall through (browser-HTML rendering in middleware_browser.go)
      → type == "image" + no display headers → fall through (browser HTML)
      → type == "image" + display headers present:
          parseDisplayHeaders (validate) → optional swipe redirect
          → buildRenderData (flattened theme palette, widget-theme font,
             custom.css, query params, page rows)
          → static.Render: widgets emit HTML fragments → master.html template
          → image.RenderToImage: litehtml-go layout → *image.RGBA
             (background image drawn directly on canvas — litehtml has no
             CSS background-image support, which is why themes.ImageCSS
             returns the image ref separately from the CSS string)
          → optional RotateImage (90/180/270)
          → encodeForFormat → bytes + Content-Type

## Display protocol (ESP32 / e-ink clients)

Headers take precedence over query params; either works:

| Header | Query | Values |
|---|---|---|
| `X-Display-Format` | `format` | `png`, `png-bw`, `png-spectra6`, `bw`, `spectra6` (required) |
| `X-Display-Width/Height` | `width`/`height` | required ints, 1 to 4096 (`maxDisplaySide`); anything else is a 400 naming the limit |
| `X-Display-Rotation` | `rotation` | 0/90/180/270 |
| `X-Action` | `action` | `refresh` (default), `swipe_left`, `swipe_right` → page redirect |

The size cap (2026-11-21) exists because a render allocates width×height RGBA
up front, so an unbounded size let one request exhaust memory; real panels
(largest: 13.3" Spectra 6, 1600×1200) sit far below it.

Responses set `X-Refresh-Interval` from the page's `refreshInterval`. The
packed formats (`bw`: 1bpp MSB-first; `spectra6`: 4bpp nibble codes) are raw
bytes for direct panel upload — their bit layouts are contracts with firmware;
changing them breaks deployed devices. See `internal/dashboard/image/dither.go`.

## Dithering: two-stage color pipeline (added 2026-05-01, PR #18)

`lib/einkimage` is a standalone library (full API in `lib/einkimage/README.md`).
The core invariant: **dither against calibrated colors (what the panel looks
like), then map to native device colors just before output** —
`DitherImage(...)` then `ReplaceColors(...)`. Skipping the second stage, or
dithering against native colors, produces wrong-looking panels.
`internal/dashboard/image/dither.go` wires this into the request path with a
custom `spectra6WirePalette` whose `DeviceColor` values double as wire codes.

### Dither modes per dashboard (2026-11-21)

`Dashboard.Dither` (`dither` in dashboard.json) picks the processing for every
dithered format (`png-bw`, `png-spectra6`, `bw`, `spectra6`); `png` ignores it.
`encodeForFormat` passes it to the four `Dither*` functions, which build their
options in `ditherOptions`:

| Value | Behaviour |
|---|---|
| absent / `""` | plain Floyd-Steinberg, palette only — byte-identical to the output before modes existed (pinned by `TestDitherPlainIsUnchanged`) |
| `auto` | `einkimage.SuggestProcessingOptions` on the rendered RGBA, every request; its options are used but the **target palette is kept** |
| `balanced`, `dynamic`, `vivid`, `soft`, `grayscale` | `DitherOptions.ProcessingPreset` (`einkimage.GetProcessingPresetNames`) |

Every mode still outputs palette `Color`s only, so `ReplaceColors` and the
spectra6 wire-index map keep working (`TestDitherModesOutputPaletteColors`).
The store rejects any other value on create/update/import with
`dashboard.ErrInvalidDither` (400) — `internal/dashboard/dither.go` owns the
list (`DitherModes`) and the editor mirrors it in `webui/src/lib/dither.ts`. A
value that slips past (hand-edited file) renders as plain rather than failing.

What `auto` picks for dashboards: a BW panel always gets `grayscale` (the
recommender forces it for two-colour palettes). A typical page (flat
backgrounds, text) classifies as a flat illustration, so Spectra 6 gets
`vivid` with error diffusion; photo-heavy pages (image/xkcd widgets) can land
on `dynamic` or `soft`.

Cost on 800×480 (`BenchmarkDither800x480`, a rendered text/UI page, noisy
shared machine): plain ≈ 25–30 ms; the classification `auto` adds is ≈ 2 ms
(`BenchmarkSuggestProcessingOptions800x480`); the presets cost what their
processing costs — `vivid`/`dynamic` ≈ 50–70 ms, `grayscale` ≈ 240–290 ms
(L*a*b* matching plus range compression, dominated by `math.Pow`), so `auto`
on BW ≈ 240 ms. `lib/einkimage` precomputes the palette's L*a*b* once per
image and tables the 8-bit sRGB→linear step (exact, both tested), which
halved `grayscale` on Spectra 6 (≈ 670 → 290 ms).

## Fonts

litehtml cannot fetch fonts, so they are registered with the renderer in
advance. `syncImageFonts` (`app/router/main.go`) calls
`imageRenderer.SetFonts(widgetThemeStore.AllFonts())`, which holds every
pack's icon font (`icon-font-{name}`) and display fonts (by manifest name), at
startup and again from `widgetThemeStore.OnChange` after every upload or delete.
An uploaded pack therefore paints in the next PNG with no restart. `SetFonts`
replaces the whole set, so a deleted pack's fonts are dropped and a re-upload
(delete, then upload under the same name) takes the new bytes. A font family
that isn't registered falls back to Inter with no error. Packs copied onto disk
by hand (or created with `dashi widget-theme`) are only loaded by the store at
startup, so they still need a restart.

`dashimage.Renderer` is safe for concurrent use. Fonts are parsed once when
registered and kept in a copy-on-write map, so a render already in progress
keeps the set it started with. litehtml itself is **not** reentrant:
litehtml-go's `goGetDefaultFontName` frees and reallocates a package-global C
string, so overlapping documents corrupt the heap. A package-level
`litehtmlMu` in `renderer.go` serialises document lifetimes; PNG renders run
one at a time.

## Palette in the image stack

litehtml has no CSS custom properties, so widgets read concrete colours from
`RenderContext.Palette`, built as `theme.Palette(colorMode).Flatten()`.
`Flatten` composites any `#rrggbbaa` value over `bg` into an opaque `#rrggbb`
(a translucent `surface` would otherwise reach litehtml as 8 digits). The
eleven roles are listed in [viewer.md](viewer.md#theming-two-entities-css-variables-and-concrete-values).
Image-stack widgets paint with these roles rather than hardcoded greys. Fallback Inter TTFs are embedded in
`internal/dashboard/image/`.

## Viewer and editor notes

- The viewer renders `/{id}` paths server-side (see [viewer.md](viewer.md)).
  Root `/` resolves the default dashboard server-side and redirects
  (`app/router/main.go`, viewer setup).
- The editor is a Vue SPA: root `/` redirects to `/admin`; routes are `/admin`
  (+ `admin/*` children, incl. `admin/docs/*`), `/dashboards/:id/edit`,
  `/dashboards/:id/settings`, `/dashboards/:id/eink-preview` and `/:id`
  (`webui/src/router/index.ts`). The editor falls through to `app/spa`
  (embedded Vue build).
- **The SPA never renders a dashboard for viewing.** The admin list's view
  action links to the server-rendered viewer, using the base URL that
  `GET /api/v0/info` advertises (`webui/src/lib/serverInfo.ts`). Its `/:id`
  route is a leftover: the Go servers intercept those paths with
  `staticMid` before the SPA ever sees them, so it only ever triggers on the
  Vite dev server.
- **E-ink preview (2026-11-21; dialog replaced by a page the same day)**: the
  SPA page `/dashboards/:id/eink-preview` (`DashboardEinkPreviewView.vue`,
  full-width outside `AdminLayout`) shows the PNG the server renders, not an
  SPA rendering. Layout: the image fills the left/centre, the controls
  (device preset or custom W×H, format, rotation, page, Actual pixels,
  Refresh) sit in a 300px panel on the right; below md they stack under the
  image. Two entry points, both image dashboards only: the editor toolbar
  button (flushes pending autosave, then navigates with `page` = the active
  page and `from=edit`) and a tablet icon on each image row of the admin
  dashboards list (`from=list`). An interactive dashboard opened by URL gets
  a short message and no request (it would answer with browser HTML).
  - **URL params** (all optional; parsed by `parsePreviewQuery`, written by
    `previewQuery` in `webui/src/lib/einkPreview.ts`): `preset` (a
    `DEVICE_PRESETS` key or `custom`), `width`/`height` (only with
    `custom`; clamped to 4096), `format` (`png-bw|png-spectra6|png`),
    `rotation` (`0|90|180|270`), `page` (0-based, as the display protocol),
    `actual=1`, `from` (`edit|list`). Invalid values fall back to the
    defaults (first preset, its format, 0°, page 0; a page the dashboard no
    longer has → 0). Every control change is mirrored with
    `router.replace`, so a reload or bookmark shows the same render and
    Back is unaffected; an invalid or partial query is rewritten on load.
  - **Back**: `router.back()` when `history.state.back` is set (in-app
    navigation, survives reloads); otherwise the editor for `from=edit`,
    else the dashboards list.
  - **Request**: `fetch`es `withBase('/{id}?format=…&width=…&height=…
    &rotation=…&page=…')` from the **editor** server, whose router mounts
    the same `staticMid` as the viewer — so no extra route exists, query
    params stand in for the display headers, and the viewer's per-dashboard
    basic auth (viewer-only) never applies. `fetch` + object URL rather than
    `<img src>` so 400 reasons can be shown and nothing is cached
    (`cache: 'no-store'`); control changes re-render after a 350 ms
    debounce. Formats offered: `png-bw`, `png-spectra6`, `png` (the packed
    formats are raw bytes). The Vite dev server proxies only display
    requests (`^/[a-z0-9]+\?(.*&)?format=`) to the editor port. Custom sizes
    are clamped to 4096 (`MAX_PREVIEW_SIDE`), the same limit the server
    enforces (`maxDisplaySide`), so a typo renders at the cap instead of
    failing with a 400.
  - **Upright display**: with rotation the server lays the page out in
    portrait (90/270) and `RotateImage` turns it onto the panel — 90°
    clockwise, 270° counter-clockwise, 180° flipped — so the PNG is always
    the panel's W×H. The page shows it turned back (`uprightView`: 90 →
    −90°, 180 → 180°, 270 → +90°) so it reads as on the mounted panel, with
    a caption "Shown upright — the panel receives it rotated N°". The PNG
    itself is untouched. Layout: a frame sized to the upright dimensions
    (W/H swapped for 90/270, via `--disp-w/--disp-h` CSS vars and
    `aspect-ratio`) holds the image absolutely positioned at its top-left,
    turned about that corner with a rotate + translate(−100%) so it lands
    on whole pixels. Fit mode sizes the frame with container-query units
    against the stage (`container-type: size`; `inline-size` when stacked
    below md, where it fits the width only), never upscaling; Actual pixels
    sets it to exact px and the stage scrolls (auto margins keep the
    top-left reachable). The frame keeps the shape of the image shown
    until the next render arrives.
- Frontend dev: `cd webui && npm run dev` proxies `/api` and `/auth` to
  `http://localhost:8088` (the **editor** port) — run `make run` alongside.
  Because the proxy sets `changeOrigin`, `/api/v0/info` reports the viewer at
  `http://localhost:8087`, so view links leave the dev server and hit Go.

## Row layout — column placement (12-column grid)

Both render stacks lay a row's widgets on a 12-column grid. Each `Widget` has a
`Width` (column span) and an optional `Column` (1-based start column; `0` = flow
after the previous widget, which is what pre-column dashboards carry).
`dashboard.PlaceRow` (`internal/dashboard/layout.go`) is the single resolver: it
converts each widget's column into a **leading gap** (empty columns before it)
and pushes any overlap right, so widgets never collide. Renderers apply the gap
as `margin-left: gap/12*100%` and the span as `width: width/12*100%` on the cell
— litehtml supports percentage margins
(`docs/project/litehtml-rendering-reference.md`), so the same mechanism works in
the PNG/litehtml stack and the browser. The editor mirrors `PlaceRow` in
`webui/src/lib/rowLayout.ts` and renders the gap with PrimeFlex `col-offset-N`.
A row whose spans exceed 12 columns still wraps to a second line, as before.

## Container placement (alignment)

`Container.HorizontalAlign`/`VerticalAlign` move the whole container (all
rows) inside the viewport; rows keep filling it. `dashboard.PlaceContainer`
(`internal/dashboard/alignment.go`) is the single mapping for both stacks:
left/center/right become the container's auto margin (empty or unknown =
center), top/center/bottom become `justify-content` on a full-viewport-height
flex column (empty or unknown = top). The browser shell hands the margins to
`.dashi-container` as CSS variables (inline geometry is banned there, see
[viewer.md](viewer.md#responsive-layout-mobile)). In `master.html` a plain
`.page-track` block sits between the flex column and the container because
litehtml paints nothing for a flex item with auto margins.
`TestRenderer_Render_AlignmentInLitehtml` pins the real PNG positions.

## Analysis docs

`docs/project/widget-rendering-architectures.md` — deep comparison of the two
pipelines (its "shared `internal/weather`" style paths predate both the
2026-05-15 package merge and the 2026-08-17 move to
`internal/providers/weather`). `docs/project/litehtml-rendering-reference.md` — what CSS
litehtml actually supports. `docs/project/refactor-ideas.md` — the analyzed
(not decided) idea of moving display-mode rendering to vanilla JS.
