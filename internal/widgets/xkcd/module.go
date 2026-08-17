package xkcd

import (
	"log/slog"
	"net/http"

	xkcdclient "github.com/andresbott/dashi/internal/providers/xkcd"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

// Module implements widgets.Module for the xkcd widget.
type Module struct {
	widgets.NoopModule // inherits no-op Warmup
	client             *xkcdclient.Client
	logger             *slog.Logger
}

// NewModule constructs an xkcd Module.
func NewModule(client *xkcdclient.Client, logger *slog.Logger) *Module {
	return &Module{client: client, logger: logger}
}

// Type returns the widget type string.
func (m *Module) Type() string { return "xkcd" }

// Renderer returns the xkcd static renderer.
func (m *Module) Renderer() widgets.StaticRenderer {
	return NewStaticRenderer(m.client)
}

// RegisterRoutes mounts the xkcd interactive endpoint.
func (m *Module) RegisterRoutes(r *mux.Router) {
	h := newHandler(m.client, m.logger)
	r.Path("/widgets/xkcd").Methods(http.MethodGet).HandlerFunc(h.GetComic)
}
