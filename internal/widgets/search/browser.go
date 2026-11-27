package search

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/andresbott/dashi/internal/widgets"
)

//go:embed browser.html
var browserHTML string

var browserTmpl = template.Must(template.New("search-browser").Parse(browserHTML))

type browserData struct {
	view
	IconFontFamily string
	SearchIcon     string
	ArrowIcon      string
}

// RenderBrowser renders the search widget for the browser stack: a plain
// GET form to the configured engine that opens results in a new tab. It
// needs no JavaScript, so the widget ships no browser assets.
func (m *Module) RenderBrowser(config json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
	v, err := resolve(config)
	if err != nil {
		return "", err
	}
	data := browserData{view: v, IconFontFamily: iconFontFamily, SearchIcon: searchGlyph, ArrowIcon: arrowGlyph}

	var buf bytes.Buffer
	if err := browserTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("search browser render: %w", err)
	}
	return template.HTML(buf.String()), nil
}
