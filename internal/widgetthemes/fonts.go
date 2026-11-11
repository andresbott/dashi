package widgetthemes

import (
	"fmt"
	"net/url"
	"strings"
)

// FontFamily returns the theme's first display font name. An unknown theme,
// or one that declares no fonts, yields the default theme's font.
func (s *Store) FontFamily(themeName string) string {
	_, th := s.effective(themeName, hasFonts)
	if len(th.manifest.Fonts) == 0 {
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

// IconFontFamilyFor returns the icon-font family a widget theme actually
// serves: its own when it ships an icon font, otherwise the default theme's.
// Unlike probing ResolveIcon with a sample icon, it does not depend on which
// icons the pack maps.
func (s *Store) IconFontFamilyFor(name string) string {
	effName, _ := s.effective(name, hasIconFont)
	return IconFontFamily(effName)
}

// FontData resolves a font by the name the browser stack requests it under:
// a display font declared in the manifest, or the theme's icon font under
// IconFontFamily(theme). Display fonts win a name collision, which keeps
// existing URLs stable.
//
// A font the named theme does not provide (or an unknown theme) is looked up
// in the default theme, mirroring the fallback FontFacesCSS applies.
func (s *Store) FontData(themeName, fontName string) ([]byte, error) {
	name, th := s.effective(themeName, func(t *theme) bool {
		for _, f := range t.manifest.Fonts {
			if f.Name == fontName {
				return true
			}
		}
		// The icon font is requested under the caller's theme name, so match
		// that name here; the effective name is only known after this runs.
		// The lookup below also accepts the effective name, for a request
		// that already fell back to the default.
		return hasIconFont(t) && fontName == IconFontFamily(themeName)
	})
	if data, err := s.GetDisplayFontData(name, fontName); err == nil {
		return data, nil
	}
	if hasIconFont(th) && (fontName == IconFontFamily(name) || fontName == IconFontFamily(themeName)) {
		return s.GetFontData(name)
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
	// Display fonts and the icon font fall back independently: a fonts-only
	// pack still gets the default's icon font and vice versa.
	fontsName, fonts := s.effective(themeName, hasFonts)
	for _, font := range fonts.manifest.Fonts {
		if font.Name == "" || font.File == "" {
			continue
		}
		fmt.Fprintf(b, "@font-face{font-family:'%s';src:url('/api/v0/widget-themes/%s/fonts/%s');font-display:swap;}\n",
			cssIdent(font.Name), url.PathEscape(fontsName), url.PathEscape(font.Name))
	}
	iconName, _ := s.effective(themeName, hasIconFont)
	family := IconFontFamily(iconName)
	fmt.Fprintf(b, "@font-face{font-family:'%s';src:url('/api/v0/widget-themes/%s/fonts/%s');font-display:block;}\n",
		cssIdent(family), url.PathEscape(iconName), url.PathEscape(family))
}

// FontFacesCSS renders a widget theme's fonts for the browser stack: its
// @font-face rules, followed by --dashi-font on :root when the theme has a
// display font. Colours are not a widget-theme concern; they come from the
// theme entity's stylesheet (themes.BrowserCSS).
//
// An unknown theme, or one missing fonts or an icon font, is served the
// default theme's.
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
