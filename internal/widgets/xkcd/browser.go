package xkcd

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

var browserTmpl = template.Must(template.New("xkcd-browser").Parse(browserHTML))

// RenderBrowser renders the comic for the browser stack: a title row with
// the comic number, the strip on a white card (comics are black-on-white
// whatever the theme), with the alt text as its tooltip only. It fetches through the
// same client and mode switch as the image renderer and, like it, returns
// the fetch error rather than rendering a placeholder. It ships no JS or CSS.
func (m *Module) RenderBrowser(config json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
	data, err := loadComic(m.client, config)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := browserTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("xkcd browser render: %w", err)
	}
	return template.HTML(buf.String()), nil
}
