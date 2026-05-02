package image

import (
	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/widgets"
)

// Module implements widgets.Module for the image widget.
type Module struct {
	widgets.NoopModule
	store *dashboard.Store
}

// NewModule constructs an image Module. The dashboard store is required
// for resolving uploaded image assets.
func NewModule(store *dashboard.Store) *Module {
	return &Module{store: store}
}

// Type returns the widget type string.
func (m *Module) Type() string { return "image" }

// Renderer returns the image's static renderer.
func (m *Module) Renderer() widgets.StaticRenderer {
	return NewStaticRenderer(m.store)
}
