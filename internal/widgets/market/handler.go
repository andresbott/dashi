package market

import (
	"encoding/json"
	"log/slog"
	"net/http"

	marketpkg "github.com/andresbott/dashi/internal/providers/market"
)

type handler struct {
	client *marketpkg.Client
	logger *slog.Logger
}

func newHandler(client *marketpkg.Client, logger *slog.Logger) *handler {
	return &handler{client: client, logger: logger}
}

func (h *handler) GetMarketData(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		writeJSONError(w, "symbol query parameter is required", http.StatusBadRequest)
		return
	}

	rangeID := r.URL.Query().Get("range")
	if rangeID == "" {
		rangeID = "1mo"
	}

	data, err := h.client.GetMarketData(symbol, rangeID)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("fetch market data", slog.String("error", err.Error()))
		}
		writeJSONError(w, "failed to fetch market data", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
