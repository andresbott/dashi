package static

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/widgets"
)

//go:embed master.html
var masterHTML string

var masterTmpl = template.Must(template.New("master").Parse(masterHTML))

// Renderer assembles a full HTML page from a dashboard and its widgets.
type Renderer struct {
	registry *widgets.Registry
}

// NewRenderer creates a Renderer with the given widget registry.
func NewRenderer(registry *widgets.Registry) *Renderer {
	return &Renderer{registry: registry}
}

// RenderData holds the data needed to render a dashboard page as HTML.
type RenderData struct {
	Name          string
	DashboardID   string
	MaxWidth      string
	HAlign        string
	VAlign        string
	Theme         string
	ColorMode     string
	FontFamily    string
	CustomCSS     string
	BackgroundCSS string
	QueryParams   map[string]string
	Rows          []dashboard.Row
	PageIndex     int
	TotalPages    int
}

type pageData struct {
	Name               string
	MaxWidth           string
	HAlign             string
	VAlign             string
	IsDark             bool
	FontFamily         string
	CustomCSS     template.CSS
	BackgroundCSS template.CSS
	Rows          []rowData
}

type rowData struct {
	Title             string
	Height            string
	Width             string
	HasExplicitHeight bool
	Widgets           []widgetData
}

type widgetData struct {
	Width        int
	WidthPercent float64
	HTML         template.HTML
	DebugColor   string
}

var debugColors = []string{"#ffcccc", "#ccffcc", "#ccccff", "#ffffcc", "#ffccff", "#ccffff"}

// Render writes the complete HTML page for the given data to w.
func (r *Renderer) Render(w io.Writer, data RenderData) error {
	pData := pageData{
		Name:               data.Name,
		MaxWidth:           data.MaxWidth,
		HAlign:             mapHAlign(data.HAlign),
		VAlign:             mapVAlign(data.VAlign),
		IsDark:             data.ColorMode == "dark",
		FontFamily:         data.FontFamily,
		CustomCSS:     template.CSS(data.CustomCSS),
		BackgroundCSS: template.CSS(data.BackgroundCSS),
	}

	for _, row := range data.Rows {
		hasExplicitHeight := row.Height != "" && row.Height != "auto"

		// Skip empty rows with auto height — they should not add any space.
		if len(row.Widgets) == 0 && !hasExplicitHeight {
			continue
		}

		rd := rowData{
			Title:             row.Title,
			Height:            row.Height,
			Width:             row.Width,
			HasExplicitHeight: hasExplicitHeight,
		}
		ctx := widgets.RenderContext{DashboardID: data.DashboardID, Theme: data.Theme, QueryParams: data.QueryParams, PageIndex: data.PageIndex, TotalPages: data.TotalPages}
		debug := data.QueryParams["debug"] == "1"
		colorIdx := 0
		for _, widget := range row.Widgets {
			rendered, err := r.registry.Render(widget.Type, widget.Config, ctx)
			if err != nil {
				return fmt.Errorf("render widget %s (%s): %w", widget.ID, widget.Type, err)
			}
			w := widget.Width
			if w < 1 {
				w = 12
			}
			wd := widgetData{
				Width:        w,
				WidthPercent: float64(w) / 12.0 * 100.0,
				HTML:         rendered,
			}
			if debug {
				wd.DebugColor = debugColors[colorIdx%len(debugColors)]
				colorIdx++
			}
			rd.Widgets = append(rd.Widgets, wd)
		}
		pData.Rows = append(pData.Rows, rd)
	}

	return masterTmpl.Execute(w, pData)
}

func mapHAlign(align string) string {
	switch align {
	case "left":
		return "flex-start"
	case "right":
		return "flex-end"
	case "center":
		return "center"
	default:
		return "center"
	}
}

func mapVAlign(align string) string {
	switch align {
	case "top":
		return "flex-start"
	case "bottom":
		return "flex-end"
	case "center":
		return "center"
	default:
		return "flex-start"
	}
}
