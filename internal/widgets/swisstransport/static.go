package swisstransport

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"strings"
	"time"

	_ "embed"

	tr "github.com/andresbott/dashi/internal/swisstransport"
	"github.com/andresbott/dashi/internal/widgets"
)

//go:embed static.html
var staticHTML string

var tmpl = template.Must(template.New("transport").Parse(staticHTML))

type transportConfig struct {
	StationID   string `json:"stationId"`
	StationName string `json:"stationName"`
	Limit       int    `json:"limit"`
}

type departureRow struct {
	Icon        string
	Number      string
	Destination string
	TimeLabel   string
	Platform    string
}

type transportTemplateData struct {
	Configured     bool
	StationName    string
	IconFontFamily string
	Departures     []departureRow
}

func categoryIcon(category string) string {
	switch category {
	case "B":
		return string(rune(0xebe4)) // tabler ti-bus
	case "T":
		return string(rune(0xed96)) // tabler ti-train
	default:
		return string(rune(0xed96)) // tabler ti-train
	}
}

func NewStaticRenderer(client *tr.Client) func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
	return func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
		var cfg transportConfig
		if len(config) > 0 {
			if err := json.Unmarshal(config, &cfg); err != nil {
				return "", fmt.Errorf("transport config: %w", err)
			}
		}

		if cfg.StationID == "" {
			var buf strings.Builder
			if err := tmpl.Execute(&buf, transportTemplateData{Configured: false}); err != nil {
				return "", fmt.Errorf("transport template: %w", err)
			}
			return template.HTML(buf.String()), nil
		}

		limit := cfg.Limit
		if limit <= 0 {
			limit = 5
		}

		deps, err := client.GetDepartures(cfg.StationID, limit)
		if err != nil {
			return "", fmt.Errorf("transport fetch: %w", err)
		}

		now := time.Now()
		rows := make([]departureRow, 0, len(deps))
		for _, d := range deps {
			mins := int(math.Round(time.Until(d.Expected).Minutes()))
			var label string
			if mins <= 0 {
				label = "0'"
			} else if mins > 60 {
				arrival := now.Add(time.Duration(mins) * time.Minute)
				label = fmt.Sprintf("%02d:%02d", arrival.Hour(), arrival.Minute())
			} else {
				label = fmt.Sprintf("%d'", mins)
			}
			rows = append(rows, departureRow{
				Icon:        categoryIcon(d.Category),
				Number:      d.Number,
				Destination: d.Destination,
				TimeLabel:   label,
				Platform:    d.Platform,
			})
		}

		iconFontFamily := ""
		if ctx.Theme != "" {
			iconFontFamily = "icon-font-" + ctx.Theme
		}

		data := transportTemplateData{
			Configured:     true,
			StationName:    cfg.StationName,
			IconFontFamily: iconFontFamily,
			Departures:     rows,
		}

		var buf strings.Builder
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("transport template: %w", err)
		}
		return template.HTML(buf.String()), nil
	}
}
