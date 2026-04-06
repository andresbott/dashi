package bookmark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	_ "embed"

	"github.com/andresbott/dashi/internal/widgets"
)

// sanitizeURL rejects dangerous URI schemes (javascript:, data:, vbscript:).
// Only http://, https://, protocol-relative (//), and relative URLs are allowed.
func sanitizeURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "#"
	}
	if strings.HasPrefix(trimmed, "//") {
		return trimmed
	}
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return trimmed
	}
	// Allow relative URLs (no scheme)
	if !strings.Contains(strings.SplitN(lower, "/", 2)[0], ":") {
		return trimmed
	}
	return "#"
}

//go:embed bookmark.html
var bookmarkHTML string

var tmpl = template.Must(template.New("bookmark").Parse(bookmarkHTML))

type bookmarkConfig struct {
	URL       string `json:"url"`
	Icon      string `json:"icon"`
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle"`
	TextBelow bool   `json:"textBelow,omitempty"`
}

type bookmarkData struct {
	URL      string
	Title    string
	Subtitle string
}

// NewStaticRenderer returns a StaticRenderer for bookmark widgets.
func NewStaticRenderer() func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
	return func(config json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
		var cfg bookmarkConfig
		if len(config) > 0 {
			if err := json.Unmarshal(config, &cfg); err != nil {
				return "", fmt.Errorf("bookmark config: %w", err)
			}
		}

		data := bookmarkData{
			URL:      sanitizeURL(cfg.URL),
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
