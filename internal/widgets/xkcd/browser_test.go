package xkcd

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	xkcdclient "github.com/andresbott/dashi/internal/providers/xkcd"
	"github.com/andresbott/dashi/internal/widgets"
)

var sampleComic = map[string]any{
	"num": 3329, "title": "Lunar Ring", "safe_title": "Lunar Ring",
	"img": "https://imgs.xkcd.com/comics/lunar_ring.png",
	"alt": "Some alt <text> & more", "day": "1", "month": "11", "year": "2026",
}

func TestRenderBrowser_Structure(t *testing.T) {
	m := NewModule(newTestClient(t, sampleComic), slog.Default())

	out, err := m.RenderBrowser(json.RawMessage(`{"mode":"latest"}`), widgets.RenderContext{})
	if err != nil {
		t.Fatalf("RenderBrowser: %v", err)
	}
	html := string(out)

	for _, want := range []string{
		`class="dashi-xkcd `,
		`<span class="font-semibold">Lunar Ring</span>`,
		`text-xs text-[var(--dashi-text-muted)]">#3329</span>`,
		`src="https://imgs.xkcd.com/comics/lunar_ring.png"`,
		// comics are black-on-white: the card stays white in every theme
		`rounded-md bg-white px-6 py-2`,
		`max-w-full object-contain`,
		// alt text is only a tooltip, HTML-escaped in both attributes
		`alt="Some alt &lt;text&gt; &amp; more"`,
		`title="Some alt &lt;text&gt; &amp; more"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q in:\n%s", want, html)
		}
	}
	if strings.Contains(html, "dashi-xkcd__alt") || strings.Count(html, "Some alt") != 2 {
		t.Errorf("the alt text must not render as a visible caption:\n%s", html)
	}
	if strings.Contains(html, "<text>") {
		t.Errorf("alt text must be escaped:\n%s", html)
	}
	if strings.Count(html, "#3329") != 1 {
		t.Errorf("the number belongs in the title row only:\n%s", html)
	}
}

func TestRenderBrowser_FetchErrorIsReturned(t *testing.T) {
	// Same contract as the image renderer: a failed fetch is an error, not a
	// placeholder.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "down", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)
	client := xkcdclient.NewClient(t.TempDir())
	client.SetBaseURL(srv.URL)

	m := NewModule(client, slog.Default())
	_, err := m.RenderBrowser(json.RawMessage(`{}`), widgets.RenderContext{})
	if err == nil || !strings.Contains(err.Error(), "xkcd fetch") {
		t.Fatalf("want an xkcd fetch error, got %v", err)
	}
}

func TestRenderBrowser_InvalidJSON(t *testing.T) {
	m := NewModule(newTestClient(t, sampleComic), slog.Default())
	_, err := m.RenderBrowser(json.RawMessage(`{invalid`), widgets.RenderContext{})
	if err == nil || !strings.Contains(err.Error(), "xkcd config") {
		t.Fatalf("want an xkcd config error, got %v", err)
	}
}

func TestBrowserTemplateHasNoComputedClasses(t *testing.T) {
	// Tailwind extracts classes statically, so no class attribute may
	// interpolate data.
	for _, attr := range regexp.MustCompile(`class="[^"]*"`).FindAllString(browserHTML, -1) {
		if strings.Contains(attr, "{{") {
			t.Errorf("class attribute interpolates data: %s", attr)
		}
	}
}

func TestBrowserShipsNoAssets(t *testing.T) {
	// Styling is pure Tailwind; add a scoping test if an xkcd.css appears.
	var m any = NewModule(nil, nil)
	if _, ok := m.(widgets.BrowserAssets); ok {
		t.Error("xkcd ships no CSS/JS")
	}
}

func TestRenderBrowser_ImplementsBrowserRenderable(t *testing.T) {
	var _ widgets.BrowserRenderable = NewModule(nil, nil)
}

// Only the comic card links to the comic's page on xkcd.com, in a new tab;
// the title and the rest of the widget are not a link.
func TestRenderBrowser_ComicCardLinksInNewTab(t *testing.T) {
	m := NewModule(newTestClient(t, sampleComic), slog.Default())
	out, err := m.RenderBrowser(json.RawMessage(`{"mode":"latest"}`), widgets.RenderContext{})
	if err != nil {
		t.Fatalf("RenderBrowser: %v", err)
	}
	html := string(out)

	if !strings.HasPrefix(html, `<figure class="dashi-xkcd `) {
		t.Errorf("the widget root must not be the link:\n%s", html)
	}
	if strings.Count(html, "<a ") != 1 {
		t.Errorf("want exactly one link:\n%s", html)
	}
	link := html[strings.Index(html, "<a "):]
	link = link[:strings.Index(link, "</a>")]
	for _, want := range []string{
		`class="dashi-xkcd__card `,
		`href="https://xkcd.com/3329/" target="_blank" rel="noopener noreferrer"`,
		`<img class="dashi-xkcd__image`,
	} {
		if !strings.Contains(link, want) {
			t.Errorf("missing %q in the link:\n%s", want, link)
		}
	}
	if strings.Contains(link, "Lunar Ring</span>") {
		t.Errorf("the title must stay outside the link:\n%s", link)
	}
}
