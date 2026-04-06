package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/andresbott/dashi/internal/weather"
)

type WeatherHandler struct {
	client *weather.Client
	logger *slog.Logger
}

func NewWeatherHandler(client *weather.Client, logger *slog.Logger) *WeatherHandler {
	return &WeatherHandler{client: client, logger: logger}
}

func (h *WeatherHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	if latStr == "" || lonStr == "" {
		ErrorJSON(w, "lat and lon query parameters are required", http.StatusBadRequest)
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		ErrorJSON(w, "invalid lat parameter", http.StatusBadRequest)
		return
	}
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		ErrorJSON(w, "invalid lon parameter", http.StatusBadRequest)
		return
	}

	data, err := h.client.GetWeather(lat, lon)
	if err != nil {
		h.logger.Error("fetch weather data", slog.String("error", err.Error()))
		ErrorJSON(w, "failed to fetch weather data", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		ErrorJSON(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *WeatherHandler) Geocode(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		ErrorJSON(w, "city query parameter is required", http.StatusBadRequest)
		return
	}

	locations, err := h.client.Geocode(city)
	if err != nil {
		h.logger.Error("geocode city", slog.String("error", err.Error()))
		ErrorJSON(w, "failed to geocode city", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(locations); err != nil {
		ErrorJSON(w, "failed to encode response", http.StatusInternalServerError)
	}
}
