package markdown

import (
	"github.com/andresbott/dashi/internal/data/notes"
	"github.com/andresbott/dashi/internal/widgets"
)

// Module implements widgets.Module for the markdown widget.
type Module struct {
	widgets.NoopModule
	store *notes.Store
}

// NewModule constructs a markdown Module backed by the shared notes
// data store.
func NewModule(store *notes.Store) *Module {
	return &Module{store: store}
}

// Type returns the widget type string.
func (m *Module) Type() string { return "markdown" }

// Renderer returns the markdown static renderer.
func (m *Module) Renderer() widgets.StaticRenderer {
	return NewStaticRenderer(m.store)
}
