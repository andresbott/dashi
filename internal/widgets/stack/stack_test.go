package stack

import (
	"encoding/json"
	"errors"
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
	ctx := widgets.RenderContext{WidgetTheme: "dark", PageIndex: 2, TotalPages: 5}
	_, err := renderer(config, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedCtx.WidgetTheme != "dark" || receivedCtx.PageIndex != 2 || receivedCtx.TotalPages != 5 {
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
	if strings.Contains(html, "widget-placeholder") {
		t.Errorf("unknown child type should render nothing, not a placeholder, got: %s", html)
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

// In the browser stack, children render through their browser renderers
// (so ported widgets keep their browser look inside a stack), unported
// children fall back to their image renderer, and the wrapper is the
// spaced Tailwind column.
func TestRenderBrowser_UsesChildBrowserRenderers(t *testing.T) {
	reg := widgets.NewRegistry()
	reg.Register("ported", func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
		return "<p>ported-image</p>", nil
	})
	reg.RegisterBrowser("ported", func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
		return "<p>ported-browser</p>", nil
	})
	reg.Register("unported", func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
		return "<p>unported-image</p>", nil
	})

	m := NewModule(reg)
	config := json.RawMessage(`{"widgets":[{"id":"a","type":"ported"},{"id":"b","type":"unported"}]}`)
	got, err := m.RenderBrowser(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("RenderBrowser: %v", err)
	}
	want := template.HTML(`<div class="dashi-stack flex flex-col gap-5"><p>ported-browser</p><p>unported-image</p></div>`)
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// The image renderer is untouched by the browser port.
	img, err := NewStaticRenderer(reg)(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("image render: %v", err)
	}
	if img != `<div class="widget-stack"><p>ported-image</p><p>unported-image</p></div>` {
		t.Errorf("image output changed: %q", img)
	}
}

func TestRenderBrowser_ChildErrorAndBadConfig(t *testing.T) {
	reg := widgets.NewRegistry()
	reg.Register("boom", func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
		return "", errors.New("kaput")
	})
	m := NewModule(reg)

	if _, err := m.RenderBrowser(json.RawMessage(`{"widgets":[{"id":"x","type":"boom"}]}`), widgets.RenderContext{}); err == nil || !strings.Contains(err.Error(), "stack child x (boom): kaput") {
		t.Errorf("child error = %v", err)
	}
	if _, err := m.RenderBrowser(json.RawMessage(`{"widgets":5}`), widgets.RenderContext{}); err == nil {
		t.Error("expected a config error")
	}
}

func TestChildren(t *testing.T) {
	m := NewModule(widgets.NewRegistry())
	got := m.Children(json.RawMessage(`{"widgets":[{"id":"a","type":"market","config":{"symbol":"X"}},{"id":"b","type":"bookmark"}]}`))
	if len(got) != 2 || got[0].Type != "market" || string(got[0].Config) != `{"symbol":"X"}` || got[1].Type != "bookmark" {
		t.Errorf("Children = %+v", got)
	}
	if got := m.Children(json.RawMessage(`nope`)); got != nil {
		t.Errorf("undecodable config: Children = %+v, want nil", got)
	}
	if got := m.Children(nil); got != nil {
		t.Errorf("empty config: Children = %+v, want nil", got)
	}
}
