# Litehtml Rendering Reference

What CSS/HTML features are supported by litehtml, based on analysis of the litehtml test suite
(~9,200 test cases: 86 top-level, 8,544 CSS 2.1, 577 Flexbox).

Tests are pixel-perfect visual regression tests: each HTML file is rendered to an 800x600 PNG at
96 DPI and compared against a reference image. Files prefixed with `-` are disabled/unsupported.

**Test suite analyzed:** https://github.com/nicktrandafil/litehtml-tests (or local copy at
`/home/odo/tmp/litehtml-tests/`)

---

## Test Suite Overview

| Category | Total | Enabled | Disabled | Coverage |
|----------|-------|---------|----------|----------|
| Top-level tests | 86 | 82 | 4 | Core features, gradients, tables, selectors |
| CSS 2.1 tests | 8,544 | ~4,784 | ~3,760 | W3C CSS 2.1 compliance |
| Flexbox tests | 577 | ~408 | ~169 | CSS Flexbox Level 1 |
| **Total** | **~9,207** | **~5,274** | **~3,933** | |

Reference PNGs available: **5,995** (ground truth for pixel comparison).

---

## Supported Features (Enabled Tests Pass)

### Box Model & Dimensions

Fully tested and working:

- `width`, `height` (fixed, percentage, auto)
- `min-width`, `min-height`, `max-width`, `max-height`
- `margin` (all sides, shorthand, percentage, auto centering)
- `padding` (all sides, shorthand 1-4 values, percentage)
- `border` (width, style, color, shorthand per side)
- `box-sizing: border-box` and `content-box`
- Margin collapsing between adjacent blocks
- Block, inline, inline-block, and replaced element sizing

### Display Modes

- `display: block`
- `display: inline`
- `display: inline-block`
- `display: inline-table`
- `display: table`, `table-row`, `table-cell`, `table-header-group`, `table-footer-group`
- `display: list-item`
- `display: flex`, `display: inline-flex`
- `display: none`
- Anonymous box generation

### Positioning

- `position: static` (default)
- `position: relative` with `top`, `left`, `bottom`, `right` offsets
- `position: absolute` with containing block resolution
- `position: fixed`
- `z-index` stacking order (including with absolute positioning)
- Absolute positioned element sizing (non-replaced and replaced)

### Floats & Clear

- `float: left`, `float: right`, `float: none`
- `clear: left`, `clear: right`, `clear: both`, `clear: none`
- Float placement, adjacent floats, float wrapping
- Floats with block formatting context
- Float interaction with navigation menus, tables

### Backgrounds

- `background-color`
- `background-image` (URL-based)
- `background-repeat`, `background-position`, `background-attachment`
- `background-size`
- Multiple background layers
- Body element background inheritance to canvas
- **Linear gradients:** `linear-gradient()` with directions, color stops
- **Radial gradients:** `radial-gradient()` with color stops
- **Conic gradients:** `conic-gradient()`
- **Repeating gradients:** `repeating-linear-gradient()`, `repeating-radial-gradient()`, `repeating-conic-gradient()`

### Text & Typography

- `font-family`, `font-size`, `font-weight`, `font-style`, `font-variant`
- `line-height` (unitless, relative, absolute)
- `text-align: left | center | right | justify`
- `text-decoration` (underline, overline, line-through, with styles: solid, dashed, dotted, double, wavy, and colors/thicknesses)
- `text-transform: uppercase | lowercase | capitalize`
- `text-indent`
- `white-space: normal | nowrap | pre | pre-wrap | pre-line`
- `vertical-align` (baseline, super, sub, top, middle, bottom, text-top, text-bottom)
- Whitespace collapsing and processing

### Colors & Units

- Named colors, hex (`#rgb`, `#rrggbb`), `rgb()`, `rgba()`
- Transparent backgrounds
- Alpha channel support
- Invalid color handling (graceful fallback)
- CSS units: `px`, `em`, `%`, `pt`, `cm`, `mm`, `in`, `pc`, `ex`, `ch`, `rem`, `vw`, `vh`

### Selectors

All standard CSS selectors work:

- **Basic:** type, class (`.class`), ID (`#id`), universal (`*`)
- **Combinators:** descendant (` `), child (`>`), adjacent sibling (`+`), general sibling (`~`)
- **Attribute:** `[attr]`, `[attr=val]`, `[attr~=val]`, `[attr|=val]`, `[attr^=val]`, `[attr$=val]`, `[attr*=val]`, case-sensitivity modifier (`[attr=val s]`)
- **Pseudo-classes:** `:first-child`, `:last-child`, `:nth-child()`, `:nth-of-type()`, `:not()`, `:is()`, `:link`, `:visited`, `:hover`, `:active`, `:focus`, `:lang()`
- **Pseudo-elements:** `::before`, `::after`, `::first-line`, `::first-letter`
- Specificity and cascade precedence rules
- Case sensitivity in quirks mode

### Generated Content

- `content` property with text strings, `attr()`, `url()`, `counter()`, `counters()`
- `counter-reset`, `counter-increment`
- Nested counters with `counters()` function
- `::before` and `::after` pseudo-elements
- Generated content with floats and absolute positioning
- `quotes` property (partial, some tests disabled)

### Tables

- `table-layout: auto` and `table-layout: fixed`
- Width calculations (table, column, cell level)
- `border-collapse: collapse` and `border-spacing`
- Table backgrounds (both collapse and separated border models)
- `caption-side: top | bottom`
- `empty-cells: show | hide`
- Table vertical alignment
- Anonymous table element generation (partial)

### Lists

- `list-style-type` (disc, circle, square, decimal, lower-alpha, upper-alpha, lower-roman, upper-roman, lower-greek, armenian, georgian)
- `list-style-image`
- `list-style-position: inside | outside`
- List backgrounds

### Flexbox

Comprehensive flexbox support:

- **Container:** `display: flex`, `display: inline-flex`
- **Direction:** `flex-direction: row | row-reverse | column | column-reverse`
- **Wrapping:** `flex-wrap: wrap | nowrap | wrap-reverse`
- **Shorthand:** `flex-flow`
- **Item sizing:** `flex-basis`, `flex-grow`, `flex-shrink`, `flex` shorthand
- **Main axis alignment:** `justify-content: flex-start | flex-end | center | space-between | space-around | space-evenly`
- **Cross axis alignment:** `align-items: flex-start | flex-end | center | baseline | stretch`
- **Individual alignment:** `align-self`
- **Multi-line alignment:** `align-content: flex-start | flex-end | center | space-between | space-around | stretch`
- **Ordering:** `order` property
- **Min sizing:** `min-width: auto`, `min-height: auto` in flex context
- Absolutely positioned children in flex containers
- Flex items as stacking contexts
- Flex with `::before`/`::after` pseudo-elements
- `margin: auto` centering in flex
- `box-sizing` interaction with flex

### Media Queries

- `@media` rules with `min-width`, `max-width`, `min-height`, `max-height`
- `orientation: portrait | landscape`
- `aspect-ratio`
- `color`
- `resolution`
- Boolean operators: AND, OR (comma), NOT

### @-Rules

- `@import` (including nested imports)
- `@media`
- `@charset` (partial, some tests disabled)

### Visibility & Overflow

- `visibility: visible | hidden | collapse`
- `overflow: visible | hidden | scroll | auto`
- Clipping regions (partial)

### Inheritance

- Property inheritance follows CSS spec
- `inherit` keyword
- Correct inherited vs non-inherited property defaults

---

## Known Limitations (Disabled Tests)

Features with significant numbers of disabled tests, indicating partial or no support:

### Not Supported / Partial Support

| Feature | Disabled Tests | Notes |
|---------|---------------|-------|
| `::first-letter` with punctuation | ~406 | Unicode punctuation handling with `::first-letter` |
| Outline properties | ~183 | `outline-color`, `outline-width`, `outline-style` |
| Table border conflict resolution | ~205 | Complex `border-collapse` conflict rules |
| Bidirectional text (RTL/LTR) | ~87 | `direction: rtl`, `unicode-bidi` interactions |
| Advanced counter/quote generation | ~140 | Complex counter scoping, automatic quotes |
| `clip` property | ~44 | CSS `clip` for absolute positioned elements |
| Page break properties | ~13 | `page-break-before`, `page-break-after` (print) |
| `cursor` property | ~19 | CSS cursor styling |
| Advanced `letter-spacing` / `word-spacing` | ~23 | Edge cases in character/word spacing |
| Table anonymous objects | ~66 | Anonymous table element generation edge cases |
| Flexbox writing modes | ~16 | Flex with `writing-mode: vertical-lr/rl` |
| Flexbox aspect-ratio | ~4 | `aspect-ratio` property in flex context |
| Flexbox `flex-basis: content` | ~5 | Content-based flex basis |
| Tables as flex items | ~8 | Table elements used as flex items |
| Zero-radius radial gradient | 1 | Degenerate radial gradient (solid fill) |
| Conic gradient color hints | 1 | Gradient midpoint hints |

### Specific Disabled Top-Level Tests

| Test | Feature | Reason |
|------|---------|--------|
| `-radial-gradient-2.htm` | Radial gradients | Zero-radius radial gradient rendering |
| `-conic-gradient-hints.htm` | Conic gradients | Color hint interpolation (max_color_diff: 34) |
| `-issue-419-3.html` | Flex-shrink | flex-shrink with space-around justify-content |
| `-table-caption.htm` | Table captions | inline-table with caption and text-transform |

---

## Practical Guidelines for Dashi

When generating HTML/CSS for litehtml rendering in Dashi dashboards:

### Safe to Use

1. **Layout:** Block/inline/inline-block, floats, absolute/relative positioning, flexbox (row/column with standard alignment)
2. **Sizing:** Fixed px, percentages, em/rem, viewport units
3. **Box model:** margin, padding, border, box-sizing
4. **Backgrounds:** Solid colors, images, linear/radial/conic gradients
5. **Text:** All standard font properties, text-align, text-decoration, line-height, white-space
6. **Colors:** Named, hex, rgb(), rgba(), transparent
7. **Selectors:** All standard selectors including `:nth-child`, `:not()`, `:is()`, attribute selectors
8. **Generated content:** `::before`/`::after` with text, counters, attr()
9. **Tables:** Standard table layout with border-collapse or separate borders
10. **Lists:** All list-style-type values, list-style-position
11. **Media queries:** Width, height, orientation based

### Avoid or Use with Caution

1. **Outlines:** `outline` properties are largely unsupported
2. **RTL/bidirectional text:** Limited support for `direction: rtl` and `unicode-bidi`
3. **`::first-letter` with punctuation:** Complex punctuation handling fails
4. **`clip` property:** Not fully supported for absolute elements
5. **Page/print styles:** `page-break-*` properties not supported
6. **Complex table border conflicts:** Advanced border-collapse conflict resolution may fail
7. **Flexbox with writing modes:** Avoid mixing flex with `writing-mode: vertical-*`
8. **`flex-basis: content`:** Not supported, use `auto` instead
9. **Tables as flex items:** Avoid placing `<table>` elements directly as flex children
10. **`cursor` property:** Not supported (irrelevant for image rendering)
11. **Advanced counter scoping:** Keep counter usage simple, avoid deeply nested counter scopes
12. **Gradient color hints:** Midpoint color hints in gradients may render incorrectly

### Rendering Notes

- Render canvas is 800x600 at 96 DPI in tests (actual size may vary in Dashi)
- Rendering uses Cairo + Pango for text, so font rendering follows those libraries
- Pixel-perfect accuracy is expected; avoid features that produce approximate results
- litehtml does NOT support JavaScript -- all styling must be pure CSS
- litehtml does NOT support CSS Grid -- use flexbox or floats for layout
- litehtml does NOT support CSS animations or transitions
- litehtml does NOT support CSS custom properties (variables)
