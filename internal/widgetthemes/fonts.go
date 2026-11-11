package widgetthemes

import (
	"fmt"
	"net/url"
	"strings"
)

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
// (see internal/widgets/weather/image.go) and what the litehtml renderer
// registers at startup (app/router/main.go), so image-stack markup rendered
// into a browser resolves the same family.
func IconFontFamily(themeName string) string {
	return "icon-font-" + themeName
}

// FontData resolves a font by the name the browser stack requests it under:
// a display font declared in the manifest, or the theme's icon font under
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
		fmt.Fprintf(b, "@font-face{font-family:'%s';src:url('/api/v0/widget-themes/%s/fonts/%s');font-display:swap;}\n",
			cssIdent(font.Name), escapedTheme, url.PathEscape(font.Name))
	}
	if ic := th.icons(); ic != nil && ic.Type == ThemeTypeFont && ic.FontFile != "" {
		family := IconFontFamily(themeName)
		fmt.Fprintf(b, "@font-face{font-family:'%s';src:url('/api/v0/widget-themes/%s/fonts/%s');font-display:block;}\n",
			cssIdent(family), escapedTheme, url.PathEscape(family))
	}
}

// FontFacesCSS renders a widget theme's fonts for the browser stack: its
// @font-face rules, followed by --dashi-font on :root when the theme has a
// display font. Colours are not a widget-theme concern; they come from the
// theme entity's stylesheet (themes.BrowserCSS).
//
// An unknown theme yields empty CSS for now; resolving unknown names
// through the default widget theme is a separate change.
func (s *Store) FontFacesCSS(themeName string) []byte {
	var b strings.Builder
	s.fontFaces(&b, themeName)
	if font := s.FontFamily(themeName); font != "" {
		fmt.Fprintf(&b, ":root{--dashi-font:'%s';}\n", cssIdent(font))
	}
	return []byte(b.String())
}

// cssIdent strips characters that could terminate a CSS string,
// declaration or block. The manifest is user-uploaded, so a font name must
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
