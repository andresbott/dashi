package bookmark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"

	_ "embed"
)

//go:embed bookmark.html
var bookmarkHTML string

var tmpl = template.Must(template.New("bookmark").Parse(bookmarkHTML))

type bookmarkConfig struct {
	URL      string `json:"url"`
	Icon     string `json:"icon"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

type bookmarkData struct {
	URL      string
	Title    string
	Subtitle string
}

// NewStaticRenderer returns a StaticRenderer for bookmark widgets.
func NewStaticRenderer() func(json.RawMessage) (template.HTML, error) {
	return func(config json.RawMessage) (template.HTML, error) {
		var cfg bookmarkConfig
		if len(config) > 0 {
			if err := json.Unmarshal(config, &cfg); err != nil {
				return "", fmt.Errorf("bookmark config: %w", err)
			}
		}

		data := bookmarkData{
			URL:      cfg.URL,
			Title:    cfg.Title,
			Subtitle: cfg.Subtitle,
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("bookmark render: %w", err)
		}
		return template.HTML(buf.String()), nil
	}
}
