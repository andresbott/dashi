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
	reg.Register("test", func(config json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
		return template.HTML("<p>test-widget</p>"), nil
	})

	renderer := NewRenderer(reg)

	data := RenderData{
		Name:     "Test Dashboard",
		MaxWidth: "800px",
		HAlign:   "center",
		VAlign:   "top",
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
	err := renderer.Render(&buf, data)
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
	if strings.Contains(html, "widget-placeholder") {
		t.Error("unknown widget should render nothing, not a placeholder")
	}
	if !strings.Contains(html, "width: 50.0000%") {
		t.Error("expected percentage width for span-6 widget")
	}
	if !strings.Contains(html, "max-width: 800px") {
		t.Error("expected container max-width in CSS")
	}
}

func TestRenderer_Render_EmptyDashboard(t *testing.T) {
	reg := widgets.NewRegistry()
	renderer := NewRenderer(reg)

	data := RenderData{
		Name:     "Empty",
		MaxWidth: "100%",
		HAlign:   "center",
		VAlign:   "top",
		Rows:     []dashboard.Row{},
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, data)
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

func TestRenderer_Render_DarkMode(t *testing.T) {
	reg := widgets.NewRegistry()
	renderer := NewRenderer(reg)

	data := RenderData{
		Name:      "Dark Dashboard",
		ColorMode: "dark",
		Rows:      []dashboard.Row{},
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("expected DOCTYPE")
	}
}

func TestRenderer_Render_DebugMode(t *testing.T) {
	reg := widgets.NewRegistry()
	reg.Register("test", func(config json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
		return template.HTML("<p>widget</p>"), nil
	})

	renderer := NewRenderer(reg)

	data := RenderData{
		Name: "Debug Dashboard",
		QueryParams: map[string]string{
			"debug": "1",
		},
		Rows: []dashboard.Row{
			{
				ID: "row-1",
				Widgets: []dashboard.Widget{
					{ID: "w1", Type: "test", Width: 6},
					{ID: "w2", Type: "test", Width: 6},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "background: #ffcccc") {
		t.Error("expected first debug color")
	}
	if !strings.Contains(html, "background: #ccffcc") {
		t.Error("expected second debug color")
	}
}

func TestRenderer_Render_EmptyRowSkipped(t *testing.T) {
	reg := widgets.NewRegistry()
	renderer := NewRenderer(reg)

	data := RenderData{
		Name: "Dashboard with empty row",
		Rows: []dashboard.Row{
			{
				ID:      "empty-row",
				Height:  "auto",
				Widgets: []dashboard.Widget{},
			},
		},
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("expected DOCTYPE")
	}
}

func TestRenderer_Render_ExplicitHeight(t *testing.T) {
	reg := widgets.NewRegistry()
	reg.Register("test", func(config json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
		return template.HTML("<p>widget</p>"), nil
	})

	renderer := NewRenderer(reg)

	data := RenderData{
		Name: "Dashboard with explicit height",
		Rows: []dashboard.Row{
			{
				ID:     "row-1",
				Height: "200px",
				Widgets: []dashboard.Widget{
					{ID: "w1", Type: "test", Width: 12},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "200px") {
		t.Error("expected explicit height in HTML")
	}
}

func TestRenderer_Render_WidgetWidthDefault(t *testing.T) {
	reg := widgets.NewRegistry()
	reg.Register("test", func(config json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
		return template.HTML("<p>widget</p>"), nil
	})

	renderer := NewRenderer(reg)

	data := RenderData{
		Name: "Dashboard with default widget width",
		Rows: []dashboard.Row{
			{
				ID: "row-1",
				Widgets: []dashboard.Widget{
					{ID: "w1", Type: "test", Width: 0},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "width: 100.0000%") {
		t.Error("expected 100% width for default width 12")
	}
}

func TestRenderer_Render_ColumnProducesLeadingMargin(t *testing.T) {
	reg := widgets.NewRegistry()
	reg.Register("test", func(config json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
		return template.HTML("<p>widget</p>"), nil
	})

	renderer := NewRenderer(reg)

	data := RenderData{
		Name: "Positioned widget",
		Rows: []dashboard.Row{
			{
				ID: "row-1",
				Widgets: []dashboard.Widget{
					{ID: "w1", Type: "test", Width: 6, Column: 4},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := renderer.Render(&buf, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := buf.String()
	// Column 4 leaves 3 empty columns (3/12 = 25%) before a span-6 (50%) widget.
	if !strings.Contains(html, "margin-left: 25.0000%") {
		t.Errorf("expected 25%% leading margin for a column-4 widget, got: %s", html)
	}
	if !strings.Contains(html, "width: 50.0000%") {
		t.Error("expected the span-6 width to be preserved alongside the margin")
	}
}

func TestRenderer_Render_NoColumnHasZeroMargin(t *testing.T) {
	reg := widgets.NewRegistry()
	reg.Register("test", func(config json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
		return template.HTML("<p>widget</p>"), nil
	})

	var buf bytes.Buffer
	err := NewRenderer(reg).Render(&buf, RenderData{
		Name: "Flowing widget",
		Rows: []dashboard.Row{{
			ID:      "row-1",
			Widgets: []dashboard.Widget{{ID: "w1", Type: "test", Width: 6}},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "margin-left: 0.0000%") {
		t.Errorf("expected a zero leading margin for a widget without a column, got: %s", buf.String())
	}
}

func TestRenderer_Render_HAlignVariants(t *testing.T) {
	reg := widgets.NewRegistry()
	renderer := NewRenderer(reg)

	tests := []struct {
		name   string
		hAlign string
	}{
		{"left", "left"},
		{"right", "right"},
		{"center", "center"},
		{"default", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := RenderData{
				Name:   "Test",
				HAlign: tt.hAlign,
				Rows:   []dashboard.Row{},
			}

			var buf bytes.Buffer
			err := renderer.Render(&buf, data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			html := buf.String()
			if !strings.Contains(html, "<!DOCTYPE html>") {
				t.Error("expected valid HTML output")
			}
		})
	}
}

func TestRenderer_Render_VAlignVariants(t *testing.T) {
	reg := widgets.NewRegistry()
	renderer := NewRenderer(reg)

	tests := []struct {
		name   string
		vAlign string
	}{
		{"top", "top"},
		{"bottom", "bottom"},
		{"center", "center"},
		{"default", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := RenderData{
				Name:   "Test",
				VAlign: tt.vAlign,
				Rows:   []dashboard.Row{},
			}

			var buf bytes.Buffer
			err := renderer.Render(&buf, data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			html := buf.String()
			if !strings.Contains(html, "<!DOCTYPE html>") {
				t.Error("expected valid HTML output")
			}
		})
	}
}

func TestRenderer_Render_CustomCSS(t *testing.T) {
	reg := widgets.NewRegistry()
	renderer := NewRenderer(reg)

	data := RenderData{
		Name:          "Custom CSS Dashboard",
		CustomCSS:     ".custom { color: red; }",
		BackgroundCSS: "background: linear-gradient(to right, blue, green);",
		FontFamily:    "Arial, sans-serif",
		Rows:          []dashboard.Row{},
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, ".custom { color: red; }") {
		t.Error("expected custom CSS")
	}
	if !strings.Contains(html, "background: linear-gradient(to right, blue, green);") {
		t.Error("expected background CSS")
	}
	if !strings.Contains(html, "Arial, sans-serif") {
		t.Error("expected font family")
	}
}

func TestRenderer_Render_RenderContext(t *testing.T) {
	var capturedContext widgets.RenderContext
	reg := widgets.NewRegistry()
	reg.Register("test", func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
		capturedContext = ctx
		return template.HTML("<p>widget</p>"), nil
	})

	renderer := NewRenderer(reg)

	data := RenderData{
		Name:  "Dashboard",
		Theme: "dark",
		QueryParams: map[string]string{
			"foo": "bar",
		},
		PageIndex:  2,
		TotalPages: 5,
		Rows: []dashboard.Row{
			{
				ID: "row-1",
				Widgets: []dashboard.Widget{
					{ID: "w1", Type: "test", Width: 12},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := renderer.Render(&buf, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedContext.Theme != "dark" {
		t.Errorf("expected theme 'dark', got %s", capturedContext.Theme)
	}
	if capturedContext.QueryParams["foo"] != "bar" {
		t.Errorf("expected query param foo=bar, got %v", capturedContext.QueryParams)
	}
	if capturedContext.PageIndex != 2 {
		t.Errorf("expected page index 2, got %d", capturedContext.PageIndex)
	}
	if capturedContext.TotalPages != 5 {
		t.Errorf("expected total pages 5, got %d", capturedContext.TotalPages)
	}
}

func TestMasterTemplateHasNoWidgetSpecificCSS(t *testing.T) {
	// Widgets own their own styling (spec decision D2/D6); the page shell
	// must not know about individual widget types.
	for _, selector := range []string{
		".widget-clock", ".widget-bookmark", ".widget-page-indicator",
	} {
		if strings.Contains(masterHTML, selector) {
			t.Errorf("master.html still styles %s — move it into the widget's image template", selector)
		}
	}
}

func TestRenderer_Render_CustomCSSComesAfterWidgetContent(t *testing.T) {
	// litehtml collects <style> elements in document order and breaks
	// equal-specificity ties by insertion order. Widget CSS now ships inside
	// each widget's body, so custom.css must be the last <style> in the
	// document or a dashboard's overrides silently lose.
	reg := widgets.NewRegistry()
	reg.Register("styled", func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
		return template.HTML(`<style>.widget-styled .label{color:#000}</style><div class="widget-styled"><span class="label">x</span></div>`), nil
	})

	var buf bytes.Buffer
	err := NewRenderer(reg).Render(&buf, RenderData{
		Name:      "Cascade",
		CustomCSS: ".widget-styled .label{color:#f00}",
		Rows: []dashboard.Row{{
			ID:      "row-1",
			Widgets: []dashboard.Widget{{ID: "w1", Type: "styled", Width: 12}},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	html := buf.String()

	widgetCSS := strings.Index(html, ".widget-styled .label{color:#000}")
	customCSS := strings.Index(html, ".widget-styled .label{color:#f00}")
	if widgetCSS < 0 || customCSS < 0 {
		t.Fatalf("expected both stylesheets in output:\n%s", html)
	}
	if customCSS < widgetCSS {
		t.Errorf("custom.css must be inserted after the widget's own CSS to win the cascade:\n%s", html)
	}
	if bodyEnd := strings.Index(html, "</body>"); customCSS > bodyEnd {
		t.Errorf("custom.css must sit inside <body>, not after it:\n%s", html)
	}
}

func TestRenderer_Render_CustomCSSCannotBreakOutOfStyleElement(t *testing.T) {
	var buf bytes.Buffer
	err := NewRenderer(widgets.NewRegistry()).Render(&buf, RenderData{
		Name:      "XSS",
		CustomCSS: "</style><script>alert(1)</script>",
		Rows:      []dashboard.Row{},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	html := buf.String()

	if strings.Contains(html, "</style><script") {
		t.Errorf("custom.css broke out of the style element:\n%s", html)
	}
	payload := strings.Index(html, "alert(1)")
	closing := strings.Index(html[payload:], "</style>")
	if payload < 0 || closing < 0 {
		t.Errorf("payload must remain inside the style element:\n%s", html)
	}
}

func TestMasterTemplateUsesPaletteNotHardcodedColors(t *testing.T) {
	if strings.Contains(masterHTML, "#1a1a2e") || strings.Contains(masterHTML, "#e0e0e0") {
		t.Error("master.html should take colours from .Palette, not hardcode them")
	}
	if !strings.Contains(masterHTML, ".Palette.") {
		t.Error("master.html should reference palette fields")
	}
}
