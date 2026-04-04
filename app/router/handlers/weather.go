package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/andresbott/dashi/internal/weather"
)

type WeatherHandler struct {
	client *weather.Client
}

func NewWeatherHandler(client *weather.Client) *WeatherHandler {
	return &WeatherHandler{client: client}
}

func (h *WeatherHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	if latStr == "" || lonStr == "" {
		http.Error(w, `{"error":"lat and lon query parameters are required"}`, http.StatusBadRequest)
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid lat parameter"}`, http.StatusBadRequest)
		return
	}
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid lon parameter"}`, http.StatusBadRequest)
		return
	}

	data, err := h.client.GetWeather(lat, lon)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch weather data"}`, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *WeatherHandler) Geocode(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, `{"error":"city query parameter is required"}`, http.StatusBadRequest)
		return
	}

	locations, err := h.client.Geocode(city)
	if err != nil {
		http.Error(w, `{"error":"failed to geocode city"}`, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locations)
}
