package pageindicator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"

	_ "embed"

	"github.com/andresbott/dashi/internal/widgets"
)

//go:embed pageindicator.html
var pageIndicatorHTML string

var tmpl = template.Must(template.New("pageindicator").Parse(pageIndicatorHTML))

type dot struct {
	Active bool
}

type pageIndicatorData struct {
	Dots []dot
}

// NewStaticRenderer returns a StaticRenderer for page indicator widgets.
func NewStaticRenderer() func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
	return func(_ json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
		total := ctx.TotalPages
		if total < 1 {
			total = 1
		}

		dots := make([]dot, total)
		for i := range dots {
			dots[i] = dot{Active: i == ctx.PageIndex}
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, pageIndicatorData{Dots: dots}); err != nil {
			return "", fmt.Errorf("page-indicator render: %w", err)
		}
		return template.HTML(buf.String()), nil
	}
}
