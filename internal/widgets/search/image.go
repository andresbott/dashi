package search

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/widgets"
)

//go:embed image.html
var imageHTML string

var imageTmpl = template.Must(template.New("search-image").Parse(imageHTML))

type imageData struct {
	view
	IconFontFamily string
	SearchIcon     string
	ArrowIcon      string
	Palette        themes.Palette
}

// renderImage draws the search widget for the image (e-ink) stack. A PNG
// cannot be typed into, so it is a static picture of the bar: icon,
// placeholder text, submit button and the engine label, as in the editor
// preview.
func renderImage(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
	v, err := resolve(config)
	if err != nil {
		return "", err
	}
	data := imageData{
		view:           v,
		IconFontFamily: iconFontFamily,
		SearchIcon:     searchGlyph,
		ArrowIcon:      arrowGlyph,
		Palette:        ctx.EffectivePalette(),
	}

	var buf bytes.Buffer
	if err := imageTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("search render: %w", err)
	}
	return template.HTML(buf.String()), nil
}
