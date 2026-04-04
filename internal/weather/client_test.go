package weather

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetWeather(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/forecast" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		lat := r.URL.Query().Get("latitude")
		lon := r.URL.Query().Get("longitude")
		if lat != "52.520000" || lon != "13.410000" {
			t.Fatalf("unexpected coords: lat=%s lon=%s", lat, lon)
		}

		// Verify hourly params are requested
		hourly := r.URL.Query().Get("hourly")
		if hourly == "" {
			t.Error("expected hourly query parameter")
		}

		resp := openMeteoForecastResponse{
			Current: openMeteoCurrent{
				Temperature:  18.5,
				Humidity:     65,
				ApparentTemp: 16.2,
				WeatherCode:  2,
				WindSpeed:    12.3,
			},
			Hourly: openMeteoHourly{
				Time:        []string{"2026-04-04T12:00", "2026-04-04T13:00", "2026-04-04T14:00"},
				Temperature: []float64{18.0, 19.0, 20.0},
				WeatherCode: []int{2, 3, 61},
			},
			Daily: openMeteoDaily{
				Time:        []string{"2026-04-03", "2026-04-04"},
				WeatherCode: []int{2, 61},
				TempMax:     []float64{20.0, 15.0},
				TempMin:     []float64{10.0, 8.0},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient(&http.Client{}, WithBaseURL(srv.URL))
	data, err := c.GetWeather(52.52, 13.41)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Current.Temperature != 18.5 {
		t.Errorf("temperature = %f, want 18.5", data.Current.Temperature)
	}
	if data.Current.Description != "Partly cloudy" {
		t.Errorf("description = %q, want %q", data.Current.Description, "Partly cloudy")
	}
	if len(data.Forecast) != 2 {
		t.Fatalf("forecast len = %d, want 2", len(data.Forecast))
	}
	if data.Forecast[1].Description != "Slight rain" {
		t.Errorf("forecast[1] description = %q, want %q", data.Forecast[1].Description, "Slight rain")
	}
}

func TestClient_GetWeather_Cached(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := openMeteoForecastResponse{
			Current: openMeteoCurrent{Temperature: 20.0, WeatherCode: 0},
			Daily:   openMeteoDaily{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient(&http.Client{}, WithBaseURL(srv.URL))

	_, _ = c.GetWeather(52.52, 13.41)
	_, _ = c.GetWeather(52.52, 13.41)

	if callCount != 1 {
		t.Errorf("expected 1 API call (cached), got %d", callCount)
	}
}

func TestClient_Geocode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		name := r.URL.Query().Get("name")
		if name != "Berlin" {
			t.Fatalf("unexpected name: %s", name)
		}

		resp := openMeteoGeoResponse{
			Results: []openMeteoGeoResult{
				{Name: "Berlin", Country: "Germany", Latitude: 52.52, Longitude: 13.405},
				{Name: "Berlin", Country: "United States", Latitude: 44.47, Longitude: -71.18},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient(&http.Client{}, WithGeoBaseURL(srv.URL))
	locs, err := c.Geocode("Berlin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locs) != 2 {
		t.Fatalf("expected 2 locations, got %d", len(locs))
	}
	if locs[0].Country != "Germany" {
		t.Errorf("first result country = %q, want Germany", locs[0].Country)
	}
}

func TestClient_Geocode_EmptyResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(openMeteoGeoResponse{})
	}))
	defer srv.Close()

	c := NewClient(&http.Client{}, WithGeoBaseURL(srv.URL))
	locs, err := c.Geocode("xyznonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locs) != 0 {
		t.Fatalf("expected 0 locations, got %d", len(locs))
	}
}
