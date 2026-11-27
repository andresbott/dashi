package stack

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"github.com/andresbott/dashi/internal/widgets"
)

type childWidget struct {
	ID     string          `json:"id"`
	Type   string          `json:"type"`
	Title  string          `json:"title"`
	Config json.RawMessage `json:"config,omitempty"`
}

type stackConfig struct {
	Widgets []childWidget `json:"widgets"`
}

//go:embed browser.html
var browserHTML string

var browserTmpl = template.Must(template.New("stack-browser").Parse(strings.TrimSpace(browserHTML)))

// NewStaticRenderer returns a StaticRenderer that renders child widgets vertically.
func NewStaticRenderer(registry *widgets.Registry) func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
	return func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
		children, err := renderChildren(config, ctx, registry.Render)
		if err != nil {
			return "", err
		}
		return `<div class="widget-stack">` + children + `</div>`, nil
	}
}

// RenderBrowser implements widgets.BrowserRenderable: children render
// through their browser renderers (falling back to their image renderer
// when unported), so a ported widget looks the same inside a stack as on
// its own.
func (m *Module) RenderBrowser(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
	children, err := renderChildren(config, ctx, m.registry.RenderBrowser)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if err := browserTmpl.Execute(&b, children); err != nil {
		return "", fmt.Errorf("stack browser render: %w", err)
	}
	return template.HTML(b.String()), nil
}

// Children implements widgets.Container: the stack's direct children, so
// the browser page loads the scripts of widgets that only appear nested.
func (m *Module) Children(config json.RawMessage) []widgets.Child {
	var cfg stackConfig
	if len(config) == 0 || json.Unmarshal(config, &cfg) != nil {
		return nil
	}
	children := make([]widgets.Child, 0, len(cfg.Widgets))
	for _, child := range cfg.Widgets {
		children = append(children, widgets.Child{Type: child.Type, Config: child.Config})
	}
	return children
}

// renderChildren renders the stack's children, in order, with renderChild.
func renderChildren(config json.RawMessage, ctx widgets.RenderContext, renderChild widgets.TypedRenderer) (template.HTML, error) {
	var cfg stackConfig
	if len(config) > 0 {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return "", fmt.Errorf("stack config: %w", err)
		}
	}

	var b strings.Builder
	for _, child := range cfg.Widgets {
		rendered, err := renderChild(child.Type, child.Config, ctx)
		if err != nil {
			return "", fmt.Errorf("stack child %s (%s): %w", child.ID, child.Type, err)
		}
		b.WriteString(string(rendered))
	}

	return template.HTML(b.String()), nil
}
