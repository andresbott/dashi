package widgets

import (
	"encoding/json"
	"html/template"
	"testing"
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
	want := template.HTML(`<div class="widget-placeholder"></div>`)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
