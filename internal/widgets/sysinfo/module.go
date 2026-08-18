package sysinfo

import (
	"log/slog"
	"net/http"

	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

// Module implements widgets.Module for the sysinfo widget.
type Module struct {
	widgets.NoopModule
	logger *slog.Logger
}

// NewModule constructs a sysinfo Module.
func NewModule(logger *slog.Logger) *Module {
	return &Module{logger: logger}
}

// Type returns the widget type string.
func (m *Module) Type() string { return "sysinfo" }

// Renderer returns the sysinfo static renderer.
func (m *Module) Renderer() widgets.StaticRenderer {
	return NewStaticRenderer()
}

// RegisterRoutes mounts the sysinfo interactive endpoint.
func (m *Module) RegisterRoutes(r *mux.Router) {
	h := newHandler(m.logger)
	r.Path("/widgets/sysinfo").Methods(http.MethodGet).HandlerFunc(h.GetSysinfo)
}
