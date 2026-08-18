package search

import (
	"encoding/json"
	"html/template"

	"github.com/andresbott/dashi/internal/widgets"
)

// Module implements widgets.Module for the search widget. The search
// widget is frontend-only (interactive mode); in image mode it renders
// a placeholder div.
type Module struct {
	widgets.NoopModule
}

// NewModule constructs a search Module.
func NewModule() *Module { return &Module{} }

// Type returns the widget type string.
func (m *Module) Type() string { return "search" }

// Renderer returns a placeholder renderer for image-mode dashboards.
func (m *Module) Renderer() widgets.StaticRenderer {
	return func(_ json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
		return template.HTML(`<div class="widget-search-placeholder">&nbsp;</div>`), nil
	}
}
