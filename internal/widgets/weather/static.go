package weather

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"time"

	_ "embed"

	"github.com/andresbott/dashi/internal/themes"
	weatherpkg "github.com/andresbott/dashi/internal/weather"
)

//go:embed static.html
var staticHTML string

var tmpl = template.Must(template.New("weather").Parse(staticHTML))

type weatherConfig struct {
	City         string  `json:"city"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Compact      bool    `json:"compact"`
	ShowCurrent  *bool   `json:"showCurrent"`
	ShowDetails  bool    `json:"showDetails"`
	ShowHourly   bool    `json:"showHourly"`
	HourlyCount  int     `json:"hourlyCount"`
	HourlySlots  int     `json:"hourlySlots"`
	ShowForecast bool    `json:"showForecast"`
	Layout       string  `json:"layout"`
	IconTheme    string  `json:"iconTheme"`
}

type hourlyEntry struct {
	Time        string
	IconHTML    template.HTML
	Temperature float64
}

type weatherData struct {
	City         string
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
func NewStaticRenderer(client *weatherpkg.Client, themeStore *themes.Store) func(json.RawMessage) (template.HTML, error) {
	return func(config json.RawMessage) (template.HTML, error) {
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

		themeName := cfg.IconTheme
		if themeName == "" {
			themeName = "default"
		}

		showCurrent := cfg.ShowCurrent == nil || *cfg.ShowCurrent

		currentIconHTML := resolveIconHTML(themeStore, themeName, wd.Current.Icon)

		data := weatherData{
			City:         cfg.City,
			Temperature:  wd.Current.Temperature,
			Description:  wd.Current.Description,
			FeelsLike:    wd.Current.FeelsLike,
			Humidity:     wd.Current.Humidity,
			WindSpeed:    wd.Current.WindSpeed,
			ShowCurrent:  showCurrent,
			ShowDetails:  cfg.ShowDetails,
			ShowHourly:   cfg.ShowHourly,
			ShowForecast: cfg.ShowForecast,
			IconHTML:     currentIconHTML,
		}

		if cfg.ShowHourly {
			hourlyCount := cfg.HourlyCount
			if hourlyCount <= 0 {
				hourlyCount = 12
			}
			if hourlyCount > 24 {
				hourlyCount = 24
			}
			// Collect all entries within the hour range
			var allHourly []weatherpkg.HourlyForecast
			for i, h := range wd.Hourly {
				if i >= hourlyCount {
					break
				}
				allHourly = append(allHourly, h)
			}
			// Pick evenly spaced slots
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
			for i := 0; i < slots; i++ {
				idx := ((i + 1) * len(allHourly) / slots) - 1
				h := allHourly[idx]
				hourTime := h.Time
				if t, err := time.Parse("2006-01-02T15:04", h.Time); err == nil {
					hourTime = t.Format("15:04")
				}
				data.Hourly = append(data.Hourly, hourlyEntry{
					Time:        hourTime,
					IconHTML:    resolveIconHTML(themeStore, themeName, h.Icon),
					Temperature: h.Temperature,
				})
			}
		}

		if cfg.ShowForecast {
			for _, f := range wd.Forecast {
				dayName := f.Date
				if t, err := time.Parse("2006-01-02", f.Date); err == nil {
					dayName = t.Format("Mon")
				}
				data.Forecast = append(data.Forecast, forecastDay{
					Date:        f.Date,
					DayName:     dayName,
					IconHTML:    resolveIconHTML(themeStore, themeName, f.Icon),
					TempMin:     f.TempMin,
					TempMax:     f.TempMax,
					Description: f.Description,
				})
			}
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("weather render: %w", err)
		}
		return template.HTML(buf.String()), nil
	}
}

func resolveIconHTML(store *themes.Store, themeName, canonicalName string) template.HTML {
	resolved, err := store.ResolveIcon(themeName, canonicalName)
	if err != nil {
		return template.HTML(`<i class="ti ti-cloud-question"></i>`)
	}
	switch resolved.Type {
	case themes.ThemeTypeFont:
		return template.HTML(fmt.Sprintf(`<i class="%s"></i>`, template.HTMLEscapeString(resolved.CSSClass)))
	case themes.ThemeTypeImage:
		return template.HTML(fmt.Sprintf(`<img src="/api/v0/themes/%s/icons/%s" alt="%s" style="width:1.4em;height:1.4em;">`,
			template.HTMLEscapeString(themeName),
			template.HTMLEscapeString(canonicalName),
			template.HTMLEscapeString(canonicalName)))
	default:
		return template.HTML(`<i class="ti ti-cloud-question"></i>`)
	}
}
