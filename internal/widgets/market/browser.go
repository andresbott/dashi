package market

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"strconv"
	"strings"

	mkt "github.com/andresbott/dashi/internal/providers/market"
	"github.com/andresbott/dashi/internal/widgets"
)

//go:embed browser.html
var browserHTML string

var browserTmpl = template.Must(template.New("market-browser").Parse(browserHTML))

// The browser chart is drawn in a fixed 3:1 coordinate space and stretched
// to the cell width by the SVG (preserveAspectRatio="none"); the stroke
// keeps its width through vector-effect="non-scaling-stroke".
const (
	svgChartWidth  = 300.0
	svgChartHeight = 100.0
)

type browserData struct {
	Configured bool
	Symbol     string
	Name       string
	Currency   string
	quoteText
	ShowChange bool
	Chart      *svgChart // nil when the chart is hidden or has too few points
	RangeLabel string
}

// svgChart holds the path data of the browser area chart.
type svgChart struct {
	ViewBox string
	Line    string // the price curve
	Area    string // the curve closed down to the bottom edge, for the fill
	Up      bool   // the series ended at or above where it started
}

// RenderBrowser renders the market widget for the browser stack: the same
// content as the image stack, styled with Tailwind and the theme's CSS
// variables, with the chart as inline SVG instead of a PNG. It needs no
// JavaScript and no stylesheet of its own, so the widget ships no browser
// assets.
func (m *Module) RenderBrowser(config json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
	cfg, err := parseConfig(config)
	if err != nil {
		return "", err
	}

	var data browserData
	if cfg.Symbol != "" {
		md, err := m.client.GetMarketData(cfg.Symbol, cfg.RangeID)
		if err != nil {
			return "", fmt.Errorf("market fetch: %w", err)
		}
		data = browserData{
			Configured: true,
			Symbol:     md.Quote.Symbol,
			Name:       md.Quote.Name,
			Currency:   md.Quote.Currency,
			quoteText:  formatQuote(md.Quote),
			ShowChange: cfg.ShowChange,
			RangeLabel: rangeLabels[cfg.RangeID],
		}
		if cfg.ShowChart && len(md.Points) > 1 {
			data.Chart = buildSVGChart(md.Points)
		}
	}

	var buf bytes.Buffer
	if err := browserTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("market browser render: %w", err)
	}
	return template.HTML(buf.String()), nil
}

// buildSVGChart scales the closes into the SVG coordinate space with the
// same 15% vertical padding and smoothing as the PNG chart (chart.go), so
// both stacks draw the same curve. It needs at least two points.
func buildSVGChart(points []mkt.PricePoint) *svgChart {
	minPrice, maxPrice := points[0].Close, points[0].Close
	for _, p := range points[1:] {
		minPrice = min(minPrice, p.Close)
		maxPrice = max(maxPrice, p.Close)
	}
	priceRange := maxPrice - minPrice
	if priceRange < 0.01 {
		priceRange = 1
	}
	minPrice -= priceRange * 0.15
	priceRange *= 1.3

	n := len(points)
	x := func(i int) float64 { return float64(i) / float64(n-1) * svgChartWidth }
	y := func(price float64) float64 { return (1 - (price-minPrice)/priceRange) * svgChartHeight }

	var line strings.Builder
	line.WriteString("M" + coord(x(0), y(points[0].Close)))
	for i := 0; i < n-1; i++ {
		x0, y0 := x(i), y(points[i].Close)
		x1, y1 := x(i+1), y(points[i+1].Close)
		if n == 2 {
			line.WriteString(" L" + coord(x1, y1))
			continue
		}
		cx := (x1 - x0) / 3
		line.WriteString(" C" + coord(x0+cx, y0) + " " + coord(x1-cx, y1) + " " + coord(x1, y1))
	}

	area := line.String() + " L" + coord(svgChartWidth, svgChartHeight) + " L" + coord(0, svgChartHeight) + " Z"
	return &svgChart{
		ViewBox: "0 0 " + num(svgChartWidth) + " " + num(svgChartHeight),
		Line:    line.String(),
		Area:    area,
		Up:      points[n-1].Close >= points[0].Close,
	}
}

func coord(x, y float64) string { return num(x) + "," + num(y) }

// num formats a coordinate with at most two decimals and no trailing zeros.
func num(v float64) string { return strconv.FormatFloat(math.Round(v*100)/100, 'f', -1, 64) }
