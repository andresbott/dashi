package themes

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Palette is the closed set of theme colours. Every field is guaranteed
// non-empty: keys a theme.yaml omits fall back to DefaultPalette, whose
// values are the ones internal/dashboard/static/master.html hardcoded
// before themes owned colours.
//
// The browser stack consumes these as CSS custom properties (see
// ThemeCSS); the image stack consumes concrete values, because litehtml
// cannot read custom properties.
type Palette struct {
	FG       string // primary text
	Muted    string // secondary text
	BG       string // page background
	Surface  string // raised surfaces: cards, bars
	Border   string // hairlines
	Accent   string // interactive / highlight
	AccentFG string // text drawn on top of Accent
	Warn     string
	Danger   string
}

var paletteLight = Palette{
	FG: "#333333", Muted: "#666666", BG: "#ffffff", Surface: "#f5f5f7",
	Border: "#e0e0e5", Accent: "#0a84ff", AccentFG: "#ffffff",
	Warn: "#d97706", Danger: "#dc2626",
}

var paletteDark = Palette{
	FG: "#e0e0e0", Muted: "#999999", BG: "#1a1a2e", Surface: "#23233d",
	Border: "#33334d", Accent: "#4a90d9", AccentFG: "#ffffff",
	Warn: "#f59e0b", Danger: "#ef4444",
}

// DefaultPalette returns the built-in palette for a colour mode. Any
// value other than "dark" is treated as light.
func DefaultPalette(colorMode string) Palette {
	if colorMode == "dark" {
		return paletteDark
	}
	return paletteLight
}

var hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// set assigns a manifest colour key, reporting whether the key is known.
func (p *Palette) set(key, value string) bool {
	switch key {
	case "fg":
		p.FG = value
	case "muted":
		p.Muted = value
	case "bg":
		p.BG = value
	case "surface":
		p.Surface = value
	case "border":
		p.Border = value
	case "accent":
		p.Accent = value
	case "accentFg":
		p.AccentFG = value
	case "warn":
		p.Warn = value
	case "danger":
		p.Danger = value
	default:
		return false
	}
	return true
}

// merge applies manifest overrides onto base. Unknown keys and values
// that are not 6-digit hex are ignored — theme loading elsewhere in this
// package is lenient rather than fatal, and a malformed colour should
// not take a dashboard down.
func merge(base Palette, over map[string]string) Palette {
	out := base
	for key, value := range over {
		if !hexColorRe.MatchString(value) {
			continue
		}
		out.set(key, value)
	}
	return out
}

// Palette returns the resolved palette for a theme and colour mode.
// Unknown themes get the built-in defaults.
func (s *Store) Palette(themeName, colorMode string) Palette {
	base := DefaultPalette(colorMode)
	th, ok := s.lookupTheme(themeName)
	if !ok {
		return base
	}
	if colorMode == "dark" {
		return merge(base, th.manifest.ColorsDark)
	}
	return merge(base, th.manifest.Colors)
}

// FontFamily returns the theme's first display font name, or "" when the
// theme declares none.
func (s *Store) FontFamily(themeName string) string {
	th, ok := s.lookupTheme(themeName)
	if !ok || len(th.manifest.Fonts) == 0 {
		return ""
	}
	return th.manifest.Fonts[0].Name
}

// IconFontFamily returns the CSS font-family name under which a theme's
// icon font is registered. It matches what the widget templates emit
// (see internal/widgets/weather/static.go) and what the litehtml renderer
// registers at startup (app/router/main.go), so image-stack markup rendered
// into a browser resolves the same family.
func IconFontFamily(themeName string) string {
	return "icon-font-" + themeName
}

// FontData resolves a font by the name the browser stack requests it under:
// a display font declared in theme.yaml, or the theme's icon font under
// IconFontFamily(theme). Display fonts win a name collision, which keeps
// existing URLs stable.
func (s *Store) FontData(themeName, fontName string) ([]byte, error) {
	if data, err := s.GetDisplayFontData(themeName, fontName); err == nil {
		return data, nil
	}
	if fontName == IconFontFamily(themeName) {
		return s.GetFontData(themeName)
	}
	return nil, fmt.Errorf("font %q not found in theme %q", fontName, themeName)
}

// fontFaces writes an @font-face rule per font the theme serves: its
// display fonts, plus its icon font when the theme uses one.
//
// Without these, the browser stack has no way to load a theme font at all:
// --dashi-font would name a family the browser has never seen, and
// font-icon themes would render tofu, because widgets emit
// `font-family: icon-font-{theme}` — a family that otherwise only exists
// inside the litehtml renderer.
//
// No format() hint is emitted on purpose. The route serves whatever bytes
// the theme shipped under a fixed font/ttf Content-Type; browsers sniff the
// actual format, so claiming truetype would break a theme shipping woff2.
func (s *Store) fontFaces(b *strings.Builder, themeName string) {
	th, ok := s.lookupTheme(themeName)
	if !ok {
		return
	}
	escapedTheme := url.PathEscape(themeName)
	for _, font := range th.manifest.Fonts {
		if font.Name == "" || font.File == "" {
			continue
		}
		fmt.Fprintf(b, "@font-face{font-family:'%s';src:url('/api/v0/themes/%s/fonts/%s');font-display:swap;}\n",
			cssIdent(font.Name), escapedTheme, url.PathEscape(font.Name))
	}
	if ic := th.icons(); ic != nil && ic.Type == ThemeTypeFont && ic.FontFile != "" {
		family := IconFontFamily(themeName)
		fmt.Fprintf(b, "@font-face{font-family:'%s';src:url('/api/v0/themes/%s/fonts/%s');font-display:block;}\n",
			cssIdent(family), escapedTheme, url.PathEscape(family))
	}
}

// ThemeCSS renders a theme as CSS custom properties for the browser
// stack: light values on :root, dark values under the
// [data-color-mode="dark"] attribute the page shell sets on <html>. It also
// carries the theme's @font-face rules, since the browser has no other way
// to obtain a theme's fonts.
func (s *Store) ThemeCSS(themeName string) []byte {
	var b strings.Builder

	s.fontFaces(&b, themeName)

	b.WriteString(":root{")
	writePaletteVars(&b, s.Palette(themeName, "light"))
	if font := s.FontFamily(themeName); font != "" {
		fmt.Fprintf(&b, "--dashi-font:'%s';", cssIdent(font))
	}
	b.WriteString("}\n")

	b.WriteString(`:root[data-color-mode="dark"]{`)
	writePaletteVars(&b, s.Palette(themeName, "dark"))
	b.WriteString("}\n")

	return []byte(b.String())
}

func writePaletteVars(b *strings.Builder, p Palette) {
	fmt.Fprintf(b, "--dashi-fg:%s;", p.FG)
	fmt.Fprintf(b, "--dashi-muted:%s;", p.Muted)
	fmt.Fprintf(b, "--dashi-bg:%s;", p.BG)
	fmt.Fprintf(b, "--dashi-surface:%s;", p.Surface)
	fmt.Fprintf(b, "--dashi-border:%s;", p.Border)
	fmt.Fprintf(b, "--dashi-accent:%s;", p.Accent)
	fmt.Fprintf(b, "--dashi-accent-fg:%s;", p.AccentFG)
	fmt.Fprintf(b, "--dashi-warn:%s;", p.Warn)
	fmt.Fprintf(b, "--dashi-danger:%s;", p.Danger)
}

// cssIdent strips characters that could terminate a CSS string,
// declaration or block. theme.yaml is user-uploaded, so a font name must
// not be able to inject arbitrary CSS into the generated stylesheet.
func cssIdent(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\'', '"', '\\', ';', '{', '}', '<', '>', '\n', '\r':
			return -1
		}
		return r
	}, s)
}
