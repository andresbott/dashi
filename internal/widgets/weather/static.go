package weather

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"image/color"
	"math"
	"strconv"
	"strings"
	"time"

	_ "embed"

	"github.com/andresbott/dashi/internal/themes"
	weatherpkg "github.com/andresbott/dashi/internal/weather"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/andresbott/dashi/internal/widgets/weather/chart"
)

//go:embed static.html
var staticHTML string

var tmpl = template.Must(template.New("weather").Parse(staticHTML))

type weatherConfig struct {
	City         string  `json:"city"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Compact            bool  `json:"compact"`
	CompactCity        *bool `json:"compactCity"`
	CompactFeelsLike   bool  `json:"compactFeelsLike"`
	CompactDescription *bool  `json:"compactDescription"`
	CompactAlign       string `json:"compactAlign"`
	ShowCurrent        *bool `json:"showCurrent"`
	ShowDetails  bool    `json:"showDetails"`
	ShowHourly   bool    `json:"showHourly"`
	HourlyCount  int     `json:"hourlyCount"`
	HourlySlots  int     `json:"hourlySlots"`
	ShowForecast bool   `json:"showForecast"`
	ForecastDays int    `json:"forecastDays"`
	ShowSunrise    bool `json:"showSunrise"`
	ShowSunset     bool `json:"showSunset"`
	ShowWind       bool `json:"showWind"`
	ShowHumidity   bool `json:"showHumidity"`
	ShowPressure   bool `json:"showPressure"`
	ShowUV         bool `json:"showUV"`
	ShowVisibility bool `json:"showVisibility"`
	ShowAirQuality bool   `json:"showAirQuality"`
	ShowGraph      bool   `json:"showGraph"`
	GraphHours     int    `json:"graphHours"`
	GraphTempColor string `json:"graphTempColor"`
	GraphRainColor string `json:"graphRainColor"`
	GraphBgColor   string `json:"graphBgColor"`
	GraphHeight     int    `json:"graphHeight"`
	GraphShowTemp   *bool  `json:"graphShowTemp"`
	GraphShowRain   *bool  `json:"graphShowRain"`
}

type hourlyEntry struct {
	Time        string
	IconHTML    template.HTML
	Temperature float64
}

type extraInfoItem struct {
	IconHTML template.HTML
	Title    string
	Value    string
}

type weatherData struct {
	Compact            bool
	CompactCity        bool
	CompactFeelsLike   bool
	CompactDescription bool
	CompactAlign       string
	City               string
	Temperature  float64
	Description  string
	FeelsLike    float64
	Humidity     int
	WindSpeed    float64
	ShowCurrent  bool
	ShowDetails  bool
	ShowHourly   bool
	ShowForecast bool
	Hourly       []hourlyEntry
	Forecast     []forecastDay
	IconHTML     template.HTML
	ExtraInfo    []extraInfoItem
	ShowGraph       bool
	GraphImage      string // base64-encoded PNG
	GraphTempMin    string
	GraphTempMax    string
	GraphTimeLabels []string
	GraphShowTemp   bool
	GraphShowRain   bool
}

type forecastDay struct {
	Date        string
	DayName     string
	IconHTML    template.HTML
	TempMin     float64
	TempMax     float64
	Description string
}

// NewStaticRenderer returns a StaticRenderer for weather widgets.
// It uses the provided weather client to fetch live data and the
// theme store to resolve canonical icon names.
func NewStaticRenderer(client *weatherpkg.Client, themeStore *themes.Store) func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
	return func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
		var cfg weatherConfig
		if len(config) > 0 {
			if err := json.Unmarshal(config, &cfg); err != nil {
				return "", fmt.Errorf("weather config: %w", err)
			}
		}

		wd, err := client.GetWeather(cfg.Latitude, cfg.Longitude)
		if err != nil {
			return "", fmt.Errorf("weather fetch: %w", err)
		}

		themeName := ctx.Theme
		if themeName == "" {
			themeName = "default"
		}

		showCurrent := cfg.ShowCurrent == nil || *cfg.ShowCurrent

		currentIconHTML := resolveIconHTML(themeStore, themeName, wd.Current.Icon)

		compactCity := cfg.CompactCity == nil || *cfg.CompactCity
		compactDesc := cfg.CompactDescription == nil || *cfg.CompactDescription

		compactAlign := cfg.CompactAlign
		if compactAlign == "" {
			compactAlign = "left"
		}

		data := weatherData{
			Compact:            cfg.Compact,
			CompactCity:        compactCity,
			CompactFeelsLike:   cfg.CompactFeelsLike,
			CompactDescription: compactDesc,
			CompactAlign:       compactAlign,
			City:               cfg.City,
			Temperature:  wd.Current.Temperature,
			Description:  wd.Current.Description,
			FeelsLike:    wd.Current.FeelsLike,
			Humidity:     wd.Current.Humidity,
			WindSpeed:    wd.Current.WindSpeed,
			ShowCurrent:  showCurrent,
			ShowDetails:  cfg.ShowDetails,
			ShowHourly:   cfg.ShowHourly && !cfg.Compact,
			ShowForecast: cfg.ShowForecast && !cfg.Compact,
			IconHTML:     currentIconHTML,
		}

		data.ExtraInfo = buildExtraInfo(cfg, wd, themeStore, themeName)

		if cfg.ShowHourly {
			data.Hourly = buildHourlyEntries(cfg, wd.Hourly, themeStore, themeName)
		}

		if cfg.ShowGraph && !cfg.Compact {
			addGraphData(&data, cfg, wd.Hourly)
		}

		if cfg.ShowForecast {
			data.Forecast = buildForecastDays(cfg, wd.Forecast, themeStore, themeName)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("weather render: %w", err)
		}
		return template.HTML(buf.String()), nil
	}
}

// NewStaticCompactRenderer returns a StaticRenderer for compact weather widgets.
// It works like NewStaticRenderer but always renders in compact mode.
func NewStaticCompactRenderer(client *weatherpkg.Client, themeStore *themes.Store) func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
	inner := NewStaticRenderer(client, themeStore)
	return func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
		// Inject compact:true into the config
		var raw map[string]any
		if len(config) > 0 {
			if err := json.Unmarshal(config, &raw); err != nil {
				raw = make(map[string]any)
			}
		} else {
			raw = make(map[string]any)
		}
		raw["compact"] = true
		modified, err := json.Marshal(raw)
		if err != nil {
			return "", fmt.Errorf("marshal compact config: %w", err)
		}
		return inner(modified, ctx)
	}
}

func resolveIconHTML(store *themes.Store, themeName, canonicalName string) template.HTML {
	resolved, err := store.ResolveIcon(themeName, canonicalName)
	if err != nil {
		return template.HTML(`<i class="ti ti-cloud-question"></i>`)
	}
	switch resolved.Type {
	case themes.ThemeTypeFont:
		if resolved.Codepoint != "" {
			cp, err := strconv.ParseUint(resolved.Codepoint, 16, 32)
			if err == nil {
				return template.HTML(fmt.Sprintf(
					`<span style="font-family: icon-font-%s; font-size: 1.4em;">%s</span>`,
					template.HTMLEscapeString(themeName),
					string(rune(cp))))
			}
		}
		// fallback: CSS class (won't render icons in PNG but harmless)
		return template.HTML(fmt.Sprintf(`<i class="%s"></i>`,
			template.HTMLEscapeString(resolved.CSSClass)))
	case themes.ThemeTypeImage:
		return template.HTML(fmt.Sprintf(`<img src="%s" alt="%s" style="width:1.4em;height:1.4em;">`,
			template.HTMLEscapeString(resolved.FilePath),
			template.HTMLEscapeString(canonicalName)))
	default:
		return template.HTML(`<i class="ti ti-cloud-question"></i>`)
	}
}

func parseHexColor(hex string, fallback color.NRGBA) color.NRGBA {
	if len(hex) == 0 {
		return fallback
	}
	if hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return fallback
	}
	var r, g, b uint8
	_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return fallback
	}
	return color.NRGBA{R: r, G: g, B: b, A: 255}
}

// buildExtraInfo collects and formats extra weather information (sunrise, sunset, etc).
func buildExtraInfo(cfg weatherConfig, wd weatherpkg.WeatherData, themeStore *themes.Store, themeName string) []extraInfoItem {
	var items []extraInfoItem
	if cfg.ShowSunrise && len(wd.Forecast) > 0 && wd.Forecast[0].Sunrise != "" {
		sunrise := extractTime(wd.Forecast[0].Sunrise)
		items = append(items, extraInfoItem{
			IconHTML: resolveIconHTML(themeStore, themeName, "sunrise"),
			Title:    "Sunrise",
			Value:    sunrise,
		})
	}
	if cfg.ShowSunset && len(wd.Forecast) > 0 && wd.Forecast[0].Sunset != "" {
		sunset := extractTime(wd.Forecast[0].Sunset)
		items = append(items, extraInfoItem{
			IconHTML: resolveIconHTML(themeStore, themeName, "sunset"),
			Title:    "Sunset",
			Value:    sunset,
		})
	}
	if cfg.ShowWind {
		items = append(items, extraInfoItem{
			IconHTML: resolveIconHTML(themeStore, themeName, "wind"),
			Title:    "Wind",
			Value:    fmt.Sprintf("%.0f km/h", wd.Current.WindSpeed),
		})
	}
	if cfg.ShowHumidity {
		items = append(items, extraInfoItem{
			IconHTML: resolveIconHTML(themeStore, themeName, "humidity"),
			Title:    "Humidity",
			Value:    fmt.Sprintf("%d%%", wd.Current.Humidity),
		})
	}
	if cfg.ShowPressure {
		items = append(items, extraInfoItem{
			IconHTML: resolveIconHTML(themeStore, themeName, "pressure"),
			Title:    "Pressure",
			Value:    fmt.Sprintf("%.0f hPa", wd.Current.Pressure),
		})
	}
	if cfg.ShowUV && len(wd.Forecast) > 0 {
		items = append(items, extraInfoItem{
			IconHTML: resolveIconHTML(themeStore, themeName, "uv-index"),
			Title:    "UV Index",
			Value:    fmt.Sprintf("%.1f", wd.Forecast[0].UVIndex),
		})
	}
	if cfg.ShowVisibility {
		items = append(items, extraInfoItem{
			IconHTML: resolveIconHTML(themeStore, themeName, "visibility"),
			Title:    "Visibility",
			Value:    fmt.Sprintf("%.0f km", wd.Current.Visibility),
		})
	}
	if cfg.ShowAirQuality && wd.AirQuality != nil {
		items = append(items, extraInfoItem{
			IconHTML: resolveIconHTML(themeStore, themeName, "air-quality"),
			Title:    "AQI",
			Value:    fmt.Sprintf("%d", wd.AirQuality.EuropeanAQI),
		})
	}
	return items
}

// extractTime extracts the time portion from an ISO timestamp string.
func extractTime(timestamp string) string {
	if parts := strings.SplitN(timestamp, "T", 2); len(parts) == 2 {
		return parts[1]
	}
	return timestamp
}

// buildHourlyEntries creates hourly forecast entries for display.
func buildHourlyEntries(cfg weatherConfig, hourly []weatherpkg.HourlyForecast, themeStore *themes.Store, themeName string) []hourlyEntry {
	hourlyCount := cfg.HourlyCount
	if hourlyCount <= 0 {
		hourlyCount = 12
	}
	if hourlyCount > 24 {
		hourlyCount = 24
	}

	var allHourly []weatherpkg.HourlyForecast
	for i, h := range hourly {
		if i >= hourlyCount {
			break
		}
		allHourly = append(allHourly, h)
	}

	slots := cfg.HourlySlots
	if slots <= 0 {
		slots = 6
	}
	if slots > 12 {
		slots = 12
	}
	if slots > len(allHourly) {
		slots = len(allHourly)
	}

	var entries []hourlyEntry
	for i := 0; i < slots; i++ {
		idx := ((i + 1) * len(allHourly) / slots) - 1
		h := allHourly[idx]
		hourTime := h.Time
		if t, err := time.Parse("2006-01-02T15:04", h.Time); err == nil {
			hourTime = t.Format("15:04")
		}
		entries = append(entries, hourlyEntry{
			Time:        hourTime,
			IconHTML:    resolveIconHTML(themeStore, themeName, h.Icon),
			Temperature: h.Temperature,
		})
	}
	return entries
}

// buildForecastDays creates daily forecast entries.
func buildForecastDays(cfg weatherConfig, forecast []weatherpkg.DailyForecast, themeStore *themes.Store, themeName string) []forecastDay {
	forecastDays := cfg.ForecastDays
	if forecastDays <= 0 {
		forecastDays = 7
	}

	var days []forecastDay
	for i, f := range forecast {
		if i >= forecastDays {
			break
		}
		dayName := f.Date
		if t, err := time.Parse("2006-01-02", f.Date); err == nil {
			dayName = t.Format("Mon")
		}
		days = append(days, forecastDay{
			Date:        f.Date,
			DayName:     dayName,
			IconHTML:    resolveIconHTML(themeStore, themeName, f.Icon),
			TempMin:     f.TempMin,
			TempMax:     f.TempMax,
			Description: f.Description,
		})
	}
	return days
}

// addGraphData generates and adds weather graph data to the template data.
func addGraphData(data *weatherData, cfg weatherConfig, hourly []weatherpkg.HourlyForecast) {
	graphHours := cfg.GraphHours
	if graphHours <= 0 {
		graphHours = 24
	}
	if graphHours > len(hourly) {
		graphHours = len(hourly)
	}

	graphPoints := make([]chart.HourlyPoint, 0, graphHours)
	for i := 0; i < graphHours; i++ {
		h := hourly[i]
		t, _ := time.Parse("2006-01-02T15:04", h.Time)
		graphPoints = append(graphPoints, chart.HourlyPoint{
			Time:        t,
			Temperature: h.Temperature,
			RainPercent: h.PrecipitationProbability,
		})
	}

	tempColor := parseHexColor(cfg.GraphTempColor, color.NRGBA{R: 0xFF, G: 0x8C, B: 0x42, A: 0xFF})
	rainColor := parseHexColor(cfg.GraphRainColor, color.NRGBA{R: 0x4A, G: 0x90, B: 0xD9, A: 0xFF})
	bgColor := parseHexColor(cfg.GraphBgColor, color.NRGBA{A: 0})
	if cfg.GraphBgColor == "" || cfg.GraphBgColor == "transparent" {
		bgColor = color.NRGBA{A: 0}
	}

	graphHeight := cfg.GraphHeight
	if graphHeight <= 0 {
		graphHeight = 250
	}

	chartPNG, err := chart.Generate(graphPoints, chart.Options{
		Width:     800,
		Height:    graphHeight,
		TempColor: tempColor,
		RainColor: rainColor,
		BgColor:   bgColor,
	})
	if err != nil || len(chartPNG) == 0 {
		return
	}

	data.ShowGraph = true
	data.GraphImage = base64.StdEncoding.EncodeToString(chartPNG)

	// Compute temp range for labels
	minT, maxT := graphPoints[0].Temperature, graphPoints[0].Temperature
	for _, p := range graphPoints[1:] {
		minT = math.Min(minT, p.Temperature)
		maxT = math.Max(maxT, p.Temperature)
	}
	data.GraphTempMin = fmt.Sprintf("%.0f°", minT)
	data.GraphTempMax = fmt.Sprintf("%.0f°", maxT)

	// Time labels — every 3 hours
	var timeLabels []string
	for i, p := range graphPoints {
		if i%(graphHours/8+1) == 0 {
			timeLabels = append(timeLabels, p.Time.Format("15:04"))
		}
	}
	data.GraphTimeLabels = timeLabels

	data.GraphShowTemp = cfg.GraphShowTemp == nil || *cfg.GraphShowTemp
	data.GraphShowRain = cfg.GraphShowRain == nil || *cfg.GraphShowRain
}
