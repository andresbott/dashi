package weather

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	weatherpkg "github.com/andresbott/dashi/internal/weather"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

// Module implements widgets.Module for the full weather widget.
type Module struct {
	client     *weatherpkg.Client
	themeStore *themes.Store
	logger     *slog.Logger
}

// NewModule constructs the weather Module.
func NewModule(client *weatherpkg.Client, themeStore *themes.Store, logger *slog.Logger) *Module {
	return &Module{client: client, themeStore: themeStore, logger: logger}
}

func (m *Module) Type() string                 { return "weather" }
func (m *Module) Renderer() widgets.StaticRenderer { return NewStaticRenderer(m.client, m.themeStore) }

// RegisterRoutes mounts /widgets/weather and /widgets/weather/geocode.
// weather-compact (see CompactModule) intentionally does not mount
// these — both widgets share the same HTTP endpoints.
func (m *Module) RegisterRoutes(r *mux.Router) {
	h := newHandler(m.client, m.logger)
	r.Path("/widgets/weather").Methods(http.MethodGet).HandlerFunc(h.GetWeather)
	r.Path("/widgets/weather/geocode").Methods(http.MethodGet).HandlerFunc(h.Geocode)
}

// Warmup pre-fetches weather data for every configured lat/lon pair.
// Errors are logged and swallowed (warmup is best-effort at boot).
func (m *Module) Warmup(ctx context.Context, configs []json.RawMessage) {
	locations := extractLocations(configs)
	if len(locations) == 0 {
		return
	}
	if m.logger != nil {
		m.logger.Info("weather warmup: pre-fetching data", slog.Int("locations", len(locations)))
	}
	m.client.WarmupLocations(ctx, locations)
	if m.logger != nil {
		m.logger.Info("weather warmup: done")
	}
}

// CompactModule implements widgets.Module for weather-compact. It shares
// the weather client but has a different type string and a different
// renderer; it inherits no-op RegisterRoutes (to avoid double-mounting)
// and implements its own Warmup so compact-only dashboards still get
// data pre-fetched.
type CompactModule struct {
	widgets.NoopModule // inherits no-op RegisterRoutes
	client             *weatherpkg.Client
	themeStore         *themes.Store
	logger             *slog.Logger
}

// NewCompactModule constructs the weather-compact Module.
func NewCompactModule(client *weatherpkg.Client, themeStore *themes.Store, logger *slog.Logger) *CompactModule {
	return &CompactModule{client: client, themeStore: themeStore, logger: logger}
}

func (m *CompactModule) Type() string                 { return "weather-compact" }
func (m *CompactModule) Renderer() widgets.StaticRenderer { return NewStaticCompactRenderer(m.client, m.themeStore) }

func (m *CompactModule) Warmup(ctx context.Context, configs []json.RawMessage) {
	locations := extractLocations(configs)
	if len(locations) == 0 {
		return
	}
	if m.logger != nil {
		m.logger.Info("weather-compact warmup: pre-fetching data", slog.Int("locations", len(locations)))
	}
	m.client.WarmupLocations(ctx, locations)
	if m.logger != nil {
		m.logger.Info("weather-compact warmup: done")
	}
}

// extractLocations parses the lat/lon fields out of each config,
// skipping malformed entries and (0, 0) pairs (treated as unset).
func extractLocations(configs []json.RawMessage) [][2]float64 {
	var out [][2]float64
	for _, raw := range configs {
		var cfg struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		}
		if err := json.Unmarshal(raw, &cfg); err != nil {
			continue
		}
		if cfg.Latitude == 0 && cfg.Longitude == 0 {
			continue
		}
		out = append(out, [2]float64{cfg.Latitude, cfg.Longitude})
	}
	return out
}
