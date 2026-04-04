package weather

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/themes"
	weatherpkg "github.com/andresbott/dashi/internal/weather"
	"github.com/andresbott/dashi/internal/widgets"
)

func newTestWeatherServer() (*httptest.Server, *httptest.Server) {
	forecast := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"current": map[string]any{
				"temperature_2m":       18.5,
				"relative_humidity_2m": 65,
				"apparent_temperature": 16.2,
				"weather_code":         2,
				"wind_speed_10m":       12.3,
				"surface_pressure":     1013.25,
			},
			"hourly": map[string]any{
				"time":           []string{"2099-01-01T12:00", "2099-01-01T13:00", "2099-01-01T14:00"},
				"temperature_2m": []float64{18.0, 19.0, 20.0},
				"weather_code":   []int{2, 3, 61},
				"visibility":     []float64{24000.0, 20000.0, 15000.0},
			},
			"daily": map[string]any{
				"time":               []string{"2026-04-03", "2026-04-04"},
				"weather_code":       []int{2, 61},
				"temperature_2m_max": []float64{20.0, 15.0},
				"temperature_2m_min": []float64{10.0, 8.0},
				"sunrise":            []string{"2026-04-03T06:30", "2026-04-04T06:31"},
				"sunset":             []string{"2026-04-03T18:45", "2026-04-04T18:44"},
				"uv_index_max":       []float64{5.2, 3.1},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))

	aq := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"current": map[string]any{
				"european_aqi": 42,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))

	return forecast, aq
}

func TestRenderStatic_Basic(t *testing.T) {
	forecastSrv, aqSrv := newTestWeatherServer()
	defer forecastSrv.Close()
	defer aqSrv.Close()

	client := weatherpkg.NewClient(&http.Client{}, weatherpkg.WithBaseURL(forecastSrv.URL), weatherpkg.WithAirQualityBaseURL(aqSrv.URL))
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
	got, err := renderer(config, widgets.RenderContext{})
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
	forecastSrv, aqSrv := newTestWeatherServer()
	defer forecastSrv.Close()
	defer aqSrv.Close()

	client := weatherpkg.NewClient(&http.Client{}, weatherpkg.WithBaseURL(forecastSrv.URL), weatherpkg.WithAirQualityBaseURL(aqSrv.URL))
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
	got, err := renderer(config, widgets.RenderContext{})
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
	forecastSrv, aqSrv := newTestWeatherServer()
	defer forecastSrv.Close()
	defer aqSrv.Close()

	client := weatherpkg.NewClient(&http.Client{}, weatherpkg.WithBaseURL(forecastSrv.URL), weatherpkg.WithAirQualityBaseURL(aqSrv.URL))
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
	got, err := renderer(config, widgets.RenderContext{})
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
	if !strings.Contains(html, "font-family: icon-font-default") {
		t.Errorf("expected icon font-family in output, got: %s", html)
	}
}

func TestRenderStatic_ForecastOnly(t *testing.T) {
	forecastSrv, aqSrv := newTestWeatherServer()
	defer forecastSrv.Close()
	defer aqSrv.Close()

	client := weatherpkg.NewClient(&http.Client{}, weatherpkg.WithBaseURL(forecastSrv.URL), weatherpkg.WithAirQualityBaseURL(aqSrv.URL))
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
	got, err := renderer(config, widgets.RenderContext{})
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
	if !strings.Contains(html, "font-family: icon-font-default") {
		t.Errorf("expected icon font-family in output, got: %s", html)
	}
	if strings.Contains(html, "Basel") {
		t.Error("forecast-only should not show city name")
	}
	if strings.Contains(html, "18.5") {
		t.Error("forecast-only should not show current temperature")
	}
}

func TestRenderStatic_Compact(t *testing.T) {
	forecastSrv, aqSrv := newTestWeatherServer()
	defer forecastSrv.Close()
	defer aqSrv.Close()

	client := weatherpkg.NewClient(&http.Client{}, weatherpkg.WithBaseURL(forecastSrv.URL), weatherpkg.WithAirQualityBaseURL(aqSrv.URL))
	themeStore := themes.NewStore("")
	config := json.RawMessage(`{
		"city": "Basel",
		"latitude": 47.558,
		"longitude": 7.573,
		"compact": true,
		"showForecast": true,
		"showHourly": true
	}`)

	renderer := NewStaticRenderer(client, themeStore)
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "Basel") {
		t.Errorf("expected city name in compact output, got: %s", html)
	}
	if !strings.Contains(html, "18.5") {
		t.Errorf("expected temperature in compact output, got: %s", html)
	}
	if !strings.Contains(html, "weather-compact") {
		t.Errorf("expected weather-compact class in output, got: %s", html)
	}
	if strings.Contains(html, "forecast-box") {
		t.Error("compact mode should not contain forecast boxes")
	}
	if strings.Contains(html, "hourly-box") {
		t.Error("compact mode should not contain hourly boxes")
	}
}

func TestRenderStatic_WithHourly(t *testing.T) {
	forecastSrv, aqSrv := newTestWeatherServer()
	defer forecastSrv.Close()
	defer aqSrv.Close()

	client := weatherpkg.NewClient(&http.Client{}, weatherpkg.WithBaseURL(forecastSrv.URL), weatherpkg.WithAirQualityBaseURL(aqSrv.URL))
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
	got, err := renderer(config, widgets.RenderContext{})
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

func TestRenderStatic_ImageThemeUsesFilePath(t *testing.T) {
	forecastSrv, aqSrv := newTestWeatherServer()
	defer forecastSrv.Close()
	defer aqSrv.Close()

	// Create a temp image theme with a dummy icon file
	themeDir := t.TempDir()
	themePath := filepath.Join(themeDir, "imgtheme")
	iconsDir := filepath.Join(themePath, "widgets", "weather", "icons")
	if err := os.MkdirAll(iconsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Write a minimal PNG file (1x1 pixel)
	if err := os.WriteFile(filepath.Join(iconsDir, "partly-cloudy.png"), testPNG(), 0o644); err != nil {
		t.Fatal(err)
	}
	// Write theme manifest
	manifest := []byte("name: imgtheme\ndescription: test image theme\ntype: image\n")
	if err := os.WriteFile(filepath.Join(themePath, "theme.yaml"), manifest, 0o644); err != nil {
		t.Fatal(err)
	}

	client := weatherpkg.NewClient(&http.Client{}, weatherpkg.WithBaseURL(forecastSrv.URL), weatherpkg.WithAirQualityBaseURL(aqSrv.URL))
	themeStore := themes.NewStore(themeDir)
	config := json.RawMessage(`{
		"city": "Basel",
		"latitude": 47.558,
		"longitude": 7.573,
		"compact": true,
		"showDetails": false,
		"showForecast": false
	}`)

	renderer := NewStaticRenderer(client, themeStore)
	got, err := renderer(config, widgets.RenderContext{Theme: "imgtheme"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	// Should contain an absolute file path, not an API URL
	if strings.Contains(html, "/api/v0/themes/") {
		t.Error("image theme should not use API URL in img src")
	}
	if !strings.Contains(html, iconsDir) {
		t.Errorf("expected file path containing %s in output, got: %s", iconsDir, html)
	}
}

func TestRenderStatic_FontIconCodepoint(t *testing.T) {
	forecastSrv, aqSrv := newTestWeatherServer()
	defer forecastSrv.Close()
	defer aqSrv.Close()

	client := weatherpkg.NewClient(&http.Client{}, weatherpkg.WithBaseURL(forecastSrv.URL), weatherpkg.WithAirQualityBaseURL(aqSrv.URL))
	themeStore := themes.NewStore("")
	config := json.RawMessage(`{
		"city": "Basel",
		"latitude": 47.558,
		"longitude": 7.573,
		"compact": true,
		"showDetails": false,
		"showForecast": false
	}`)

	renderer := NewStaticRenderer(client, themeStore)
	got, err := renderer(config, widgets.RenderContext{Theme: "default"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	// Weather code 2 = "partly-cloudy" → codepoint f237 → Unicode \uf237
	// The HTML should contain a font-family reference, not a CSS class
	if strings.Contains(html, "ti ti-") {
		t.Error("expected Unicode glyph, not CSS class in output")
	}
	if !strings.Contains(html, "font-family") {
		t.Error("expected font-family style in output")
	}
}

func TestRenderStatic_WithExtraInfo(t *testing.T) {
	forecastSrv, aqSrv := newTestWeatherServer()
	defer forecastSrv.Close()
	defer aqSrv.Close()

	client := weatherpkg.NewClient(&http.Client{}, weatherpkg.WithBaseURL(forecastSrv.URL), weatherpkg.WithAirQualityBaseURL(aqSrv.URL))
	themeStore := themes.NewStore("")
	config := json.RawMessage(`{
		"city": "Basel",
		"latitude": 47.558,
		"longitude": 7.573,
		"showSunrise": true,
		"showSunset": true,
		"showWind": true,
		"showHumidity": true,
		"showPressure": true,
		"showUV": true,
		"showVisibility": true,
		"showAirQuality": true
	}`)

	renderer := NewStaticRenderer(client, themeStore)
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)

	checks := map[string]string{
		"extra-info-item": "extra info item class",
		"Sunrise":         "sunrise title",
		"06:30":           "sunrise value",
		"Sunset":          "sunset title",
		"18:45":           "sunset value",
		"Wind":            "wind title",
		"km/h":            "wind unit",
		"Humidity":        "humidity title",
		"65%":             "humidity value",
		"Pressure":        "pressure title",
		"1013 hPa":        "pressure value",
		"UV Index":        "UV title",
		"5.2":             "UV value",
		"Visibility":      "visibility title",
		"24 km":           "visibility value",
		"AQI":             "AQI title",
		"42":              "AQI value",
	}

	for needle, desc := range checks {
		if !strings.Contains(html, needle) {
			t.Errorf("expected %s (%q) in output", desc, needle)
		}
	}
}

func TestRenderStatic_ImageThemeWithExtraInfo(t *testing.T) {
	forecastSrv, aqSrv := newTestWeatherServer()
	defer forecastSrv.Close()
	defer aqSrv.Close()

	// Create a temp image theme with all required icon files
	themeDir := t.TempDir()
	themePath := filepath.Join(themeDir, "imgtheme")
	iconsDir := filepath.Join(themePath, "widgets", "weather", "icons")
	if err := os.MkdirAll(iconsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	allIcons := []string{
		"partly-cloudy", "sunrise", "sunset", "wind",
		"humidity", "pressure", "uv-index", "visibility", "air-quality",
	}
	for _, icon := range allIcons {
		if err := os.WriteFile(filepath.Join(iconsDir, icon+".svg"), []byte("<svg/>"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	manifest := []byte("name: imgtheme\ndescription: test image theme\ntype: image\n")
	if err := os.WriteFile(filepath.Join(themePath, "theme.yaml"), manifest, 0o644); err != nil {
		t.Fatal(err)
	}

	client := weatherpkg.NewClient(&http.Client{}, weatherpkg.WithBaseURL(forecastSrv.URL), weatherpkg.WithAirQualityBaseURL(aqSrv.URL))
	themeStore := themes.NewStore(themeDir)
	config := json.RawMessage(`{
		"city": "Basel",
		"latitude": 47.558,
		"longitude": 7.573,
		"showSunrise": true,
		"showSunset": true,
		"showWind": true,
		"showHumidity": true,
		"showPressure": true,
		"showUV": true,
		"showVisibility": true,
		"showAirQuality": true
	}`)

	renderer := NewStaticRenderer(client, themeStore)
	got, err := renderer(config, widgets.RenderContext{Theme: "imgtheme"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)

	// Detail icons should use <img> tags pointing to icon files
	for _, icon := range []string{"sunrise", "sunset", "wind", "humidity", "pressure", "uv-index", "visibility", "air-quality"} {
		if !strings.Contains(html, icon+".svg") {
			t.Errorf("expected image path for %q icon in output", icon)
		}
	}
}

// testPNG returns the bytes of a minimal valid 1x1 red PNG.
func testPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}
