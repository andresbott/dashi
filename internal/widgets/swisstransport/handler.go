package swisstransport

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	swisstransportpkg "github.com/andresbott/dashi/internal/swisstransport"
)

type handler struct {
	client *swisstransportpkg.Client
	logger *slog.Logger
}

func newHandler(client *swisstransportpkg.Client, logger *slog.Logger) *handler {
	return &handler{client: client, logger: logger}
}

func (h *handler) GetDepartures(w http.ResponseWriter, r *http.Request) {
	stationID := r.URL.Query().Get("id")
	if stationID == "" {
		writeJSONError(w, "id query parameter is required", http.StatusBadRequest)
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
		if h.logger != nil {
			h.logger.Error("fetch departures", slog.String("error", err.Error()))
		}
		writeJSONError(w, "failed to fetch departures", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(deps)
}

func (h *handler) SearchStations(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		writeJSONError(w, "query parameter is required", http.StatusBadRequest)
		return
	}

	stations, err := h.client.SearchStations(query)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("search stations", slog.String("error", err.Error()))
		}
		writeJSONError(w, "failed to search stations", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stations)
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
