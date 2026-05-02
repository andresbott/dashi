package markdown

import (
	"log/slog"
	"net/http"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

// Module implements widgets.Module for the markdown widget.
type Module struct {
	widgets.NoopModule
	store  *dashboard.Store
	logger *slog.Logger
}

// NewModule constructs a markdown Module. The dashboard store is
// required for reading markdown files from dashboard asset folders.
func NewModule(store *dashboard.Store, logger *slog.Logger) *Module {
	return &Module{store: store, logger: logger}
}

// Type returns the widget type string.
func (m *Module) Type() string { return "markdown" }

// Renderer returns the markdown static renderer.
func (m *Module) Renderer() widgets.StaticRenderer {
	return NewStaticRenderer(m.store)
}

// RegisterRoutes mounts the markdown dashboard-scoped endpoints.
func (m *Module) RegisterRoutes(r *mux.Router) {
	h := newHandler(m.store, m.logger)
	r.Path("/dashboards/{id}/markdown").Methods(http.MethodGet).HandlerFunc(h.ListMarkdown)
	r.Path("/dashboards/{id}/markdown/{filename}").Methods(http.MethodGet).HandlerFunc(h.GetMarkdown)
}
