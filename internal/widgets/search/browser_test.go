package search

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
)

func renderBrowser(t *testing.T, config string) string {
	t.Helper()
	out, err := NewModule().RenderBrowser(json.RawMessage(config), widgets.RenderContext{})
	if err != nil {
		t.Fatalf("RenderBrowser: %v", err)
	}
	return string(out)
}

func TestRenderBrowserIsAPlainGetFormToTheEngine(t *testing.T) {
	html := renderBrowser(t, `{"engine":"duckduckgo","placeholder":"Search the web..."}`)

	for _, want := range []string{
		`class="dashi-search`,
		`<form`,
		`action="https://duckduckgo.com/"`,
		`method="get"`,
		`target="_blank"`,
		`rel="noopener"`,
		`name="q"`,
		`placeholder="Search the web..."`,
		`type="submit"`,
		`dashi-search__engine`,
		`>DuckDuckGo</span>`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q in:\n%s", want, html)
		}
	}
	if strings.Contains(html, "<script") {
		t.Errorf("search must work without JavaScript:\n%s", html)
	}
}

func TestRenderBrowserUnknownEngineSubmitsToGoogle(t *testing.T) {
	html := renderBrowser(t, `{"engine":"https://evil.example/collect"}`)
	if !strings.Contains(html, `action="https://www.google.com/search"`) {
		t.Errorf("unknown engine must fall back to google:\n%s", html)
	}
	if strings.Contains(html, "evil.example") {
		t.Errorf("config engine value must never reach the markup:\n%s", html)
	}
	if !strings.Contains(html, `placeholder="Search..."`) {
		t.Errorf("missing default placeholder:\n%s", html)
	}
}

func TestRenderBrowserEscapesPlaceholder(t *testing.T) {
	html := renderBrowser(t, `{"placeholder":"\"><script>alert(1)</script>"}`)
	if strings.Contains(html, "<script>") || !strings.Contains(html, `placeholder="&#34;&gt;&lt;script&gt;`) {
		t.Errorf("placeholder not escaped inside its attribute:\n%s", html)
	}
}

func TestRenderBrowserRejectsMalformedConfig(t *testing.T) {
	if _, err := NewModule().RenderBrowser(json.RawMessage(`{`), widgets.RenderContext{}); err == nil {
		t.Fatal("expected an error for malformed config")
	}
}
