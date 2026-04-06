package market

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	_ "embed"

	mkt "github.com/andresbott/dashi/internal/market"
	"github.com/andresbott/dashi/internal/widgets"
)

//go:embed static.html
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
}

var rangeLabels = map[string]string{
	"1d": "1 Day", "5d": "5 Days", "1mo": "1 Month",
	"3mo": "3 Months", "6mo": "6 Months", "1y": "1 Year",
}

func NewStaticRenderer(client *mkt.Client) func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
	return func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
		var cfg marketConfig
		if len(config) > 0 {
			if err := json.Unmarshal(config, &cfg); err != nil {
				return "", fmt.Errorf("market config: %w", err)
			}
		}

		if cfg.Symbol == "" {
			var buf strings.Builder
			if err := tmpl.Execute(&buf, marketTemplateData{Configured: false}); err != nil {
				return "", fmt.Errorf("market template: %w", err)
			}
			return template.HTML(buf.String()), nil
		}

		rangeID := cfg.Range
		if rangeID == "" {
			rangeID = "1mo"
		}

		showChart := true
		if cfg.ShowChart != nil {
			showChart = *cfg.ShowChart
		}
		showChange := true
		if cfg.ShowChange != nil {
			showChange = *cfg.ShowChange
		}

		data, err := client.GetMarketData(cfg.Symbol, rangeID)
		if err != nil {
			return "", fmt.Errorf("market fetch: %w", err)
		}

		changeColor := "#22c55e"
		changePrefix := "+"
		if data.Quote.Change < 0 {
			changeColor = "#ef4444"
			changePrefix = ""
		}

		td := marketTemplateData{
			Configured:             true,
			Symbol:                 data.Quote.Symbol,
			Name:                   data.Quote.Name,
			Currency:               data.Quote.Currency,
			PriceFormatted:         fmt.Sprintf("%.2f", data.Quote.Price),
			ChangeFormatted:        fmt.Sprintf("%s%.2f", changePrefix, data.Quote.Change),
			ChangePercentFormatted: fmt.Sprintf("%s%.2f%%", changePrefix, data.Quote.ChangePercent),
			ChangeColor:            changeColor,
			ShowChange:             showChange,
			ShowChart:              showChart,
			RangeLabel:             rangeLabels[rangeID],
		}

		if showChart && len(data.Points) > 1 {
			chartPNG, err := generateChart(data.Points, chartOptions{Width: 400, Height: 150})
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
