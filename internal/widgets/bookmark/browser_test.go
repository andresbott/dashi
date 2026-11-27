package bookmark

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
	"github.com/andresbott/dashi/internal/widgetthemes"
)

func renderBrowser(t *testing.T, m *Module, config string, ctx widgets.RenderContext) string {
	t.Helper()
	out, err := m.RenderBrowser(json.RawMessage(config), ctx)
	if err != nil {
		t.Fatalf("RenderBrowser: %v", err)
	}
	return string(out)
}

func TestRenderBrowser_Structure(t *testing.T) {
	m := NewModule(widgetthemes.NewStore(""))
	html := renderBrowser(t, m, `{"url":"https://github.com","title":"GitHub","subtitle":"Code hosting"}`, widgets.RenderContext{})

	if !strings.HasPrefix(strings.TrimSpace(html), `<a href="https://github.com" class="dashi-bookmark `) {
		t.Errorf("the whole widget must be one link rooted at .dashi-bookmark:\n%s", html)
	}
	if strings.Count(html, "<a ") != 1 {
		t.Errorf("want exactly one link:\n%s", html)
	}
	for _, want := range []string{"GitHub", "Code hosting", "font-semibold", "text-[var(--dashi-text-muted)]", "flex-row"} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q in:\n%s", want, html)
		}
	}
	if strings.Contains(html, "data-text-below") {
		t.Errorf("default layout must be horizontal:\n%s", html)
	}
}

func TestRenderBrowser_NoSubtitle(t *testing.T) {
	m := NewModule(nil)
	html := renderBrowser(t, m, `{"url":"https://example.com","title":"Example"}`, widgets.RenderContext{})
	if strings.Contains(html, "dashi-bookmark__subtitle") {
		t.Errorf("empty subtitle must not render a slot:\n%s", html)
	}
}

func TestRenderBrowser_TextBelow(t *testing.T) {
	m := NewModule(nil)
	html := renderBrowser(t, m, `{"url":"https://example.com","title":"Example","textBelow":true}`, widgets.RenderContext{})
	for _, want := range []string{"data-text-below", "flex-col", "text-center"} {
		if !strings.Contains(html, want) {
			t.Errorf("textBelow: missing %q in:\n%s", want, html)
		}
	}
	if strings.Contains(html, "flex-row") {
		t.Errorf("textBelow must not use the horizontal layout:\n%s", html)
	}
}

func TestRenderBrowser_UnsafeURLRendersInert(t *testing.T) {
	m := NewModule(nil)
	for _, raw := range []string{"javascript:alert(1)", "//evil.example", "example.com", "data:text/html,x"} {
		cfg, _ := json.Marshal(map[string]string{"url": raw, "title": "x"})
		html := renderBrowser(t, m, string(cfg), widgets.RenderContext{})
		if !strings.Contains(html, `href="#"`) {
			t.Errorf("url %q: want href=\"#\", got:\n%s", raw, html)
		}
	}
}

func TestRenderBrowser_FontIcon(t *testing.T) {
	// The embedded default theme maps clear-sky to Tabler's sun (U+EB30).
	m := NewModule(widgetthemes.NewStore(""))
	for _, icon := range []string{"clear-sky", "ti-clear-sky"} {
		cfg, _ := json.Marshal(map[string]string{"url": "https://example.com", "title": "x", "icon": icon})
		html := renderBrowser(t, m, string(cfg), widgets.RenderContext{WidgetTheme: "default"})
		if !strings.Contains(html, "") {
			t.Errorf("icon %q: missing code point U+EB30 in:\n%s", icon, html)
		}
		if !strings.Contains(html, "font-family: '"+widgetthemes.IconFontFamily("default")+"'") {
			t.Errorf("icon %q: missing icon font family in:\n%s", icon, html)
		}
	}
}

func TestRenderBrowser_ImageIcon(t *testing.T) {
	dir := t.TempDir()
	themeDir := filepath.Join(dir, "pics")
	iconsDir := filepath.Join(themeDir, "widgets", "weather", "icons")
	if err := os.MkdirAll(iconsDir, 0o750); err != nil {
		t.Fatal(err)
	}
	manifest := "name: \"pics\"\nicons:\n  type: image\n"
	if err := os.WriteFile(filepath.Join(themeDir, "widget_theme.yaml"), []byte(manifest), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(iconsDir, "github.svg"), []byte("<svg/>"), 0o600); err != nil {
		t.Fatal(err)
	}

	m := NewModule(widgetthemes.NewStore(dir))
	html := renderBrowser(t, m, `{"url":"https://github.com","title":"GitHub","icon":"github"}`, widgets.RenderContext{WidgetTheme: "pics"})

	if !strings.Contains(html, `src="/api/v0/widget-themes/pics/icons/github"`) {
		t.Errorf("image icon must be served through the widget-theme icon route:\n%s", html)
	}
	if strings.Contains(html, dir) {
		t.Errorf("the server filesystem path must never reach the browser:\n%s", html)
	}
}

func TestRenderBrowser_MissingIconRendersNone(t *testing.T) {
	cases := map[string]*Module{
		"unknown name": NewModule(widgetthemes.NewStore("")),
		"nil store":    NewModule(nil),
	}
	for name, m := range cases {
		for _, icon := range []string{"", "ti-definitely-not-an-icon"} {
			cfg, _ := json.Marshal(map[string]string{"url": "https://example.com", "title": "x", "icon": icon})
			html := renderBrowser(t, m, string(cfg), widgets.RenderContext{})
			if strings.Contains(html, "dashi-bookmark__icon") {
				t.Errorf("%s, icon %q: want no icon element, got:\n%s", name, icon, html)
			}
		}
	}
}

func TestRenderBrowser_InvalidJSON(t *testing.T) {
	m := NewModule(nil)
	if _, err := m.RenderBrowser(json.RawMessage(`{bad`), widgets.RenderContext{}); err == nil {
		t.Error("want an error for invalid JSON")
	}
}

func TestBrowserTemplateHasNoComputedClasses(t *testing.T) {
	// Tailwind extracts classes statically: an {{if}} choosing between
	// literal classes is fine, an interpolated value is not.
	classAttr := regexp.MustCompile(`class="[^"]*"`)
	branches := regexp.MustCompile(`\{\{if [^}]*\}\}|\{\{else\}\}|\{\{end\}\}`)
	for _, attr := range classAttr.FindAllString(browserHTML, -1) {
		if strings.Contains(branches.ReplaceAllString(attr, ""), "{{") {
			t.Errorf("class attribute interpolates data: %s", attr)
		}
	}
}

func TestBrowserShipsNoAssets(t *testing.T) {
	// Styling is pure Tailwind; add a scoping test if a bookmark.css appears.
	var m any = NewModule(nil)
	if _, ok := m.(widgets.BrowserAssets); ok {
		t.Error("bookmark ships no CSS/JS")
	}
}

func TestRenderBrowser_ImplementsBrowserRenderable(t *testing.T) {
	var _ widgets.BrowserRenderable = NewModule(nil)
}

// Every icon kind the editor's picker offers renders in the viewer: any
// Tabler name (not only the theme manifest's), selfh.st images, and the
// dashboard's own uploaded assets.
func TestRenderBrowser_EditorIconKinds(t *testing.T) {
	m := NewModule(widgetthemes.NewStore(""))
	ctx := widgets.RenderContext{WidgetTheme: "default", DashboardID: "abc123"}

	tests := []struct {
		name string
		icon string
		want string
	}{
		{name: "tabler", icon: "ti-brand-github", want: ""},
		{name: "selfhst webp", icon: "selfhst:galaxus.webp", want: `src="https://cdn.jsdelivr.net/gh/selfhst/icons@main/webp/galaxus.webp"`},
		{name: "selfhst png", icon: "selfhst:galaxus.png", want: `src="https://cdn.jsdelivr.net/gh/selfhst/icons@main/png/galaxus.png"`},
		{name: "dashboard asset", icon: "dashboard:logo one.png", want: `src="/api/v0/dashboards/abc123/assets/logo%20one.png"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			html := renderBrowser(t, m, `{"url":"https://x.com","title":"X","icon":"`+tc.icon+`"}`, ctx)
			if !strings.Contains(html, tc.want) {
				t.Errorf("want %q in:\n%s", tc.want, html)
			}
		})
	}

	for _, bad := range []string{"selfhst:", "selfhst:../x.png", "dashboard:"} {
		html := renderBrowser(t, m, `{"url":"https://x.com","title":"X","icon":"`+bad+`"}`, ctx)
		if strings.Contains(html, "dashi-bookmark__icon") {
			t.Errorf("icon %q: expected no icon element:\n%s", bad, html)
		}
	}
}

// Hover highlights the link with a translucent theme-coloured background;
// it must not fade the bookmark, which reads as disabled.
func TestRenderBrowser_HoverHighlights(t *testing.T) {
	html := renderBrowser(t, NewModule(nil), `{"url":"https://example.com","title":"Example"}`, widgets.RenderContext{})
	for _, want := range []string{"hover:bg-[var(--dashi-text)]/10", "focus-visible:bg-[var(--dashi-text)]/10", "rounded-lg"} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q in:\n%s", want, html)
		}
	}
	if strings.Contains(html, "hover:opacity") {
		t.Errorf("hover must not fade the bookmark:\n%s", html)
	}
}

// newTab picks where the link opens: off (the default, and what bookmarks
// saved before the option carry) opens in place, on opens a new tab.
func TestRenderBrowser_NewTab(t *testing.T) {
	m := NewModule(nil)

	inPlace := renderBrowser(t, m, `{"url":"https://example.com","title":"Example"}`, widgets.RenderContext{})
	if strings.Contains(inPlace, "target=") || strings.Contains(inPlace, "rel=") {
		t.Errorf("default must open in place:\n%s", inPlace)
	}

	newTab := renderBrowser(t, m, `{"url":"https://example.com","title":"Example","newTab":true}`, widgets.RenderContext{})
	if !strings.Contains(newTab, `target="_blank" rel="noopener noreferrer"`) {
		t.Errorf("newTab must open a new tab without opener access:\n%s", newTab)
	}
}
