package widgets

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"strings"
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

func TestRegistry_ValidateConfig(t *testing.T) {
	errBad := errors.New("bad config")
	r := NewRegistry()
	r.RegisterValidator("strict", func(config json.RawMessage) error {
		if string(config) == `"bad"` {
			return errBad
		}
		return nil
	})

	tests := []struct {
		name       string
		widgetType string
		config     string
		want       error
	}{
		{name: "passes its validator", widgetType: "strict", config: `"ok"`},
		{name: "fails its validator", widgetType: "strict", config: `"bad"`, want: errBad},
		{name: "type without validator accepts anything", widgetType: "lenient", config: `"bad"`},
		{name: "unknown type accepts anything", widgetType: "", config: `"bad"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := r.ValidateConfig(tc.widgetType, json.RawMessage(tc.config)); !errors.Is(got, tc.want) {
				t.Errorf("ValidateConfig() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestEffectivePaletteReturnsContextPaletteWhenSet(t *testing.T) {
	custom := themes.Palette{Text: "#111111", TextMuted: "#222222"}
	ctx := RenderContext{ColorMode: "dark", Palette: custom}

	if got := ctx.EffectivePalette(); got.TextMuted != "#222222" {
		t.Errorf("EffectivePalette().TextMuted = %q, want the context's %q", got.TextMuted, "#222222")
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

func TestTypesInRecursesThroughContainers(t *testing.T) {
	r := NewRegistry()
	// A toy container whose config is the JSON list of its children's types;
	// a child named "box" is itself a container holding one "leaf".
	r.RegisterChildren("box", func(config json.RawMessage) []Child {
		var types []string
		if json.Unmarshal(config, &types) != nil {
			return nil
		}
		children := make([]Child, 0, len(types))
		for _, t := range types {
			cfg := json.RawMessage(`null`)
			if t == "box" {
				cfg = json.RawMessage(`["leaf"]`)
			}
			children = append(children, Child{Type: t, Config: cfg})
		}
		return children
	})

	tests := []struct {
		name   string
		typ    string
		config string
		want   []string
	}{
		{name: "plain widget", typ: "clock", config: `{}`, want: []string{"clock"}},
		{name: "container", typ: "box", config: `["clock","market"]`, want: []string{"box", "clock", "market"}},
		{name: "nested", typ: "box", config: `["box"]`, want: []string{"box", "box", "leaf"}},
		{name: "undecodable", typ: "box", config: `5`, want: []string{"box"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := r.TypesIn(tc.typ, json.RawMessage(tc.config))
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Errorf("TypesIn = %v, want %v", got, tc.want)
			}
		})
	}
}
