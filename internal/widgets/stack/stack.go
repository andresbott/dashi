package stack

import (
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

// NewStaticRenderer returns a StaticRenderer that renders child widgets vertically.
func NewStaticRenderer(registry *widgets.Registry) func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
	return func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
		var cfg stackConfig
		if len(config) > 0 {
			if err := json.Unmarshal(config, &cfg); err != nil {
				return "", fmt.Errorf("stack config: %w", err)
			}
		}

		if len(cfg.Widgets) == 0 {
			return template.HTML(`<div class="widget-stack"></div>`), nil
		}

		var b strings.Builder
		b.WriteString(`<div class="widget-stack">`)
		for _, child := range cfg.Widgets {
			rendered, err := registry.Render(child.Type, child.Config, ctx)
			if err != nil {
				return "", fmt.Errorf("stack child %s (%s): %w", child.ID, child.Type, err)
			}
			b.WriteString(string(rendered))
		}
		b.WriteString(`</div>`)

		return template.HTML(b.String()), nil
	}
}
