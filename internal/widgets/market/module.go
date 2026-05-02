package market

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	marketpkg "github.com/andresbott/dashi/internal/market"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"
)

// Module implements widgets.Module for the market widget.
type Module struct {
	client *marketpkg.Client
	logger *slog.Logger
}

// NewModule constructs a market Module.
func NewModule(client *marketpkg.Client, logger *slog.Logger) *Module {
	return &Module{client: client, logger: logger}
}

// Type returns the widget type string.
func (m *Module) Type() string { return "market" }

// Renderer returns the market static renderer.
func (m *Module) Renderer() widgets.StaticRenderer {
	return NewStaticRenderer(m.client)
}

// RegisterRoutes mounts the market interactive endpoint.
func (m *Module) RegisterRoutes(r *mux.Router) {
	h := newHandler(m.client, m.logger)
	r.Path("/widgets/market").Methods(http.MethodGet).HandlerFunc(h.GetMarketData)
}

// Warmup pre-fetches market data for every (symbol, range) pair configured
// across dashboards. Errors are logged and swallowed (best-effort at boot).
func (m *Module) Warmup(ctx context.Context, configs []json.RawMessage) {
	targets := extractTargets(configs)
	if len(targets) == 0 {
		return
	}
	if m.logger != nil {
		m.logger.Info("market warmup: pre-fetching data", slog.Int("symbols", len(targets)))
	}
	m.client.WarmupSymbols(ctx, targets)
	if m.logger != nil {
		m.logger.Info("market warmup: done")
	}
}

// extractTargets parses symbol/range out of each config, defaulting
// range to "1mo" when missing. Skips entries with empty symbol or
// malformed JSON.
func extractTargets(configs []json.RawMessage) []struct{ Symbol, Range string } {
	var out []struct{ Symbol, Range string }
	for _, raw := range configs {
		var cfg struct {
			Symbol string `json:"symbol"`
			Range  string `json:"range"`
		}
		if err := json.Unmarshal(raw, &cfg); err != nil || cfg.Symbol == "" {
			continue
		}
		if cfg.Range == "" {
			cfg.Range = "1mo"
		}
		out = append(out, struct{ Symbol, Range string }{cfg.Symbol, cfg.Range})
	}
	return out
}
