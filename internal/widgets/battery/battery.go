package battery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"

	_ "embed"

	"github.com/andresbott/dashi/internal/widgets"
)

//go:embed battery.html
var batteryHTML string

var tmpl = template.Must(template.New("battery").Parse(batteryHTML))

type batteryData struct {
	Value int
}

// NewStaticRenderer returns a StaticRenderer that displays a battery percentage
// read from the "battery" URL query parameter (0–100).
func NewStaticRenderer() func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
	return func(_ json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
		raw := ctx.QueryParams["battery"]
		value, err := strconv.Atoi(raw)
		if err != nil {
			value = 0
		}
		if value < 0 {
			value = 0
		}
		if value > 100 {
			value = 100
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, batteryData{Value: value}); err != nil {
			return "", fmt.Errorf("battery render: %w", err)
		}
		return template.HTML(buf.String()), nil
	}
}
