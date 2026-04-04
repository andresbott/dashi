package static

import (
	"bytes"
	"encoding/json"
	"html/template"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/widgets"
)

func TestRenderer_Render(t *testing.T) {
	reg := widgets.NewRegistry()
	reg.Register("test", func(config json.RawMessage) (template.HTML, error) {
		return template.HTML("<p>test-widget</p>"), nil
	})

	renderer := NewRenderer(reg)

	dash := dashboard.Dashboard{
		Name: "Test Dashboard",
		Container: dashboard.Container{
			MaxWidth:        "800px",
			VerticalAlign:   "top",
			HorizontalAlign: "center",
		},
		Rows: []dashboard.Row{
			{
				ID:     "row-1",
				Title:  "Section One",
				Height: "auto",
				Width:  "100%",
				Widgets: []dashboard.Widget{
					{
						ID:     "w1",
						Type:   "test",
						Title:  "Widget 1",
						Width:  6,
						Config: json.RawMessage(`{}`),
					},
					{
						ID:    "w2",
						Type:  "unknown",
						Title: "Widget 2",
						Width: 6,
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, dash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := buf.String()

	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("expected DOCTYPE")
	}
	if !strings.Contains(html, "<title>Test Dashboard</title>") {
		t.Errorf("expected title, got: %s", html)
	}
	if !strings.Contains(html, "Section One") {
		t.Error("expected row title")
	}
	if !strings.Contains(html, "<p>test-widget</p>") {
		t.Error("expected rendered test widget")
	}
	if !strings.Contains(html, "widget-placeholder") {
		t.Error("expected placeholder for unknown widget")
	}
	if !strings.Contains(html, "grid-column: span 6") {
		t.Error("expected grid-column span")
	}
	if !strings.Contains(html, "max-width: 800px") {
		t.Error("expected container max-width in CSS")
	}
}

func TestRenderer_Render_EmptyDashboard(t *testing.T) {
	reg := widgets.NewRegistry()
	renderer := NewRenderer(reg)

	dash := dashboard.Dashboard{
		Name: "Empty",
		Container: dashboard.Container{
			MaxWidth:        "100%",
			VerticalAlign:   "top",
			HorizontalAlign: "center",
		},
		Rows: []dashboard.Row{},
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, dash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "<title>Empty</title>") {
		t.Error("expected title")
	}
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("expected DOCTYPE")
	}
}
