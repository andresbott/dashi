package stack

import (
	"github.com/andresbott/dashi/internal/widgets"
)

// Module implements widgets.Module for the stack widget. Stack renders
// child widgets recursively, so it needs a reference to the registry.
type Module struct {
	widgets.NoopModule
	registry *widgets.Registry
}

// NewModule constructs a stack Module. The registry is captured by
// reference; it may be populated after construction because the
// renderer is only invoked at request time.
func NewModule(registry *widgets.Registry) *Module {
	return &Module{registry: registry}
}

// Type returns the widget type string.
func (m *Module) Type() string { return "stack" }

// Renderer returns the stack's static renderer.
func (m *Module) Renderer() widgets.StaticRenderer {
	return NewStaticRenderer(m.registry)
}
