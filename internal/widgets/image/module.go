package image

import (
	"github.com/andresbott/dashi/internal/data/images"
	"github.com/andresbott/dashi/internal/widgets"
)

// Module implements widgets.Module for the image widget.
type Module struct {
	widgets.NoopModule
	store *images.Store
}

// NewModule constructs an image Module backed by the shared images
// data store.
func NewModule(store *images.Store) *Module {
	return &Module{store: store}
}

// Type returns the widget type string.
func (m *Module) Type() string { return "image" }

// Renderer returns the image's static renderer.
func (m *Module) Renderer() widgets.StaticRenderer {
	return NewStaticRenderer(m.store)
}
