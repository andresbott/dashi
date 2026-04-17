package markdown

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/yuin/goldmark"
)

// NewStaticRenderer returns a StaticRenderer for markdown widgets.
// The config is a plain JSON string with the filename (e.g., "notes.md").
// The file is read from {dashboardDir}/md/{filename}.
func NewStaticRenderer(store *dashboard.Store) func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
	md := goldmark.New()

	return func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
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

		data, _, err := store.GetAsset(ctx.DashboardID, "md/"+filename)
		if err != nil {
			return template.HTML(`<div class="widget-markdown"><p class="md-not-found">Markdown file not found</p></div>`), nil
		}

		var buf bytes.Buffer
		if err := md.Convert(data, &buf); err != nil {
			return "", fmt.Errorf("markdown render: %w", err)
		}

		return template.HTML(`<div class="widget-markdown">` + buf.String() + `</div>`), nil
	}
}
