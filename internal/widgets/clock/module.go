package clock

import (
	"time"

	"github.com/andresbott/dashi/internal/widgets"
)

// Module implements widgets.Module, widgets.BrowserRenderable, and
// widgets.BrowserAssets for the clock widget. The image-stack renderer
// (for e-ink devices) lives in image.go; the browser-stack renderer and
// assets live in browser.go.
type Module struct {
	widgets.NoopModule
	now func() time.Time
}

// NewModule constructs a clock Module using the system clock.
func NewModule() *Module { return &Module{now: time.Now} }

// NewModuleWithClock constructs a clock Module with an injected time
// source, for tests.
func NewModuleWithClock(now func() time.Time) *Module { return &Module{now: now} }

// Type returns the widget type string.
func (m *Module) Type() string { return "clock" }

// Renderer returns the clock's image-stack renderer.
func (m *Module) Renderer() widgets.StaticRenderer {
	return NewStaticRenderer(m.now)
}
