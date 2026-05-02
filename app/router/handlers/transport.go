package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/andresbott/dashi/internal/swisstransport"
)

type TransportHandler struct {
	client *swisstransport.Client
	logger *slog.Logger
}

func NewTransportHandler(client *swisstransport.Client, logger *slog.Logger) *TransportHandler {
	return &TransportHandler{client: client, logger: logger}
}

func (h *TransportHandler) GetDepartures(w http.ResponseWriter, r *http.Request) {
	stationID := r.URL.Query().Get("id")
	if stationID == "" {
		ErrorJSON(w, "id query parameter is required", http.StatusBadRequest)
		return
	}

	limit := 5
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	deps, err := h.client.GetDepartures(stationID, limit)
	if err != nil {
		h.logger.Error("fetch departures", slog.String("error", err.Error()))
		ErrorJSON(w, "failed to fetch departures", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(deps)
}

func (h *TransportHandler) SearchStations(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		ErrorJSON(w, "query parameter is required", http.StatusBadRequest)
		return
	}

	stations, err := h.client.SearchStations(query)
	if err != nil {
		h.logger.Error("search stations", slog.String("error", err.Error()))
		ErrorJSON(w, "failed to search stations", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stations)
}
