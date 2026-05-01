package clock

import (
	"github.com/andresbott/dashi/internal/widgets"
)

// Module implements widgets.Module for the clock widget.
type Module struct {
	widgets.NoopModule
}

// NewModule constructs a clock Module.
func NewModule() *Module { return &Module{} }

// Type returns the widget type string.
func (m *Module) Type() string { return "clock" }

// Renderer returns the clock's static renderer.
func (m *Module) Renderer() widgets.StaticRenderer {
	return NewStaticRenderer(nil)
}
