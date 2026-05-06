package markdown

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"

	"github.com/andresbott/dashi/internal/data"
	"github.com/andresbott/dashi/internal/data/notes"
	"github.com/andresbott/dashi/internal/widgets"
)

// NewStaticRenderer returns a StaticRenderer for markdown widgets.
// Config: {"filename": "foo.md"}. The file is read from the shared
// notes store (internal/data/notes). An empty/missing filename
// renders the empty placeholder.
func NewStaticRenderer(store *notes.Store) func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
	return func(config json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
		var filename string
		if len(config) > 0 {
			var obj struct {
				Filename string `json:"filename"`
			}
			if err := json.Unmarshal(config, &obj); err != nil {
				return "", fmt.Errorf("markdown config: %w", err)
			}
			filename = obj.Filename
		}

		if filename == "" {
			return template.HTML(`<div class="widget-markdown"><p class="md-not-found">Markdown file not found</p></div>`), nil
		}

		html, err := store.GetHTML(filename)
		if err != nil {
			if errors.Is(err, data.ErrNotFound) {
				return template.HTML(`<div class="widget-markdown"><p class="md-not-found">Markdown file not found</p></div>`), nil
			}
			return "", fmt.Errorf("markdown render: %w", err)
		}

		return template.HTML(`<div class="widget-markdown">` + string(html) + `</div>`), nil
	}
}
