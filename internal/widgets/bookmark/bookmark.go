package bookmark

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"strings"

	_ "embed"

	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/widgets"
)

// errUnsafeURL is returned by validateURL for any URL outside the allowlist.
var errUnsafeURL = errors.New("url must start with http://, https:// or / (a path on this server)")

// validateURL is the bookmark URL rule, shared by the save-time validator
// and the renderer. The editor form (webui/src/widgets/bookmark/url.ts)
// implements the same rule; testdata/url_cases.json keeps the two in step.
//
// Allowed: empty (an unconfigured bookmark), or a URL that starts —
// case-insensitively, with no leading whitespace — with http:// or https://,
// or with a single / followed by anything but another / or \ (a path on
// this server). Everything else is rejected: other schemes (javascript:,
// data:, vbscript:, ...), protocol-relative //host and /\host (browsers
// treat \ as / in http URLs), and scheme-less text like example.com. ASCII
// control characters are rejected anywhere, because browsers strip tabs and
// newlines from URLs, which would turn "/\t/host" into "//host".
func validateURL(raw string) error {
	if raw == "" {
		return nil
	}
	for _, c := range raw {
		if c < 0x20 || c == 0x7f {
			return errUnsafeURL
		}
	}
	lower := strings.ToLower(raw)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return nil
	}
	if strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "//") && !strings.HasPrefix(raw, `/\`) {
		return nil
	}
	return errUnsafeURL
}

// sanitizeURL returns the URL to put in the link's href: the stored URL when
// it passes validateURL, otherwise the inert "#". Save-time validation
// keeps new bad URLs out, but dashboards stored before it (or edited on
// disk) can still carry one. html/template's own URL filter would also turn
// a javascript: href into "#ZgotmplZ"; this keeps the stricter rule
// (no //host, no scheme-less text) in the rendered output too.
func sanitizeURL(raw string) string {
	if raw == "" || validateURL(raw) != nil {
		return "#"
	}
	return raw
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
	Palette  themes.Palette
}

// NewStaticRenderer returns a StaticRenderer for bookmark widgets.
func NewStaticRenderer() func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
	return func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
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
			Palette:  ctx.EffectivePalette(),
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("bookmark render: %w", err)
		}
		return template.HTML(buf.String()), nil
	}
}
