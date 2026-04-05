package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/andresbott/dashi/internal/market"
)

type MarketHandler struct {
	client *market.Client
}

func NewMarketHandler(client *market.Client) *MarketHandler {
	return &MarketHandler{client: client}
}

func (h *MarketHandler) GetMarketData(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, `{"error":"symbol query parameter is required"}`, http.StatusBadRequest)
		return
	}

	rangeID := r.URL.Query().Get("range")
	if rangeID == "" {
		rangeID = "1mo"
	}

	data, err := h.client.GetMarketData(symbol, rangeID)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch market data"}`, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
