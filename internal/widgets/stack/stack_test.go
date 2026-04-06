package stack

import (
	"encoding/json"
	"html/template"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
)

func TestRenderStatic_Empty(t *testing.T) {
	reg := widgets.NewRegistry()
	renderer := NewStaticRenderer(reg)
	got, err := renderer(json.RawMessage(`{}`), widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := template.HTML(`<div class="widget-stack"></div>`)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderStatic_EmptyWidgets(t *testing.T) {
	reg := widgets.NewRegistry()
	renderer := NewStaticRenderer(reg)
	got, err := renderer(json.RawMessage(`{"widgets":[]}`), widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := template.HTML(`<div class="widget-stack"></div>`)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderStatic_SingleChild(t *testing.T) {
	reg := widgets.NewRegistry()
	reg.Register("fake", func(_ json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
		return template.HTML("<p>hello</p>"), nil
	})

	config := json.RawMessage(`{"widgets":[{"id":"a","type":"fake","title":"F","config":{}}]}`)
	renderer := NewStaticRenderer(reg)
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	html := string(got)
	if !strings.Contains(html, "<p>hello</p>") {
		t.Errorf("expected child HTML, got: %s", html)
	}
	if !strings.Contains(html, `class="widget-stack"`) {
		t.Errorf("expected stack wrapper, got: %s", html)
	}
}

func TestRenderStatic_MultipleChildren(t *testing.T) {
	reg := widgets.NewRegistry()
	reg.Register("a", func(_ json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
		return template.HTML("<p>first</p>"), nil
	})
	reg.Register("b", func(_ json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
		return template.HTML("<p>second</p>"), nil
	})

	config := json.RawMessage(`{"widgets":[{"id":"1","type":"a","title":"A","config":{}},{"id":"2","type":"b","title":"B","config":{}}]}`)
	renderer := NewStaticRenderer(reg)
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	html := string(got)
	firstIdx := strings.Index(html, "<p>first</p>")
	secondIdx := strings.Index(html, "<p>second</p>")
	if firstIdx == -1 || secondIdx == -1 {
		t.Fatalf("expected both children in output, got: %s", html)
	}
	if firstIdx >= secondIdx {
		t.Errorf("expected first child before second, got: %s", html)
	}
}

func TestRenderStatic_PassesContext(t *testing.T) {
	reg := widgets.NewRegistry()
	var receivedCtx widgets.RenderContext
	reg.Register("ctx", func(_ json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
		receivedCtx = ctx
		return template.HTML("<p>ok</p>"), nil
	})

	config := json.RawMessage(`{"widgets":[{"id":"1","type":"ctx","title":"C","config":{}}]}`)
	renderer := NewStaticRenderer(reg)
	ctx := widgets.RenderContext{Theme: "dark", PageIndex: 2, TotalPages: 5}
	_, err := renderer(config, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedCtx.Theme != "dark" || receivedCtx.PageIndex != 2 || receivedCtx.TotalPages != 5 {
		t.Errorf("context not passed through: %+v", receivedCtx)
	}
}

func TestRenderStatic_UnknownChildType(t *testing.T) {
	reg := widgets.NewRegistry()
	config := json.RawMessage(`{"widgets":[{"id":"1","type":"nope","title":"N","config":{}}]}`)
	renderer := NewStaticRenderer(reg)
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	html := string(got)
	if !strings.Contains(html, "widget-placeholder") {
		t.Errorf("expected placeholder for unknown type, got: %s", html)
	}
}

func TestRenderStatic_InvalidConfig(t *testing.T) {
	reg := widgets.NewRegistry()
	renderer := NewStaticRenderer(reg)
	_, err := renderer(json.RawMessage(`{invalid`), widgets.RenderContext{})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
