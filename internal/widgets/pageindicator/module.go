package pageindicator

import (
	"github.com/andresbott/dashi/internal/widgets"
)

type Module struct {
	widgets.NoopModule
}

func NewModule() *Module { return &Module{} }

// Type returns the widget type string. Note: package is "pageindicator"
// but the type string uses a hyphen to match dashboard JSON conventions.
func (m *Module) Type() string { return "page-indicator" }

func (m *Module) Renderer() widgets.StaticRenderer {
	return NewStaticRenderer()
}
