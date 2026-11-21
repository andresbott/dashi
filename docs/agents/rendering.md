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
| `X-Display-Width/Height` | `width`/`height` | required, positive ints |
| `X-Display-Rotation` | `rotation` | 0/90/180/270 |
| `X-Action` | `action` | `refresh` (default), `swipe_left`, `swipe_right` → page redirect |

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
  `/dashboards/:id/settings` and `/:id` (`webui/src/router/index.ts`). The
  editor falls through to `app/spa` (embedded Vue build).
- **The SPA never renders a dashboard for viewing.** The admin list's view
  action links to the server-rendered viewer, using the base URL that
  `GET /api/v0/info` advertises (`webui/src/lib/serverInfo.ts`). Its `/:id`
  route is a leftover: the Go servers intercept those paths with
  `staticMid` before the SPA ever sees them, so it only ever triggers on the
  Vite dev server.
- **E-ink preview (2026-11-21)**: the editor's "E-ink preview" dialog
  (`EinkPreviewDialog.vue`, image dashboards only) shows the PNG the server
  renders, not an SPA rendering. It `fetch`es `withBase('/{id}?format=…&width=…
  &height=…&rotation=…&page=…')` from the **editor** server, whose router
  mounts the same `staticMid` as the viewer — so no extra route exists, query
  params stand in for the display headers, and the viewer's per-dashboard
  basic auth (viewer-only) never applies. `fetch` + object URL rather than
  `<img src>` so 400 reasons can be shown and nothing is cached
  (`cache: 'no-store'`); pending autosave is flushed first. Formats offered:
  `png-bw`, `png-spectra6`, `png` (the packed formats are raw bytes). An
  interactive dashboard answers a display request with browser HTML, hence
  no button there. The Vite dev server proxies only display requests
  (`^/[a-z0-9]+\?(.*&)?format=`) to the editor port. Note the server puts no
  upper bound on width/height; the dialog clamps custom sizes to 4096.
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

## Analysis docs

`docs/project/widget-rendering-architectures.md` — deep comparison of the two
pipelines (its "shared `internal/weather`" style paths predate both the
2026-05-15 package merge and the 2026-08-17 move to
`internal/providers/weather`). `docs/project/litehtml-rendering-reference.md` — what CSS
litehtml actually supports. `docs/project/refactor-ideas.md` — the analyzed
(not decided) idea of moving display-mode rendering to vanilla JS.
