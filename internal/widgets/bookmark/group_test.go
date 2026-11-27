package bookmark

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
	"github.com/andresbott/dashi/internal/widgetthemes"
	"github.com/gorilla/mux"
)

const twoBookmarks = `{"bookmarks":[
	{"url":"https://github.com","icon":"ti-brand-github","title":"GitHub","subtitle":"Code hosting"},
	{"url":"/notes","title":"Notes"}
]`

func renderGroupBrowser(t *testing.T, m *GroupModule, config string, ctx widgets.RenderContext) string {
	t.Helper()
	out, err := m.RenderBrowser(json.RawMessage(config), ctx)
	if err != nil {
		t.Fatalf("RenderBrowser: %v", err)
	}
	return string(out)
}

func TestGroupModule_Type(t *testing.T) {
	if got := NewGroupModule(nil).Type(); got != "bookmarkgroup" {
		t.Errorf("Type() = %q, want %q", got, "bookmarkgroup")
	}
}

func TestGroupModule_NoopHooks(t *testing.T) {
	m := NewGroupModule(nil)
	m.RegisterRoutes(mux.NewRouter())
	m.Warmup(context.Background(), nil)
}

func TestGroupRenderBrowser_OneLinkPerBookmark(t *testing.T) {
	html := renderGroupBrowser(t, NewGroupModule(nil), twoBookmarks+`}`, widgets.RenderContext{})

	if !strings.HasPrefix(strings.TrimSpace(html), `<div class="dashi-bookmarkgroup `) {
		t.Errorf("the group must be rooted at .dashi-bookmarkgroup:\n%s", html)
	}
	if got := strings.Count(html, `<a `); got != 2 {
		t.Errorf("want 2 links, got %d:\n%s", got, html)
	}
	for _, want := range []string{`href="https://github.com"`, `href="/notes"`, "GitHub", "Code hosting", "Notes"} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q in:\n%s", want, html)
		}
	}
	if strings.Index(html, "GitHub") > strings.Index(html, "Notes") {
		t.Errorf("bookmarks must render in stored order:\n%s", html)
	}
}

func TestGroupRenderBrowser_Direction(t *testing.T) {
	m := NewGroupModule(nil)
	tests := []struct {
		name      string
		direction string
		want      []string
		notWant   []string
	}{
		{name: "default is horizontal", direction: "", want: []string{`data-direction="horizontal"`, "flex-row", "flex-wrap"}},
		{name: "horizontal", direction: "horizontal", want: []string{`data-direction="horizontal"`, "flex-row", "flex-wrap"}},
		{name: "vertical", direction: "vertical", want: []string{`data-direction="vertical"`, "flex-col"}, notWant: []string{"flex-wrap"}},
		{name: "unknown falls back to horizontal", direction: "diagonal", want: []string{`data-direction="horizontal"`}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			html := renderGroupBrowser(t, m, twoBookmarks+`,"direction":"`+tc.direction+`"}`, widgets.RenderContext{})
			root, _, _ := strings.Cut(html, "<a ")
			for _, w := range tc.want {
				if !strings.Contains(root, w) {
					t.Errorf("missing %q on the group root:\n%s", w, root)
				}
			}
			for _, nw := range tc.notWant {
				if strings.Contains(root, nw) {
					t.Errorf("unexpected %q on the group root:\n%s", nw, root)
				}
			}
		})
	}
}

// textBelow and newTab are set once on the group and apply to every link.
func TestGroupRenderBrowser_SharedOptionsApplyToEveryLink(t *testing.T) {
	m := NewGroupModule(nil)

	plain := renderGroupBrowser(t, m, twoBookmarks+`}`, widgets.RenderContext{})
	if strings.Contains(plain, "data-text-below") || strings.Contains(plain, "target=") {
		t.Errorf("defaults must be text beside the icon, opening in place:\n%s", plain)
	}

	shared := renderGroupBrowser(t, m, twoBookmarks+`,"textBelow":true,"newTab":true}`, widgets.RenderContext{})
	if got := strings.Count(shared, "data-text-below"); got != 2 {
		t.Errorf("textBelow: want it on both links, got %d:\n%s", got, shared)
	}
	if got := strings.Count(shared, `target="_blank" rel="noopener noreferrer"`); got != 2 {
		t.Errorf("newTab: want it on both links, got %d:\n%s", got, shared)
	}
}

// The group carries the negative margin once (links inside it don't), and
// text-below links in a row get a fixed tile width so they line up.
func TestGroupRenderBrowser_LinkSpacing(t *testing.T) {
	m := NewGroupModule(nil)
	links := func(html string) string {
		_, rest, _ := strings.Cut(html, "<a ")
		return rest
	}

	row := renderGroupBrowser(t, m, twoBookmarks+`,"textBelow":true}`, widgets.RenderContext{})
	if strings.Contains(links(row), "-mx-2") {
		t.Errorf("links in a group must not carry their own negative margin:\n%s", row)
	}
	if got := strings.Count(row, " w-24"); got != 2 {
		t.Errorf("text-below links in a row want a tile width, got %d:\n%s", got, row)
	}

	for _, cfg := range []string{twoBookmarks + `}`, twoBookmarks + `,"textBelow":true,"direction":"vertical"}`} {
		if html := renderGroupBrowser(t, m, cfg, widgets.RenderContext{}); strings.Contains(html, "w-24") {
			t.Errorf("config %s: only text-below links in a row are tiles:\n%s", cfg, html)
		}
	}

	single := renderBrowser(t, NewModule(nil), `{"url":"https://x.com","title":"X","textBelow":true}`, widgets.RenderContext{})
	if !strings.Contains(single, "-mx-2") || strings.Contains(single, "w-24") {
		t.Errorf("a single bookmark keeps its layout:\n%s", single)
	}
}

func TestGroupRenderBrowser_ResolvesEachIcon(t *testing.T) {
	m := NewGroupModule(widgetthemes.NewStore(""))
	cfg := `{"bookmarks":[
		{"url":"https://a.com","title":"A","icon":"ti-brand-github"},
		{"url":"https://b.com","title":"B","icon":"selfhst:galaxus.webp"},
		{"url":"https://c.com","title":"C"}
	]}`
	html := renderGroupBrowser(t, m, cfg, widgets.RenderContext{WidgetTheme: "default"})
	if !strings.Contains(html, "") {
		t.Errorf("missing the Tabler glyph:\n%s", html)
	}
	if !strings.Contains(html, `src="https://cdn.jsdelivr.net/gh/selfhst/icons@main/webp/galaxus.webp"`) {
		t.Errorf("missing the selfh.st image:\n%s", html)
	}
	if got := strings.Count(html, "dashi-bookmark__icon"); got != 2 {
		t.Errorf("want 2 icons (the third bookmark has none), got %d:\n%s", got, html)
	}
}

func TestGroupRenderBrowser_UnsafeURLRendersInert(t *testing.T) {
	html := renderGroupBrowser(t, NewGroupModule(nil), `{"bookmarks":[{"url":"javascript:alert(1)","title":"x"},{"url":"https://ok.com","title":"y"}]}`, widgets.RenderContext{})
	if !strings.Contains(html, `href="#"`) || strings.Contains(html, "javascript:") {
		t.Errorf("an unsafe URL must render as href=\"#\":\n%s", html)
	}
	if !strings.Contains(html, `href="https://ok.com"`) {
		t.Errorf("the safe sibling must keep its URL:\n%s", html)
	}
}

func TestGroupRenderBrowser_Empty(t *testing.T) {
	for _, cfg := range []string{``, `{}`, `{"bookmarks":[]}`} {
		html := renderGroupBrowser(t, NewGroupModule(nil), cfg, widgets.RenderContext{})
		if strings.Contains(html, "<a ") {
			t.Errorf("config %q: want no links, got:\n%s", cfg, html)
		}
	}
}

func TestGroupRenderBrowser_InvalidJSON(t *testing.T) {
	if _, err := NewGroupModule(nil).RenderBrowser(json.RawMessage(`{bad`), widgets.RenderContext{}); err == nil {
		t.Error("want an error for invalid JSON")
	}
}

func TestGroupRenderStatic(t *testing.T) {
	render := NewGroupModule(nil).Renderer()
	got, err := render(json.RawMessage(`{"bookmarks":[{"url":"https://a.com","title":"A","subtitle":"first"},{"url":"//evil.com","title":"B"}],"direction":"vertical"}`), widgets.RenderContext{})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	html := string(got)
	for _, want := range []string{`href="https://a.com"`, "first", `href="#"`, ">B<", "flex-direction: column"} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q in:\n%s", want, html)
		}
	}
	if strings.Contains(html, "evil.com") {
		t.Errorf("unsafe URL leaked into the image render:\n%s", html)
	}

	horizontal, err := render(json.RawMessage(twoBookmarks+`}`), widgets.RenderContext{})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if !strings.Contains(string(horizontal), "flex-wrap: wrap") {
		t.Errorf("horizontal image render must wrap:\n%s", horizontal)
	}

	if _, err := render(json.RawMessage(`{bad`), widgets.RenderContext{}); err == nil {
		t.Error("want an error for invalid JSON")
	}
}

func TestGroupModule_ValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{name: "valid", config: twoBookmarks + `}`},
		{name: "empty config", config: ``},
		{name: "no bookmarks", config: `{"bookmarks":[]}`},
		{name: "empty url", config: `{"bookmarks":[{"url":"","title":"x"}]}`},
		{name: "javascript url", config: `{"bookmarks":[{"url":"https://ok.com"},{"url":"javascript:alert(1)"}]}`, wantErr: "bookmark 2 url must start with"},
		{name: "wrong type", config: `{"bookmarks":{}}`, wantErr: "bookmarkgroup config"},
	}
	m := NewGroupModule(nil)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := m.ValidateConfig(json.RawMessage(tc.config))
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err = %v, want containing %q", err, tc.wantErr)
			}
		})
	}
}

var (
	_ widgets.Module            = (*GroupModule)(nil)
	_ widgets.ConfigValidator   = (*GroupModule)(nil)
	_ widgets.BrowserRenderable = (*GroupModule)(nil)
)
