# CSS variables in the litehtml (PNG) stack

Status: exploration, 2026-11-21. No code changed. Backlog item: "litehtml
cannot read CSS vars, explore a CSS library that transforms modern CSS into
plain CSS (like SCSS)".

## TL;DR: the premise is wrong

**The litehtml we ship already supports CSS custom properties.** The claim
"litehtml does NOT support CSS custom properties" (in
`docs/project/litehtml-rendering-reference.md:270`, `docs/agents/viewer.md:379`,
`docs/agents/rendering.md:112`, and code comments) is out of date. The analysis
came from an old test-suite snapshot. The vendored C++ (litehtml-go v0.2.0,
`vendor/github.com/andresbott/litehtml-go/litehtml/src/style.cpp:143-182,
1595-1671`, `html_tag.cpp:326-352`) parses `--x` declarations, defers any
declaration that contains `var()`, and substitutes values per element in
`compute_styles`. The lookup walks up the parent chain (inheritance), supports
fallbacks, and detects cycles. It also parses `#rrggbbaa`
(`web_color.cpp:174-210`), and Dashi's `pngContainer` composites alpha with
`draw.Over`.

So no CSS transform library and no Go resolver are needed. The recommendation
is option **(f)** below: emit a `:root{--dashi-*}` block into `master.html` and
switch the image templates to `var()`. That removes about 20 concrete emitters.

## Verification (pixel probe)

A scratch program used copies of `internal/dashboard/image/{renderer,container}.go`
and the same litehtml-go v0.2.0 from the module cache. It lives at
`/tmp/litehtml-var-probe`, outside the repo. It rendered 100×20 strips and
read back the pixels:

| Case | Result |
|---|---|
| `:root{--red:#f00}` → `background:var(--red)` | works |
| override under `html[data-mode="dark"]{…}` | works (attribute-selector theming is possible) |
| `var(--missing, #0f0)` fallback; `var(--missing, var(--red))` | works |
| nested `--a: var(--b)` | works |
| custom property set inline on a parent (`style="--c:…"`) and inherited | works |
| `var()` in shorthands (`border-bottom`, `padding`), in `color`, in `font-family` (parses) | works |
| whole gradient in a var (`background:var(--grad)`); gradient with one or two **different** vars | works |
| **same var twice in one declaration** (`linear-gradient(var(--x),var(--x))`, `padding:var(--p) var(--p)`) | **declaration dropped**: bug in the vendored snapshot, see below |
| undefined var, no fallback | declaration dropped (CSS "invalid at computed-value time"), silently |
| `#ff000080`, `rgba()`, `#rrggbbaa` gradient stops over blue | composited correctly (128,0,127) |
| `calc(10px + 10px)` | **unsupported** (ignored) |
| `color-mix()` | **unsupported** |

**Repeated-var bug.** In the vendored `subst_var` (`style.cpp:1616-1649`),
`used_vars` accumulates for the whole declaration. A second reference to the
same property therefore looks like a cycle. Upstream master fixes this: it
erases the name after resolving it, so it tracks only the active resolution
path. It also marks a failed property `invalid` explicitly.

**What the vendored snapshot lacks** (upstream status checked 2026-11 in
`litehtml/litehtml` `doc/css_support.md` and the commit log). The vendored
snapshot dates from around 2025: it has `text-emphasis` (2025-03) but not
mid/late-2026 work.

| Feature | Vendored | Upstream master (Nov 2026) |
|---|---|---|
| `var()` / custom properties | yes (repeated-var bug) | yes (fixed) |
| `calc()`, `min()`, `max()`, `clamp()` | no | yes, for lengths/percentages only (2026-11-12) |
| flex `gap` | no | yes (2026-11-13) |
| CSS grid | no | partial (2026-11-11) |
| `:has()`, `:where()`, `@supports`, `@layer` | no | yes (`@layer` ordering not implemented) |
| `color-mix()`, `opacity`, `box-shadow`, `transform` | no | no |

## Current state: where values are duplicated because of the old belief

| Place | What it emits for litehtml | Browser equivalent |
|---|---|---|
| `internal/dashboard/static/master.html:11-12` | `{{.Palette.BG}}`, `{{.Palette.Text}}` on `body` | `viewer.css` `var(--dashi-bg)` / `--dashi-text` |
| 7 image templates, **20 refs**: `weather/image.html` (9), `swisstransport/image.html` (4), `market/image.html` (3), `pageindicator/pageindicator.html` (2), `bookmark/bookmark.html` (1), `clock/image.html` (1) | `{{.Palette.TextMuted}}`, `{{.Palette.Border}}` | `clock/browser.html`, `sysinfo/browser.html` use `var(--dashi-*)` |
| 6 Go template-data structs carrying `Palette` (`bookmark.go:81`, `clock/image.go:33`, `market/image.go:42`, `pageindicator.go:26`, `swisstransport/image.go:43`, `weather/image.go:103`) via `RenderContext.EffectivePalette()` | plumbing only | none |
| `themes.Palette.Flatten` (`internal/themes/palette.go:116-150`), called in `app/router/middleware_static.go:359` | composites `#rrggbbaa` over `bg` | browser gets raw alpha |
| `themes.ImageCSS` (`internal/themes/css.go:40-62`, commit c061525) | gradient with stops flattened over `bg` | `--dashi-page-bg` |
| `master.html:10` `FontFamily` | concrete family | `--dashi-font` |
| `market/chart.go:115` `paletteLineColors` | raster colours for a PNG chart | n/a. **Not CSS; stays as is.** |

Unported widgets fall back to their image template in the browser
(`registry.go:106-111`). They get concrete, unflattened values there today. If
the templates use `var()`, the browser resolves them from the theme CSS that
the page already links.

## Options

| | Option | Verdict |
|---|---|---|
| a | **Go pass substituting `var(--x)`** from `:root`/theme values before litehtml (tokenizer: `tdewolff/parse/v2/css`, MIT; or `gorilla/css`, BSD-3; `douceur` is MIT but unmaintained) | **Redundant.** It would duplicate what litehtml already does, but worse: a string pass has no DOM, so it only works for `:root` scope, with no per-element overrides or inheritance. It also adds a dependency and a per-render parse. The only gap it could cover is the repeated-var bug, and a 10-line fix in litehtml-go covers that more cheaply. |
| b | **Build-time tools**: lightningcss (MPL-2.0, Rust), PostCSS `postcss-custom-properties` (MIT, Node), esbuild (MIT; its Go API could even run at runtime) | **Doesn't help.** Theme values are user-edited at runtime. The PostCSS plugin resolves only `:root` vars known at build time, and lightningcss/esbuild do not substitute custom properties at all. Their real feature, lowering syntax (nesting, `#rrggbbaa`→`rgba`, modern colours), targets things litehtml already parses or Dashi doesn't write. They also can't add `calc()` or `gap` semantics, which need a layout engine. |
| c | **Keep the concrete emitters** (status quo) | Works, but keeps the duplicated vocabulary: image templates speak `{{.Palette.X}}`, browser templates speak `var(--dashi-x)`. It also means `custom.css` using `var(--dashi-*)` silently does nothing in PNGs, and every new widget must thread `Palette`. |
| d | Go CSS libraries (as listed in a) | Only useful as the tokenizer for (a). Not needed. |
| e | **Bump litehtml-go's vendored litehtml to upstream master** | Worth doing **independently**. It fixes the repeated-var bug and adds `calc()`/`min()`/`max()`/`clamp()`, flex `gap`, partial grid and `:has()`. It is a big jump (pixel_t class, intrinsic sizing, grid, line-box changes) that needs bridge updates per the litehtml-go `CLAUDE.md` procedure and PNG pixel-diff review. Smaller alternative: cherry-pick only the `subst_var` fix. |
| f | **Use litehtml's native `var()`**: emit `:root{--dashi-*}` for the dashboard's colour mode in `master.html` and write `var(--dashi-…)` in image templates | **Recommended.** No new dependency, and one variable vocabulary across both stacks. |

## Recommendation

Do **(f)** now and treat **(e)** as a separate follow-up in litehtml-go. Drop
the idea of a CSS-transform library.

Migration sketch for (f), about 0.5–1 day:

1. Add a `themes` helper that writes `:root{…}` for a single colour mode by
   reusing `writePaletteVars`. Emit it first in `master.html`'s `<style>`
   (before widget CSS and `custom.css`). The server knows the mode, so no
   `[data-color-mode]` selector is needed.
2. Replace `{{.Palette.BG}}`/`{{.Palette.Text}}` in `master.html` and the 20
   template refs with `var(--dashi-bg)` etc. Remove `Palette` from the 6
   template-data structs. Keep `RenderContext.Palette`/`EffectivePalette` for
   `market/chart.go`.
3. Decide on `Flatten`. litehtml composites alpha correctly, so a translucent
   `surface` would blend over what is actually painted (gradient or image), as
   in the browser, rather than over `bg`. Output is identical on a plain `bg`
   and more faithful otherwise. Dithering works on the final RGBA, so e-ink
   doesn't care. Recommend dropping `Flatten` from `buildRenderData` and
   accepting the visible diff on gradient or image themes. Keep the
   `ImageCSS` gradient-stop flattening for now: the body gradient sits over a
   white canvas, not `bg`. Revisit with a `html{background:var(--dashi-bg)}`
   pixel test.
4. Template rule until (e) lands: **never reference the same `--dashi-*` twice
   in one declaration.** None of the current refs do (all are single
   `color:`/`border:` values).
5. Fix the stale statements: `litehtml-rendering-reference.md:270`,
   `viewer.md:379-386`, `rendering.md:110-118`, and the comments in
   `widgets/registry.go:17-19`, `static/renderer.go:45-46`,
   `middleware_static.go:330-331,363-364`, `themes/css.go:40-51`,
   `themes/palette.go:116-118`.

Bonus: `custom.css` written against `var(--dashi-primary)` starts working in
PNGs, matching the browser.

**Risks.** A typo in a var name, or an unset var, drops the declaration
silently with no error. litehtml-go bumps could regress substitution. The
repeated-var trap stays in place until (e). Perf: substitution runs once per
element with a `var()`, plus a parent-chain lookup. That is negligible at
dashboard size; check with the existing render benchmark.

**Tests to pin it:**

- `internal/dashboard/image`: pixel test for `:root` var, fallback,
  inheritance, alpha compositing. Also one test asserting the
  repeated-var-drops behaviour, which should flip (and be updated) when
  litehtml is bumped.
- `internal/dashboard/static`: `master.html` emits all eleven `--dashi-*` for
  the dashboard's mode (dark mode gets dark values) ahead of widget CSS.
- Contract test over `internal/widgets/*/*.html`: every `var(--dashi-…)` names
  a role in `paletteRoles`, no declaration repeats a var, and no
  `{{.Palette.` remains in image templates.
- `app/router`: an image render of a dashboard whose `custom.css` uses
  `var(--dashi-primary)` paints that colour. This replaces
  `TestBuildRenderDataFlattensPalette` if `Flatten` is dropped.
