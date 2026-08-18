package swisstransport

import (
	"log/slog"
	"net/http"

	swisstransportpkg "github.com/andresbott/dashi/internal/providers/swisstransport"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

// Module implements widgets.Module for the transport widget.
type Module struct {
	widgets.NoopModule
	client *swisstransportpkg.Client
	logger *slog.Logger
}

// NewModule constructs a transport Module.
func NewModule(client *swisstransportpkg.Client, logger *slog.Logger) *Module {
	return &Module{client: client, logger: logger}
}

// Type returns the widget type string. Note: the widget type is "transport"
// even though the package is "swisstransport".
func (m *Module) Type() string { return "transport" }

// Renderer returns the transport static renderer.
func (m *Module) Renderer() widgets.StaticRenderer {
	return NewStaticRenderer(m.client)
}

// RegisterRoutes mounts the transport interactive endpoints.
func (m *Module) RegisterRoutes(r *mux.Router) {
	h := newHandler(m.client, m.logger)
	r.Path("/widgets/transport/stationboard").Methods(http.MethodGet).HandlerFunc(h.GetDepartures)
	r.Path("/widgets/transport/stations").Methods(http.MethodGet).HandlerFunc(h.SearchStations)
}
