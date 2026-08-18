package clock

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

//go:embed clock.js
var browserJS []byte

var browserTmpl = template.Must(template.New("clock-browser").Parse(browserHTML))

type browserData struct {
	Time        string
	Date        string
	Hour12      bool
	ShowSeconds bool
	ShowDate    bool
	Font        string
}

// RenderBrowser renders the clock for the browser stack. Values are
// prefilled from the server clock so the widget is correct before its JS
// runs; clock.js then keeps it ticking.
func (m *Module) RenderBrowser(config json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
	var cfg clockConfig
	if len(config) > 0 {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return "", fmt.Errorf("clock config: %w", err)
		}
	}

	now := m.now()
	data := browserData{
		Time:        formatTime(now, cfg.Hour12, cfg.ShowSeconds),
		Hour12:      cfg.Hour12,
		ShowSeconds: cfg.ShowSeconds,
		ShowDate:    cfg.ShowDate,
		Font:        cfg.Font,
	}
	if cfg.ShowDate {
		data.Date = now.Format("Monday, January 2, 2006")
	}

	var buf bytes.Buffer
	if err := browserTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("clock browser render: %w", err)
	}
	return template.HTML(buf.String()), nil
}

// CSS implements widgets.BrowserAssets.
func (m *Module) CSS() []byte { return nil }

// JS implements widgets.BrowserAssets.
func (m *Module) JS() []byte { return browserJS }
