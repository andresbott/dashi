package themes

import (
	"bytes"
	"strings"
	"testing"
)

func TestDefaultPalette(t *testing.T) {
	light := DefaultPalette("light")
	if light.FG != "#333333" {
		t.Errorf("light FG = %q, want #333333", light.FG)
	}
	if light.BG != "#ffffff" {
		t.Errorf("light BG = %q, want #ffffff", light.BG)
	}

	dark := DefaultPalette("dark")
	if dark.FG != "#e0e0e0" {
		t.Errorf("dark FG = %q, want #e0e0e0", dark.FG)
	}
	if dark.BG != "#1a1a2e" {
		t.Errorf("dark BG = %q, want #1a1a2e", dark.BG)
	}

	// Anything that is not "dark" is light.
	if DefaultPalette("").FG != light.FG {
		t.Error(`DefaultPalette("") should equal the light palette`)
	}
}

func TestPaletteMergeIgnoresBadValues(t *testing.T) {
	base := DefaultPalette("light")
	got := merge(base, map[string]string{
		"fg":      "#111111", // valid, applied
		"accent":  "red",     // not 6-digit hex, ignored
		"unknown": "#222222", // unknown key, ignored
	})
	if got.FG != "#111111" {
		t.Errorf("FG = %q, want #111111", got.FG)
	}
	if got.Accent != base.Accent {
		t.Errorf("Accent = %q, want unchanged %q", got.Accent, base.Accent)
	}
}

func TestStorePaletteUnknownThemeFallsBackToDefaults(t *testing.T) {
	s := NewStore("")
	if got := s.Palette("does-not-exist", "dark"); got != DefaultPalette("dark") {
		t.Errorf("unknown theme palette = %+v, want defaults", got)
	}
}

func TestThemeCSSEmitsBothModes(t *testing.T) {
	s := NewStore("")
	css := string(s.ThemeCSS("default"))

	if !strings.Contains(css, ":root{") {
		t.Error("missing :root block")
	}
	if !strings.Contains(css, `:root[data-color-mode="dark"]{`) {
		t.Error("missing dark-mode block")
	}
	if !strings.Contains(css, "--dashi-fg:#333333;") {
		t.Errorf("missing light --dashi-fg, got:\n%s", css)
	}
	if !strings.Contains(css, "--dashi-fg:#e0e0e0;") {
		t.Errorf("missing dark --dashi-fg, got:\n%s", css)
	}
}

func TestCSSIdentStripsBreakoutCharacters(t *testing.T) {
	// theme.yaml is user-uploaded: a font name must not be able to
	// terminate the generated declaration.
	if got := cssIdent("Inter'; } body { display:none"); strings.ContainsAny(got, `'"\;{}<>`) {
		t.Errorf("cssIdent left breakout characters in %q", got)
	}
}

func TestThemeCSSEmitsFontFaceForDisplayAndIconFonts(t *testing.T) {
	// Without @font-face the browser stack can load no theme font at all:
	// --dashi-font names a family the browser never saw, and font-icon
	// themes render tofu because widgets emit font-family: icon-font-{theme}.
	s := NewStore("")
	css := string(s.ThemeCSS("default"))

	if !strings.Contains(css, "@font-face{font-family:'Inter';src:url('/api/v0/themes/default/fonts/Inter')") {
		t.Errorf("missing display-font @font-face, got:\n%s", css)
	}
	if !strings.Contains(css, "@font-face{font-family:'icon-font-default';src:url('/api/v0/themes/default/fonts/icon-font-default')") {
		t.Errorf("missing icon-font @font-face, got:\n%s", css)
	}
	// The rules must precede the variable blocks that reference the family.
	if face, root := strings.Index(css, "@font-face"), strings.Index(css, ":root{"); face > root {
		t.Errorf("@font-face rules must come before :root, got:\n%s", css)
	}
	if !strings.Contains(css, "--dashi-font:'Inter';") {
		t.Errorf("missing --dashi-font, got:\n%s", css)
	}
	// A format() hint would be a lie: the route serves whatever bytes the
	// theme shipped, under a fixed font/ttf Content-Type.
	if strings.Contains(css, "format(") {
		t.Errorf("no format() hint should be emitted, got:\n%s", css)
	}
}

func TestThemeCSSUnknownThemeEmitsNoFontFace(t *testing.T) {
	s := NewStore("")
	if css := string(s.ThemeCSS("does-not-exist")); strings.Contains(css, "@font-face") {
		t.Errorf("unknown theme must emit no @font-face, got:\n%s", css)
	}
}

func TestFontDataResolvesDisplayAndIconFonts(t *testing.T) {
	s := NewStore("")

	display, err := s.FontData("default", "Inter")
	if err != nil {
		t.Fatalf("FontData(display): %v", err)
	}
	if len(display) == 0 {
		t.Error("display font data is empty")
	}

	icon, err := s.FontData("default", IconFontFamily("default"))
	if err != nil {
		t.Fatalf("FontData(icon): %v", err)
	}
	if len(icon) == 0 {
		t.Error("icon font data is empty")
	}
	if bytes.Equal(display, icon) {
		t.Error("icon font resolved to the display font")
	}

	if _, err := s.FontData("default", "NoSuchFont"); err == nil {
		t.Error("expected an error for an unknown font name")
	}
}

func TestIconFontFamilyMatchesWidgetMarkup(t *testing.T) {
	// internal/widgets/weather/static.go emits exactly this family name, and
	// app/router/main.go registers it with the litehtml renderer. The
	// @font-face family has to agree with both.
	if got := IconFontFamily("mytheme"); got != "icon-font-mytheme" {
		t.Errorf("IconFontFamily = %q, want %q", got, "icon-font-mytheme")
	}
}
