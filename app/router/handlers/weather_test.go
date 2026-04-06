package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	weatherpkg "github.com/andresbott/dashi/internal/weather"
)

func setupWeatherTestServer() (*httptest.Server, *httptest.Server) {
	forecastSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"current": map[string]any{
				"temperature_2m":        18.5,
				"relative_humidity_2m":  65,
				"apparent_temperature":  16.2,
				"weather_code":          2,
				"wind_speed_10m":        12.3,
			},
			"daily": map[string]any{
				"time":               []string{"2026-04-03"},
				"weather_code":       []int{2},
				"temperature_2m_max": []float64{20.0},
				"temperature_2m_min": []float64{10.0},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))

	geoSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"results": []map[string]any{
				{"name": "Berlin", "country": "Germany", "latitude": 52.52, "longitude": 13.405},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))

	return forecastSrv, geoSrv
}

func TestWeatherHandler_GetWeather(t *testing.T) {
	forecastSrv, geoSrv := setupWeatherTestServer()
	defer forecastSrv.Close()
	defer geoSrv.Close()

	client := weatherpkg.NewClient(
		&http.Client{},
		weatherpkg.WithBaseURL(forecastSrv.URL),
		weatherpkg.WithGeoBaseURL(geoSrv.URL),
	)
	h := NewWeatherHandler(client, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/api/v0/widgets/weather?lat=52.52&lon=13.41", nil)
	rec := httptest.NewRecorder()

	h.GetWeather(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var data weatherpkg.WeatherData
	if err := json.NewDecoder(rec.Body).Decode(&data); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if data.Current.Temperature != 18.5 {
		t.Errorf("temperature = %f, want 18.5", data.Current.Temperature)
	}
}

func TestWeatherHandler_GetWeather_MissingParams(t *testing.T) {
	client := weatherpkg.NewClient(&http.Client{})
	h := NewWeatherHandler(client, slog.Default())

	tests := []struct {
		name string
		url  string
	}{
		{"missing lat", "/api/v0/widgets/weather?lon=13.41"},
		{"missing lon", "/api/v0/widgets/weather?lat=52.52"},
		{"missing both", "/api/v0/widgets/weather"},
		{"invalid lat", "/api/v0/widgets/weather?lat=abc&lon=13.41"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()
			h.GetWeather(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rec.Code)
			}
		})
	}
}

func TestWeatherHandler_Geocode(t *testing.T) {
	forecastSrv, geoSrv := setupWeatherTestServer()
	defer forecastSrv.Close()
	defer geoSrv.Close()

	client := weatherpkg.NewClient(
		&http.Client{},
		weatherpkg.WithBaseURL(forecastSrv.URL),
		weatherpkg.WithGeoBaseURL(geoSrv.URL),
	)
	h := NewWeatherHandler(client, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/api/v0/widgets/weather/geocode?city=Berlin", nil)
	rec := httptest.NewRecorder()

	h.Geocode(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var locs []weatherpkg.Location
	if err := json.NewDecoder(rec.Body).Decode(&locs); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(locs) != 1 {
		t.Fatalf("expected 1 location, got %d", len(locs))
	}
	if locs[0].Name != "Berlin" {
		t.Errorf("name = %q, want Berlin", locs[0].Name)
	}
}

func TestWeatherHandler_Geocode_MissingCity(t *testing.T) {
	client := weatherpkg.NewClient(&http.Client{})
	h := NewWeatherHandler(client, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/api/v0/widgets/weather/geocode", nil)
	rec := httptest.NewRecorder()
	h.Geocode(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}
