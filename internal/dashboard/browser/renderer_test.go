package browser

import (
	"encoding/json"
	"html/template"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/widgets"
)

func testRegistry() *widgets.Registry {
	r := widgets.NewRegistry()
	r.RegisterBrowser("demo", func(cfg json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
		return template.HTML(`<div class="dashi-demo">demo</div>`), nil
	})
	return r
}

func baseData() RenderData {
	return RenderData{
		Name:        "My Dash",
		DashboardID: "abc123",
		ThemeID:     "th1234",
		WidgetTheme: "mono",
		ColorMode:   "dark",
		Rows: []dashboard.Row{{
			Title:   "Top",
			Widgets: []dashboard.Widget{{ID: "w1", Type: "demo", Width: 4}},
		}},
		TotalPages: 1,
	}
}

func render(t *testing.T, r *Renderer, data RenderData) string {
	t.Helper()
	var buf strings.Builder
	if err := r.Render(&buf, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	return buf.String()
}

func TestRenderIncludesShellAssets(t *testing.T) {
	out := render(t, NewRenderer(testRegistry(), nil), baseData())

	for _, want := range []string{
		`<!DOCTYPE html>`,
		`href="/_dashi/assets/viewer.css"`,
		`href="/_dashi/assets/widgets.css"`,
		`href="/_dashi/theme/th1234.css"`,
		`href="/_dashi/widget-theme/mono.css"`,
		`data-color-mode="dark"`,
		`<title>My Dash</title>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderPlacesWidgetHTMLInSizedCell(t *testing.T) {
	out := render(t, NewRenderer(testRegistry(), nil), baseData())

	if !strings.Contains(out, `<div class="dashi-demo">demo</div>`) {
		t.Errorf("widget HTML missing:\n%s", out)
	}
	// 4 of 12 columns.
	if !strings.Contains(out, `width: 33.3333%`) {
		t.Errorf("cell width missing:\n%s", out)
	}
	if !strings.Contains(out, `Top`) {
		t.Errorf("row title missing:\n%s", out)
	}
}

func TestRenderPositionsWidgetWithLeadingMargin(t *testing.T) {
	data := baseData()
	data.Rows = []dashboard.Row{{
		Title:   "Top",
		Widgets: []dashboard.Widget{{ID: "w1", Type: "demo", Width: 6, Column: 4}},
	}}
	out := render(t, NewRenderer(testRegistry(), nil), data)

	// Column 4 leaves 3 empty columns (3/12 = 25%) before a span-6 (50%) widget.
	if !strings.Contains(out, "margin-left: 25.0000%") {
		t.Errorf("expected 25%% leading margin for a column-4 widget:\n%s", out)
	}
	if !strings.Contains(out, "width: 50.0000%") {
		t.Errorf("expected the span-6 width to be preserved:\n%s", out)
	}
}

func TestRenderIncludesWidgetScriptsOnlyForTypesPresentThatShipJS(t *testing.T) {
	r := NewRenderer(testRegistry(), map[string]bool{"demo": true, "unused": true})
	out := render(t, r, baseData())

	if !strings.Contains(out, `src="/_dashi/widgets/demo.js"`) {
		t.Errorf("expected demo script tag:\n%s", out)
	}
	if strings.Contains(out, "unused.js") {
		t.Error("script tag emitted for a widget type not on the page")
	}
}

func TestRenderSkipsScriptForTypeWithoutJS(t *testing.T) {
	out := render(t, NewRenderer(testRegistry(), map[string]bool{}), baseData())
	if strings.Contains(out, "/_dashi/widgets/") {
		t.Errorf("no widget ships JS, so no script tags expected:\n%s", out)
	}
}

func TestRenderAppendsCustomCSSLast(t *testing.T) {
	data := baseData()
	data.CustomCSS = ".dashi-demo{color:red}"
	out := render(t, NewRenderer(testRegistry(), nil), data)

	custom := strings.Index(out, ".dashi-demo{color:red}")
	theme := strings.Index(out, `/_dashi/theme/th1234.css`)
	widgetTheme := strings.Index(out, `/_dashi/widget-theme/mono.css`)
	if custom < 0 {
		t.Fatalf("custom CSS missing:\n%s", out)
	}
	if custom < theme || custom < widgetTheme {
		t.Error("custom.css must come after both theme stylesheets so it wins")
	}
}

func TestRenderDefaultsEmptyThemeReferences(t *testing.T) {
	data := baseData()
	data.ThemeID = ""
	data.WidgetTheme = ""
	out := render(t, NewRenderer(testRegistry(), nil), data)

	for _, want := range []string{
		`href="/_dashi/theme/default.css"`,
		`href="/_dashi/widget-theme/default.css"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderNoInlineRootStyle(t *testing.T) {
	// Per-dashboard colour overrides are gone: every colour comes from the
	// linked theme stylesheet, so <html> carries no style attribute and the
	// page carries no background <style> block.
	out := render(t, NewRenderer(testRegistry(), nil), baseData())

	start := strings.Index(out, "<html")
	if start < 0 {
		t.Fatalf("no <html> tag:\n%s", out)
	}
	end := strings.Index(out[start:], ">")
	if tag := out[start : start+end]; strings.Contains(tag, "style=") {
		t.Errorf("<html> must carry no inline style, got %q", tag)
	}
	if strings.Contains(out, "--dashi-page-bg") {
		t.Errorf("page must not inline background CSS:\n%s", out)
	}
}

func TestRenderSkipsEmptyAutoHeightRows(t *testing.T) {
	data := baseData()
	data.Rows = append(data.Rows, dashboard.Row{Height: "auto"})
	out := render(t, NewRenderer(testRegistry(), nil), data)

	if got := strings.Count(out, `class="dashi-row`); got != 1 {
		t.Errorf("rendered %d rows, want 1 (empty auto-height row must be skipped)", got)
	}
}

func TestRenderEscapesDashboardName(t *testing.T) {
	data := baseData()
	data.Name = `<script>alert(1)</script>`
	out := render(t, NewRenderer(testRegistry(), nil), data)

	if strings.Contains(out, "<script>alert(1)</script>") {
		t.Errorf("dashboard name was not escaped:\n%s", out)
	}
}

func TestBodyBackgroundUsesTheShorthandNotAColorUtility(t *testing.T) {
	// --dashi-page-bg carries gradients and url(...) values, which are not
	// valid background-color. A `bg-[var(--dashi-page-bg)]` Tailwind utility
	// compiles to background-color and silently drops them, so the shell must
	// rely on the `background` shorthand declared in viewer.css's base layer.
	out := render(t, NewRenderer(testRegistry(), nil), baseData())
	if strings.Contains(out, "bg-[var(--dashi-page-bg") {
		t.Errorf("body must not use a bg-* color utility for the page background:\n%s", out)
	}

	css := string(viewerCSS)
	if !strings.Contains(css, "body{background:var(--dashi-page-bg,var(--dashi-bg))}") {
		t.Error("viewer.css must declare the body background as a shorthand; run make viewer-css")
	}
	if strings.Contains(css, "background-color:var(--dashi-page-bg") {
		t.Error("viewer.css still compiles the page background to background-color")
	}
}

func TestRenderNeutralisesStyleBreakoutInCustomCSS(t *testing.T) {
	data := baseData()
	data.CustomCSS = "</style><script>alert(1)</script>"
	out := render(t, NewRenderer(testRegistry(), nil), data)

	// The payload text may survive as inert CSS text; what must not survive
	// is a </style> that ends the element early and lets the rest be parsed
	// as HTML. The shell emits exactly one <style> element for custom.css,
	// so exactly one closing tag may appear, and the payload must sit before
	// it (i.e. still inside the element).
	if got := strings.Count(out, "</style>"); got != 1 {
		t.Errorf("expected exactly 1 </style> (the shell's own), got %d:\n%s", got, out)
	}
	if strings.Contains(out, "</style><script") {
		t.Errorf("custom.css broke out of the style element:\n%s", out)
	}
	payload := strings.Index(out, "alert(1)")
	closing := strings.Index(out, "</style>")
	if payload < 0 || closing < 0 || payload > closing {
		t.Errorf("payload must remain inside the style element (payload=%d close=%d):\n%s", payload, closing, out)
	}
}

func TestRenderOmitsPageNavForSinglePageDashboard(t *testing.T) {
	out := render(t, NewRenderer(testRegistry(), nil), baseData())
	if strings.Contains(out, "dashi-pages") {
		t.Errorf("single-page dashboard must not render page navigation:\n%s", out)
	}
}

func TestRenderPageNavLinksEveryPageAndMarksCurrent(t *testing.T) {
	data := baseData()
	data.TotalPages = 3
	data.PageIndex = 1
	data.PageNames = []string{"Overview", "", "Links"}
	out := render(t, NewRenderer(testRegistry(), nil), data)

	for _, want := range []string{
		`href="?page=0"`,
		`href="?page=1"`,
		`href="?page=2"`,
		`>Overview<`,
		`>Page 2<`, // unnamed page falls back to a 1-based label
		`>Links<`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("page nav missing %q:\n%s", want, out)
		}
	}
	if got := strings.Count(out, `aria-current="page"`); got != 1 {
		t.Errorf("exactly one page must be marked current, got %d:\n%s", got, out)
	}
	// The current page marker must sit on the link for PageIndex.
	current := strings.Index(out, `aria-current="page"`)
	if !strings.Contains(out[max(current-40, 0):current], `href="?page=1"`) {
		t.Errorf("aria-current is not on the current page's link:\n%s", out)
	}
	if strings.Contains(out, "<script") && !strings.Contains(out, "/_dashi/widgets/") {
		t.Error("page navigation must not need JavaScript")
	}
}
