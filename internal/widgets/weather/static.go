package weather

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

	_ "embed"

	"github.com/andresbott/dashi/internal/themes"
	weatherpkg "github.com/andresbott/dashi/internal/weather"
	"github.com/andresbott/dashi/internal/widgets"
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
	CompactDescription *bool `json:"compactDescription"`
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
	ShowAirQuality bool `json:"showAirQuality"`
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

		data := weatherData{
			Compact:            cfg.Compact,
			CompactCity:        compactCity,
			CompactFeelsLike:   cfg.CompactFeelsLike,
			CompactDescription: compactDesc,
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

		var extraInfo []extraInfoItem
		if cfg.ShowSunrise && len(wd.Forecast) > 0 && wd.Forecast[0].Sunrise != "" {
			sunrise := wd.Forecast[0].Sunrise
			if parts := strings.SplitN(sunrise, "T", 2); len(parts) == 2 {
				sunrise = parts[1]
			}
			extraInfo = append(extraInfo, extraInfoItem{IconHTML: resolveIconHTML(themeStore, themeName, "sunrise"), Title: "Sunrise", Value: sunrise})
		}
		if cfg.ShowSunset && len(wd.Forecast) > 0 && wd.Forecast[0].Sunset != "" {
			sunset := wd.Forecast[0].Sunset
			if parts := strings.SplitN(sunset, "T", 2); len(parts) == 2 {
				sunset = parts[1]
			}
			extraInfo = append(extraInfo, extraInfoItem{IconHTML: resolveIconHTML(themeStore, themeName, "sunset"), Title: "Sunset", Value: sunset})
		}
		if cfg.ShowWind {
			extraInfo = append(extraInfo, extraInfoItem{IconHTML: resolveIconHTML(themeStore, themeName, "wind"), Title: "Wind", Value: fmt.Sprintf("%.0f km/h", wd.Current.WindSpeed)})
		}
		if cfg.ShowHumidity {
			extraInfo = append(extraInfo, extraInfoItem{IconHTML: resolveIconHTML(themeStore, themeName, "humidity"), Title: "Humidity", Value: fmt.Sprintf("%d%%", wd.Current.Humidity)})
		}
		if cfg.ShowPressure {
			extraInfo = append(extraInfo, extraInfoItem{IconHTML: resolveIconHTML(themeStore, themeName, "pressure"), Title: "Pressure", Value: fmt.Sprintf("%.0f hPa", wd.Current.Pressure)})
		}
		if cfg.ShowUV && len(wd.Forecast) > 0 {
			extraInfo = append(extraInfo, extraInfoItem{IconHTML: resolveIconHTML(themeStore, themeName, "uv-index"), Title: "UV Index", Value: fmt.Sprintf("%.1f", wd.Forecast[0].UVIndex)})
		}
		if cfg.ShowVisibility {
			extraInfo = append(extraInfo, extraInfoItem{IconHTML: resolveIconHTML(themeStore, themeName, "visibility"), Title: "Visibility", Value: fmt.Sprintf("%.0f km", wd.Current.Visibility)})
		}
		if cfg.ShowAirQuality && wd.AirQuality != nil {
			extraInfo = append(extraInfo, extraInfoItem{IconHTML: resolveIconHTML(themeStore, themeName, "air-quality"), Title: "AQI", Value: fmt.Sprintf("%d", wd.AirQuality.EuropeanAQI)})
		}
		data.ExtraInfo = extraInfo

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
			forecastDays := cfg.ForecastDays
			if forecastDays <= 0 {
				forecastDays = 7
			}
			for i, f := range wd.Forecast {
				if i >= forecastDays {
					break
				}
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
