package search

import (
	"github.com/andresbott/dashi/internal/widgets"
)

// Module implements widgets.Module and widgets.BrowserRenderable for the
// search widget. The engine allowlist and config defaults live in
// engine.go, the image-stack renderer in image.go and the browser-stack
// renderer in browser.go.
type Module struct {
	widgets.NoopModule
}

// NewModule constructs a search Module.
func NewModule() *Module { return &Module{} }

// Type returns the widget type string.
func (m *Module) Type() string { return "search" }

// Renderer returns the search widget's image-stack renderer.
func (m *Module) Renderer() widgets.StaticRenderer { return renderImage }
