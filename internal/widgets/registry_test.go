package widgets

import (
	"context"
	"encoding/json"
	"html/template"
	"testing"

	"github.com/andresbott/dashi/internal/themes"
)

func TestRegistry_Render_KnownType(t *testing.T) {
	reg := NewRegistry()
	reg.Register("test", func(config json.RawMessage, _ RenderContext) (template.HTML, error) {
		return template.HTML("<p>hello</p>"), nil
	})

	got, err := reg.Render("test", json.RawMessage(`{}`), RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := template.HTML("<p>hello</p>")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRegistry_Render_UnknownType(t *testing.T) {
	reg := NewRegistry()

	got, err := reg.Render("nonexistent", json.RawMessage(`{}`), RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := template.HTML("")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNoopModule_DoesNotPanic(t *testing.T) {
	var n NoopModule
	n.RegisterRoutes(nil)
	n.Warmup(context.Background(), nil)
}

func TestRenderBrowserUsesBrowserRenderer(t *testing.T) {
	r := NewRegistry()
	r.Register("demo", func(json.RawMessage, RenderContext) (template.HTML, error) {
		return "<b>image</b>", nil
	})
	r.RegisterBrowser("demo", func(json.RawMessage, RenderContext) (template.HTML, error) {
		return "<b>browser</b>", nil
	})

	got, err := r.RenderBrowser("demo", nil, RenderContext{})
	if err != nil {
		t.Fatalf("RenderBrowser: %v", err)
	}
	if got != "<b>browser</b>" {
		t.Errorf("got %q, want the browser renderer output", got)
	}
}

func TestRenderBrowserFallsBackToImageRenderer(t *testing.T) {
	// An unported widget has no browser renderer; the browser view must
	// still show something rather than nothing.
	r := NewRegistry()
	r.Register("demo", func(json.RawMessage, RenderContext) (template.HTML, error) {
		return "<b>image</b>", nil
	})

	got, err := r.RenderBrowser("demo", nil, RenderContext{})
	if err != nil {
		t.Fatalf("RenderBrowser: %v", err)
	}
	if got != "<b>image</b>" {
		t.Errorf("got %q, want the image renderer output as fallback", got)
	}
}

func TestRenderBrowserUnknownTypeYieldsEmpty(t *testing.T) {
	r := NewRegistry()
	got, err := r.RenderBrowser("nope", nil, RenderContext{})
	if err != nil {
		t.Fatalf("RenderBrowser: %v", err)
	}
	if got != "" {
		t.Errorf("got %q, want empty output for an unknown type", got)
	}
}

func TestEffectivePaletteReturnsContextPaletteWhenSet(t *testing.T) {
	custom := themes.Palette{FG: "#111111", Muted: "#222222"}
	ctx := RenderContext{ColorMode: "dark", Palette: custom}

	if got := ctx.EffectivePalette(); got.Muted != "#222222" {
		t.Errorf("EffectivePalette().Muted = %q, want the context's %q", got.Muted, "#222222")
	}
}

func TestEffectivePaletteFallsBackPerColorMode(t *testing.T) {
	// The bug this guards: falling back to the light palette regardless of
	// ColorMode, which renders dark-mode output with light greys.
	dark := RenderContext{ColorMode: "dark"}.EffectivePalette()
	if want := themes.DefaultPalette("dark"); dark != want {
		t.Errorf("unset palette in dark mode = %+v, want %+v", dark, want)
	}
	if dark == themes.DefaultPalette("light") {
		t.Error("dark mode must not fall back to the light palette")
	}

	light := RenderContext{ColorMode: "light"}.EffectivePalette()
	if want := themes.DefaultPalette("light"); light != want {
		t.Errorf("unset palette in light mode = %+v, want %+v", light, want)
	}

	// An empty ColorMode is treated as light, matching DefaultPalette.
	if got := (RenderContext{}).EffectivePalette(); got != themes.DefaultPalette("light") {
		t.Errorf("empty color mode = %+v, want the light palette", got)
	}
}
