package weather

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	weatherpkg "github.com/andresbott/dashi/internal/providers/weather"
)

type handler struct {
	client *weatherpkg.Client
	logger *slog.Logger
}

func newHandler(client *weatherpkg.Client, logger *slog.Logger) *handler {
	return &handler{client: client, logger: logger}
}

func (h *handler) GetWeather(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	if latStr == "" || lonStr == "" {
		writeJSONError(w, "lat and lon query parameters are required", http.StatusBadRequest)
		return
	}
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		writeJSONError(w, "invalid lat parameter", http.StatusBadRequest)
		return
	}
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		writeJSONError(w, "invalid lon parameter", http.StatusBadRequest)
		return
	}

	data, err := h.client.GetWeather(lat, lon)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("fetch weather data", slog.String("error", err.Error()))
		}
		writeJSONError(w, "failed to fetch weather data", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		writeJSONError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *handler) Geocode(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		writeJSONError(w, "city query parameter is required", http.StatusBadRequest)
		return
	}
	locations, err := h.client.Geocode(city)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("geocode city", slog.String("error", err.Error()))
		}
		writeJSONError(w, "failed to geocode city", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(locations); err != nil {
		writeJSONError(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
