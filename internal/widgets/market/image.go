package market

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	_ "embed"

	mkt "github.com/andresbott/dashi/internal/providers/market"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/widgets"
)

//go:embed image.html
var staticHTML string

var tmpl = template.Must(template.New("market").Parse(staticHTML))

type marketConfig struct {
	Symbol     string `json:"symbol"`
	Range      string `json:"range"`
	ShowChart  *bool  `json:"showChart"`
	ShowChange *bool  `json:"showChange"`
}

type marketTemplateData struct {
	Configured             bool
	Symbol                 string
	Name                   string
	Currency               string
	PriceFormatted         string
	ChangeFormatted        string
	ChangePercentFormatted string
	ChangeColor            string
	ShowChange             bool
	ShowChart              bool
	ChartImage             string
	RangeLabel             string
	Palette                themes.Palette
}

var rangeLabels = map[string]string{
	"1d": "1 Day", "5d": "5 Days", "1mo": "1 Month",
	"3mo": "3 Months", "6mo": "6 Months", "1y": "1 Year",
}

// marketOpts is a market config with its defaults applied.
type marketOpts struct {
	Symbol     string
	RangeID    string
	ShowChart  bool
	ShowChange bool
}

// parseConfig decodes a market config and applies the defaults shared by
// both render stacks: range "1mo", chart and change shown.
func parseConfig(config json.RawMessage) (marketOpts, error) {
	var cfg marketConfig
	if len(config) > 0 {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return marketOpts{}, fmt.Errorf("market config: %w", err)
		}
	}
	rc := marketOpts{Symbol: cfg.Symbol, RangeID: cfg.Range, ShowChart: true, ShowChange: true}
	if rc.RangeID == "" {
		rc.RangeID = "1mo"
	}
	if cfg.ShowChart != nil {
		rc.ShowChart = *cfg.ShowChart
	}
	if cfg.ShowChange != nil {
		rc.ShowChange = *cfg.ShowChange
	}
	return rc, nil
}

// quoteText is a quote formatted for display, shared by both stacks.
type quoteText struct {
	Price         string
	Change        string
	ChangePercent string
	Down          bool // the change is negative
}

func formatQuote(q mkt.Quote) quoteText {
	prefix := "+"
	down := q.Change < 0
	if down {
		prefix = ""
	}
	return quoteText{
		Price:         fmt.Sprintf("%.2f", q.Price),
		Change:        fmt.Sprintf("%s%.2f", prefix, q.Change),
		ChangePercent: fmt.Sprintf("%s%.2f%%", prefix, q.ChangePercent),
		Down:          down,
	}
}

func NewStaticRenderer(client *mkt.Client) func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
	return func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
		cfg, err := parseConfig(config)
		if err != nil {
			return "", err
		}

		if cfg.Symbol == "" {
			var buf strings.Builder
			if err := tmpl.Execute(&buf, marketTemplateData{Configured: false, Palette: ctx.EffectivePalette()}); err != nil {
				return "", fmt.Errorf("market template: %w", err)
			}
			return template.HTML(buf.String()), nil
		}

		rangeID := cfg.RangeID
		showChart := cfg.ShowChart

		data, err := client.GetMarketData(cfg.Symbol, rangeID)
		if err != nil {
			return "", fmt.Errorf("market fetch: %w", err)
		}

		pal := ctx.EffectivePalette()
		qt := formatQuote(data.Quote)
		changeColor := pal.Success
		if qt.Down {
			changeColor = pal.Danger
		}

		td := marketTemplateData{
			Configured:             true,
			Symbol:                 data.Quote.Symbol,
			Name:                   data.Quote.Name,
			Currency:               data.Quote.Currency,
			PriceFormatted:         qt.Price,
			ChangeFormatted:        qt.Change,
			ChangePercentFormatted: qt.ChangePercent,
			ChangeColor:            changeColor,
			ShowChange:             cfg.ShowChange,
			ShowChart:              showChart,
			RangeLabel:             rangeLabels[rangeID],
			Palette:                pal,
		}

		if showChart && len(data.Points) > 1 {
			up, down := paletteLineColors(pal)
			chartPNG, err := generateChart(data.Points, chartOptions{Width: 400, Height: 150, UpColor: up, DownColor: down})
			if err == nil {
				td.ChartImage = base64.StdEncoding.EncodeToString(chartPNG)
			}
		}

		var buf strings.Builder
		if err := tmpl.Execute(&buf, td); err != nil {
			return "", fmt.Errorf("market template: %w", err)
		}
		return template.HTML(buf.String()), nil
	}
}
