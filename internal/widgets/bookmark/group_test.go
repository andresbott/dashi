package bookmark

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/data"
	bookmarkfiles "github.com/andresbott/dashi/internal/data/bookmarks"
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
	if got := NewGroupModule(nil, nil).Type(); got != "bookmarkgroup" {
		t.Errorf("Type() = %q, want %q", got, "bookmarkgroup")
	}
}

func TestGroupModule_NoopHooks(t *testing.T) {
	m := NewGroupModule(nil, nil)
	m.RegisterRoutes(mux.NewRouter())
	m.Warmup(context.Background(), nil)
}

func TestGroupRenderBrowser_OneLinkPerBookmark(t *testing.T) {
	html := renderGroupBrowser(t, NewGroupModule(nil, nil), twoBookmarks+`}`, widgets.RenderContext{})

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
	m := NewGroupModule(nil, nil)
	tests := []struct {
		name      string
		direction string
		want      []string
		notWant   []string
	}{
		{name: "default is horizontal", direction: "", want: []string{`data-direction="horizontal"`, "auto-fit"}},
		{name: "horizontal", direction: "horizontal", want: []string{`data-direction="horizontal"`, "auto-fit"}},
		{name: "vertical", direction: "vertical", want: []string{`data-direction="vertical"`, "grid-template-columns: minmax(0, 1fr);"}, notWant: []string{"auto-fit"}},
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
	m := NewGroupModule(nil, nil)

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

// The group carries the negative margin once (links inside it don't); apart
// from that, every link is the link a single bookmark renders, with no
// group-only sizing.
func TestGroupRenderBrowser_LinksMatchSingleBookmark(t *testing.T) {
	m := NewGroupModule(nil, nil)
	firstLink := func(html string) string {
		_, rest, _ := strings.Cut(html, "<a ")
		link, _, _ := strings.Cut(rest, ">")
		return link
	}

	for _, textBelow := range []string{"false", "true"} {
		group := renderGroupBrowser(t, m, `{"bookmarks":[{"url":"https://x.com","title":"X"}],"textBelow":`+textBelow+`}`, widgets.RenderContext{})
		single := renderBrowser(t, NewModule(nil), `{"url":"https://x.com","title":"X","textBelow":`+textBelow+`}`, widgets.RenderContext{})
		want := strings.Replace(firstLink(single), "-mx-2 ", "", 1)
		if got := firstLink(group); got != want {
			t.Errorf("textBelow=%s: group link differs from a single bookmark\n got: %s\nwant: %s", textBelow, got, want)
		}
	}
}

func TestGroupRenderBrowser_Columns(t *testing.T) {
	m := NewGroupModule(nil, nil)
	tests := []struct {
		name    string
		extra   string
		want    string
		notWant string
	}{
		{name: "default fits to width", extra: ``, want: "repeat(auto-fit, minmax(min(100%, 10rem), 1fr))", notWant: "data-columns"},
		{name: "fixed row length", extra: `,"columns":3`, want: `repeat(3, minmax(0, 1fr))`},
		{name: "fixed row length is marked", extra: `,"columns":3`, want: `data-columns="3"`},
		{name: "capped at the maximum", extra: `,"columns":40`, want: `repeat(12, minmax(0, 1fr))`},
		{name: "negative fits to width", extra: `,"columns":-2`, want: "auto-fit", notWant: "data-columns"},
		{name: "vertical ignores columns", extra: `,"columns":3,"direction":"vertical"`, want: "grid-template-columns: minmax(0, 1fr);", notWant: "repeat("},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			html := renderGroupBrowser(t, m, twoBookmarks+tc.extra+`}`, widgets.RenderContext{})
			root, _, _ := strings.Cut(html, "<a ")
			if !strings.Contains(root, tc.want) {
				t.Errorf("missing %q on the group root:\n%s", tc.want, root)
			}
			if tc.notWant != "" && strings.Contains(root, tc.notWant) {
				t.Errorf("unexpected %q on the group root:\n%s", tc.notWant, root)
			}
		})
	}
}

func TestGroupRenderBrowser_ResolvesEachIcon(t *testing.T) {
	m := NewGroupModule(widgetthemes.NewStore(""), nil)
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
	html := renderGroupBrowser(t, NewGroupModule(nil, nil), `{"bookmarks":[{"url":"javascript:alert(1)","title":"x"},{"url":"https://ok.com","title":"y"}]}`, widgets.RenderContext{})
	if !strings.Contains(html, `href="#"`) || strings.Contains(html, "javascript:") {
		t.Errorf("an unsafe URL must render as href=\"#\":\n%s", html)
	}
	if !strings.Contains(html, `href="https://ok.com"`) {
		t.Errorf("the safe sibling must keep its URL:\n%s", html)
	}
}

func TestGroupRenderBrowser_Empty(t *testing.T) {
	for _, cfg := range []string{``, `{}`, `{"bookmarks":[]}`} {
		html := renderGroupBrowser(t, NewGroupModule(nil, nil), cfg, widgets.RenderContext{})
		if strings.Contains(html, "<a ") {
			t.Errorf("config %q: want no links, got:\n%s", cfg, html)
		}
	}
}

func TestGroupRenderBrowser_InvalidJSON(t *testing.T) {
	if _, err := NewGroupModule(nil, nil).RenderBrowser(json.RawMessage(`{bad`), widgets.RenderContext{}); err == nil {
		t.Error("want an error for invalid JSON")
	}
}

func TestGroupRenderStatic(t *testing.T) {
	render := NewGroupModule(nil, nil).Renderer()
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
	if strings.Contains(string(horizontal), "width:") {
		t.Errorf("a fit-to-width row must not size its links:\n%s", horizontal)
	}

	fixed, err := render(json.RawMessage(twoBookmarks+`,"columns":4}`), widgets.RenderContext{})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if got := strings.Count(string(fixed), "width: 25.0000%"); got != 2 {
		t.Errorf("columns=4: want each link at 25%% width, got %d:\n%s", got, fixed)
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
		{name: "columns in range", config: `{"bookmarks":[],"columns":12}`},
		{name: "negative columns", config: `{"bookmarks":[],"columns":-1}`, wantErr: "columns must be between 0 and 12"},
		{name: "too many columns", config: `{"bookmarks":[],"columns":13}`, wantErr: "columns must be between 0 and 12"},
		{name: "file source", config: `{"source":"file","file":"links.yaml"}`},
		{name: "dashboard source", config: `{"source":"dashboard"}`},
		{name: "unknown source", config: `{"source":"web"}`, wantErr: `source must be "dashboard" or "file"`},
	}
	m := NewGroupModule(nil, nil)
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

// fakeFiles is a FileSource keyed by file name; a missing name is
// data.ErrNotFound, an entry in errs returns that error.
type fakeFiles struct {
	files map[string][]bookmarkfiles.Item
	errs  map[string]error
}

func (f fakeFiles) Bookmarks(name string) ([]bookmarkfiles.Item, error) {
	if err, ok := f.errs[name]; ok {
		return nil, err
	}
	items, ok := f.files[name]
	if !ok {
		return nil, data.ErrNotFound
	}
	return items, nil
}

func TestGroupRender_FileSource(t *testing.T) {
	files := fakeFiles{
		files: map[string][]bookmarkfiles.Item{
			"links.yaml": {
				{URL: "https://file.example", Title: "From file", Subtitle: "sub"},
				{URL: "javascript:alert(1)", Title: "Unsafe"},
			},
		},
		errs: map[string]error{"broken.yaml": bookmarkfiles.ErrInvalidFile},
	}
	m := NewGroupModule(nil, files)
	// The dashboard's own list is kept but ignored while the source is a file.
	own := `"bookmarks":[{"url":"https://own.example","title":"Own"}]`

	tests := []struct {
		name    string
		config  string
		want    []string
		notWant []string
	}{
		{name: "dashboard source by default", config: `{` + own + `,"file":"links.yaml"}`, want: []string{"Own"}, notWant: []string{"From file"}},
		{name: "explicit dashboard source", config: `{` + own + `,"source":"dashboard","file":"links.yaml"}`, want: []string{"Own"}, notWant: []string{"From file"}},
		{name: "file source", config: `{` + own + `,"source":"file","file":"links.yaml"}`, want: []string{"From file", "sub", `href="https://file.example"`, `href="#"`}, notWant: []string{"Own", "javascript:"}},
		{name: "no file selected", config: `{` + own + `,"source":"file"}`, want: []string{"No bookmarks file selected"}, notWant: []string{"Own", "<a "}},
		{name: "missing file", config: `{"source":"file","file":"gone.yaml"}`, want: []string{"Bookmarks file &#34;gone.yaml&#34; not found"}, notWant: []string{"<a "}},
		{name: "broken file", config: `{"source":"file","file":"broken.yaml"}`, want: []string{"Bookmarks file &#34;broken.yaml&#34; could not be read"}, notWant: []string{"<a "}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			browser := renderGroupBrowser(t, m, tc.config, widgets.RenderContext{})
			image, err := m.Renderer()(json.RawMessage(tc.config), widgets.RenderContext{})
			if err != nil {
				t.Fatalf("image render: %v", err)
			}
			for stack, html := range map[string]string{"browser": browser, "image": string(image)} {
				for _, w := range tc.want {
					if !strings.Contains(html, w) {
						t.Errorf("%s: missing %q in:\n%s", stack, w, html)
					}
				}
				for _, nw := range tc.notWant {
					if strings.Contains(html, nw) {
						t.Errorf("%s: unexpected %q in:\n%s", stack, nw, html)
					}
				}
			}
		})
	}

	html := renderGroupBrowser(t, NewGroupModule(nil, nil), `{"source":"file","file":"links.yaml"}`, widgets.RenderContext{})
	if !strings.Contains(html, "Bookmarks files are not available") {
		t.Errorf("a module without a file source must say so:\n%s", html)
	}
}

var (
	_ FileSource = (*bookmarkfiles.Store)(nil)

	_ widgets.Module            = (*GroupModule)(nil)
	_ widgets.ConfigValidator   = (*GroupModule)(nil)
	_ widgets.BrowserRenderable = (*GroupModule)(nil)
)
