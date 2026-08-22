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
          → buildRenderData (theme font, custom.css, query params, page rows)
          → static.Render: widgets emit HTML fragments → master.html template
          → image.RenderToImage: litehtml-go layout → *image.RGBA
             (background image drawn directly on canvas — litehtml has no
             CSS background-image support, which is why buildBackground
             returns raw bytes separately from the CSS string)
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
"Enhanced dither" (using the library's presets/auto-suggestions in the request
path) is catalogued debt in root `TODO.md`.

## Fonts

litehtml needs fonts registered up front: `newSharedDeps` iterates all themes
and calls `imageRenderer.RegisterFont` for icon fonts (`icon-font-{theme}`) and
display fonts. A new theme font that isn't registered there simply won't render
in PNGs — no error. Fallback Inter TTFs are embedded in
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
- Frontend dev: `cd webui && npm run dev` proxies `/api` and `/auth` to
  `http://localhost:8088` (the **editor** port) — run `make run` alongside.
  Because the proxy sets `changeOrigin`, `/api/v0/info` reports the viewer at
  `http://localhost:8087`, so view links leave the dev server and hit Go.

## Analysis docs

`docs/project/widget-rendering-architectures.md` — deep comparison of the two
pipelines (its "shared `internal/weather`" style paths predate both the
2026-05-15 package merge and the 2026-08-17 move to
`internal/providers/weather`). `docs/project/litehtml-rendering-reference.md` — what CSS
litehtml actually supports. `docs/project/refactor-ideas.md` — the analyzed
(not decided) idea of moving display-mode rendering to vanilla JS.
