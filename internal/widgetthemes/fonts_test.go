package widgetthemes

import (
	"bytes"
	"strings"
	"testing"
)

func TestFontFacesCSSDefault(t *testing.T) {
	s := NewStore("")
	css := string(s.FontFacesCSS("default"))

	if !strings.Contains(css, "/api/v0/widget-themes/default/fonts/Inter") {
		t.Errorf("missing display-font source, got:\n%s", css)
	}
	if !strings.Contains(css, "--dashi-font:'Inter'") {
		t.Errorf("missing --dashi-font, got:\n%s", css)
	}
}

func TestFontFacesCSSCarriesNoPalette(t *testing.T) {
	// Colours belong to the theme entity (internal/themes); a widget theme
	// contributes fonts only.
	css := string(NewStore("").FontFacesCSS("default"))
	if strings.Contains(css, "--dashi-bg") || strings.Contains(css, "--dashi-text") {
		t.Errorf("widget-theme CSS must not carry palette variables, got:\n%s", css)
	}
}

func TestCSSIdentStripsBreakoutCharacters(t *testing.T) {
	// The manifest is user-uploaded: a font name must not be able to
	// terminate the generated declaration.
	if got := cssIdent("Inter'; } body { display:none"); strings.ContainsAny(got, `'"\;{}<>`) {
		t.Errorf("cssIdent left breakout characters in %q", got)
	}
}

func TestFontFacesCSSEmitsFontFaceForDisplayAndIconFonts(t *testing.T) {
	// Without @font-face the browser stack can load no theme font at all:
	// --dashi-font names a family the browser never saw, and font-icon
	// themes render tofu because widgets emit font-family: icon-font-{theme}.
	s := NewStore("")
	css := string(s.FontFacesCSS("default"))

	if !strings.Contains(css, "@font-face{font-family:'Inter';src:url('/api/v0/widget-themes/default/fonts/Inter')") {
		t.Errorf("missing display-font @font-face, got:\n%s", css)
	}
	if !strings.Contains(css, "@font-face{font-family:'icon-font-default';src:url('/api/v0/widget-themes/default/fonts/icon-font-default')") {
		t.Errorf("missing icon-font @font-face, got:\n%s", css)
	}
	// The rules must precede the variable block that references the family.
	if face, root := strings.Index(css, "@font-face"), strings.Index(css, ":root{"); face > root {
		t.Errorf("@font-face rules must come before :root, got:\n%s", css)
	}
	// A format() hint would be a lie: the route serves whatever bytes the
	// theme shipped, under a fixed font/ttf Content-Type.
	if strings.Contains(css, "format(") {
		t.Errorf("no format() hint should be emitted, got:\n%s", css)
	}
}

func TestFontFacesCSSUnknownThemeEmitsNoFontFace(t *testing.T) {
	s := NewStore("")
	if css := string(s.FontFacesCSS("does-not-exist")); strings.Contains(css, "@font-face") {
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
	// internal/widgets/weather/image.go emits exactly this family name, and
	// app/router/main.go registers it with the litehtml renderer. The
	// @font-face family has to agree with both.
	if got := IconFontFamily("mytheme"); got != "icon-font-mytheme" {
		t.Errorf("IconFontFamily = %q, want %q", got, "icon-font-mytheme")
	}
}
