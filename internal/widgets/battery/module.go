package battery

import (
	"github.com/andresbott/dashi/internal/widgets"
)

// Module implements widgets.Module for the battery widget.
type Module struct {
	widgets.NoopModule
}

// NewModule constructs a battery Module.
func NewModule() *Module { return &Module{} }

// Type returns the widget type string.
func (m *Module) Type() string { return "battery" }

// Renderer returns the battery's static renderer.
func (m *Module) Renderer() widgets.StaticRenderer {
	return NewStaticRenderer()
}
