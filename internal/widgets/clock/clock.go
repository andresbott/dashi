package clock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"time"

	_ "embed"

	"github.com/andresbott/dashi/internal/widgets"
)

//go:embed clock.html
var clockHTML string

var tmpl = template.Must(template.New("clock").Parse(clockHTML))

type clockConfig struct {
	Hour12      bool   `json:"hour12"`
	ShowSeconds bool   `json:"showSeconds"`
	ShowDate    bool   `json:"showDate"`
	Font        string `json:"font"`
}

type clockData struct {
	Time     string
	Date     string
	ShowDate bool
	Font     string
}

// NewStaticRenderer returns a StaticRenderer that renders the current time.
// The nowFn parameter allows injecting a custom time source for testing.
// Pass nil to use time.Now.
func NewStaticRenderer(nowFn func() time.Time) func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
	if nowFn == nil {
		nowFn = time.Now
	}
	return func(config json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
		var cfg clockConfig
		if len(config) > 0 {
			if err := json.Unmarshal(config, &cfg); err != nil {
				return "", fmt.Errorf("clock config: %w", err)
			}
		}

		now := nowFn()
		timeStr := formatTime(now, cfg.Hour12, cfg.ShowSeconds)

		data := clockData{
			Time:     timeStr,
			ShowDate: cfg.ShowDate,
			Font:     cfg.Font,
		}
		if cfg.ShowDate {
			data.Date = now.Format("Monday, January 2, 2006")
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("clock render: %w", err)
		}
		return template.HTML(buf.String()), nil
	}
}

func formatTime(t time.Time, hour12, showSeconds bool) string {
	switch {
	case hour12 && showSeconds:
		return t.Format("3:04:05 PM")
	case hour12:
		return t.Format("3:04 PM")
	case showSeconds:
		return t.Format("15:04:05")
	default:
		return t.Format("15:04")
	}
}
