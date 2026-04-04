package weather

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/themes"
	weatherpkg "github.com/andresbott/dashi/internal/weather"
)

func newTestWeatherServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"current": map[string]any{
				"temperature_2m":       18.5,
				"relative_humidity_2m": 65,
				"apparent_temperature": 16.2,
				"weather_code":         2,
				"wind_speed_10m":       12.3,
			},
			"hourly": map[string]any{
				"time":           []string{"2099-01-01T12:00", "2099-01-01T13:00", "2099-01-01T14:00"},
				"temperature_2m": []float64{18.0, 19.0, 20.0},
				"weather_code":   []int{2, 3, 61},
			},
			"daily": map[string]any{
				"time":               []string{"2026-04-03", "2026-04-04"},
				"weather_code":       []int{2, 61},
				"temperature_2m_max": []float64{20.0, 15.0},
				"temperature_2m_min": []float64{10.0, 8.0},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func TestRenderStatic_Basic(t *testing.T) {
	srv := newTestWeatherServer()
	defer srv.Close()

	client := weatherpkg.NewClient(&http.Client{}, weatherpkg.WithBaseURL(srv.URL))
	themeStore := themes.NewStore("")
	config := json.RawMessage(`{
		"city": "Basel",
		"latitude": 47.558,
		"longitude": 7.573,
		"compact": false,
		"showDetails": false,
		"showForecast": false
	}`)

	renderer := NewStaticRenderer(client, themeStore)
	got, err := renderer(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "Basel") {
		t.Errorf("expected city name in output, got: %s", html)
	}
	if !strings.Contains(html, "18.5") {
		t.Errorf("expected temperature in output, got: %s", html)
	}
	if !strings.Contains(html, "Partly cloudy") {
		t.Errorf("expected description in output, got: %s", html)
	}
}

func TestRenderStatic_WithDetails(t *testing.T) {
	srv := newTestWeatherServer()
	defer srv.Close()

	client := weatherpkg.NewClient(&http.Client{}, weatherpkg.WithBaseURL(srv.URL))
	themeStore := themes.NewStore("")
	config := json.RawMessage(`{
		"city": "Basel",
		"latitude": 47.558,
		"longitude": 7.573,
		"compact": false,
		"showDetails": true,
		"showForecast": false
	}`)

	renderer := NewStaticRenderer(client, themeStore)
	got, err := renderer(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "16.2") {
		t.Errorf("expected feels-like temp in output, got: %s", html)
	}
	if !strings.Contains(html, "65%") {
		t.Errorf("expected humidity in output, got: %s", html)
	}
	if !strings.Contains(html, "12.3") {
		t.Errorf("expected wind speed in output, got: %s", html)
	}
}

func TestRenderStatic_WithForecast(t *testing.T) {
	srv := newTestWeatherServer()
	defer srv.Close()

	client := weatherpkg.NewClient(&http.Client{}, weatherpkg.WithBaseURL(srv.URL))
	themeStore := themes.NewStore("")
	config := json.RawMessage(`{
		"city": "Basel",
		"latitude": 47.558,
		"longitude": 7.573,
		"compact": false,
		"showDetails": false,
		"showForecast": true
	}`)

	renderer := NewStaticRenderer(client, themeStore)
	got, err := renderer(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "forecast-box") {
		t.Error("expected forecast-box class")
	}
	if !strings.Contains(html, "Fri") {
		t.Errorf("expected day name in output, got: %s", html)
	}
	// Default theme resolves "partly-cloudy" → "ti ti-sun-low"
	if !strings.Contains(html, "ti ti-sun-low") {
		t.Errorf("expected resolved icon class in output, got: %s", html)
	}
}

func TestRenderStatic_ForecastOnly(t *testing.T) {
	srv := newTestWeatherServer()
	defer srv.Close()

	client := weatherpkg.NewClient(&http.Client{}, weatherpkg.WithBaseURL(srv.URL))
	themeStore := themes.NewStore("")
	config := json.RawMessage(`{
		"city": "Basel",
		"latitude": 47.558,
		"longitude": 7.573,
		"showCurrent": false,
		"showDetails": false,
		"showForecast": true
	}`)

	renderer := NewStaticRenderer(client, themeStore)
	got, err := renderer(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)

	if !strings.Contains(html, "forecast-box") {
		t.Error("expected forecast-box class")
	}
	if !strings.Contains(html, "Fri") {
		t.Errorf("expected day name 'Fri' in output, got: %s", html)
	}
	if !strings.Contains(html, "Sat") {
		t.Errorf("expected day name 'Sat' in output, got: %s", html)
	}
	if !strings.Contains(html, "ti ti-sun-low") {
		t.Errorf("expected resolved icon class for weather code 2, got: %s", html)
	}
	if strings.Contains(html, "Basel") {
		t.Error("forecast-only should not show city name")
	}
	if strings.Contains(html, "18.5") {
		t.Error("forecast-only should not show current temperature")
	}
}

func TestRenderStatic_WithHourly(t *testing.T) {
	srv := newTestWeatherServer()
	defer srv.Close()

	client := weatherpkg.NewClient(&http.Client{}, weatherpkg.WithBaseURL(srv.URL))
	themeStore := themes.NewStore("")
	config := json.RawMessage(`{
		"city": "Basel",
		"latitude": 47.558,
		"longitude": 7.573,
		"showCurrent": false,
		"showDetails": false,
		"showForecast": false,
		"showHourly": true,
		"hourlyCount": 3
	}`)

	renderer := NewStaticRenderer(client, themeStore)
	got, err := renderer(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "hourly-box") {
		t.Error("expected hourly-box class")
	}
	if !strings.Contains(html, "18") {
		t.Errorf("expected temperature in hourly output, got: %s", html)
	}
}
