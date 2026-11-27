package search

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/andresbott/dashi/internal/widgetthemes"
)

func renderImageHTML(t *testing.T, config string, ctx widgets.RenderContext) string {
	t.Helper()
	out, err := NewModule().Renderer()(json.RawMessage(config), ctx)
	if err != nil {
		t.Fatalf("Renderer: %v", err)
	}
	return string(out)
}

func TestRenderImageDrawsAStaticBarWithPlaceholderAndLabel(t *testing.T) {
	html := renderImageHTML(t, `{"engine":"bing","placeholder":"Find it"}`, widgets.RenderContext{})

	for _, want := range []string{
		`class="widget-search"`,
		`>Find it</span>`,
		`>Bing</div>`,
		widgetthemes.IconFontFamily(widgetthemes.DefaultName),
	} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q in:\n%s", want, html)
		}
	}
	// A PNG cannot be typed into: no form controls, and no engine URL.
	for _, bad := range []string{"<form", "<input", "bing.com"} {
		if strings.Contains(html, bad) {
			t.Errorf("image render must be a static picture, found %q:\n%s", bad, html)
		}
	}
}

func TestRenderImageDefaults(t *testing.T) {
	html := renderImageHTML(t, `{}`, widgets.RenderContext{})
	if !strings.Contains(html, `>Search...</span>`) || !strings.Contains(html, `>Google</div>`) {
		t.Errorf("expected default placeholder and Google label:\n%s", html)
	}
}

// litehtml has no var(): the image template must paint with the concrete
// palette from the render context.
func TestRenderImageUsesConcretePalette(t *testing.T) {
	palette := themes.DefaultPalette("dark")
	palette.Surface = "#123456"
	palette.Primary = "#abcdef"
	html := renderImageHTML(t, `{}`, widgets.RenderContext{Palette: palette})

	for _, want := range []string{"#123456", "#abcdef", palette.TextMuted, palette.OnPrimary, palette.Border} {
		if !strings.Contains(html, want) {
			t.Errorf("missing palette colour %q in:\n%s", want, html)
		}
	}
	if strings.Contains(html, "var(") {
		t.Errorf("image render must not use CSS variables:\n%s", html)
	}
}

func TestRenderImageRejectsMalformedConfig(t *testing.T) {
	if _, err := NewModule().Renderer()(json.RawMessage(`{`), widgets.RenderContext{}); err == nil {
		t.Fatal("expected an error for malformed config")
	}
}
