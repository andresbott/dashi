package bookmark

import (
	"github.com/andresbott/dashi/internal/widgets"
)

// Module implements widgets.Module for the bookmark widget.
type Module struct {
	widgets.NoopModule
}

// NewModule constructs a bookmark Module.
func NewModule() *Module { return &Module{} }

// Type returns the widget type string.
func (m *Module) Type() string { return "bookmark" }

// Renderer returns the bookmark's static renderer.
func (m *Module) Renderer() widgets.StaticRenderer {
	return NewStaticRenderer()
}
