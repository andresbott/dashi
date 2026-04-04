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

type pageData struct {
	Name     string
	MaxWidth string
	HAlign   string
	VAlign   string
	Rows     []rowData
}

type rowData struct {
	Title   string
	Height  string
	Width   string
	Widgets []widgetData
}

type widgetData struct {
	Width int
	HTML  template.HTML
}

// Render writes the complete HTML page for the given dashboard to w.
func (r *Renderer) Render(w io.Writer, dash dashboard.Dashboard) error {
	data := pageData{
		Name:     dash.Name,
		MaxWidth: dash.Container.MaxWidth,
		HAlign:   mapHAlign(dash.Container.HorizontalAlign),
		VAlign:   mapVAlign(dash.Container.VerticalAlign),
	}

	for _, row := range dash.Rows {
		rd := rowData{
			Title:  row.Title,
			Height: row.Height,
			Width:  row.Width,
		}
		for _, widget := range row.Widgets {
			rendered, err := r.registry.Render(widget.Type, widget.Config)
			if err != nil {
				return fmt.Errorf("render widget %s (%s): %w", widget.ID, widget.Type, err)
			}
			rd.Widgets = append(rd.Widgets, widgetData{
				Width: widget.Width,
				HTML:  rendered,
			})
		}
		data.Rows = append(data.Rows, rd)
	}

	return masterTmpl.Execute(w, data)
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
