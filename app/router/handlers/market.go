package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/andresbott/dashi/internal/market"
)

type MarketHandler struct {
	client *market.Client
	logger *slog.Logger
}

func NewMarketHandler(client *market.Client, logger *slog.Logger) *MarketHandler {
	return &MarketHandler{client: client, logger: logger}
}

func (h *MarketHandler) GetMarketData(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		ErrorJSON(w, "symbol query parameter is required", http.StatusBadRequest)
		return
	}

	rangeID := r.URL.Query().Get("range")
	if rangeID == "" {
		rangeID = "1mo"
	}

	data, err := h.client.GetMarketData(symbol, rangeID)
	if err != nil {
		h.logger.Error("fetch market data", slog.String("error", err.Error()))
		ErrorJSON(w, "failed to fetch market data", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}
