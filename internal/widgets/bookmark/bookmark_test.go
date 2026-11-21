package bookmark

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
)

func TestRenderStatic_FullConfig(t *testing.T) {
	config := json.RawMessage(`{
		"url": "https://example.com",
		"icon": "ti-home",
		"title": "Example",
		"subtitle": "A test bookmark"
	}`)

	renderer := NewStaticRenderer()
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, `href="https://example.com"`) {
		t.Errorf("expected href in output, got: %s", html)
	}
	if !strings.Contains(html, "Example") {
		t.Errorf("expected title in output, got: %s", html)
	}
	if !strings.Contains(html, "A test bookmark") {
		t.Errorf("expected subtitle in output, got: %s", html)
	}
}

func TestRenderStatic_NoSubtitle(t *testing.T) {
	config := json.RawMessage(`{
		"url": "https://example.com",
		"icon": "ti-home",
		"title": "Example",
		"subtitle": ""
	}`)

	renderer := NewStaticRenderer()
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "Example") {
		t.Errorf("expected title in output, got: %s", html)
	}
	if strings.Contains(html, `<span class="bookmark-subtitle">`) {
		t.Errorf("should not contain subtitle span element, got: %s", html)
	}
}

func TestRenderStatic_EmptyURL(t *testing.T) {
	config := json.RawMessage(`{
		"url": "",
		"title": "Empty URL"
	}`)

	renderer := NewStaticRenderer()
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, `href="#"`) {
		t.Errorf("expected href='#' for empty URL, got: %s", html)
	}
}

// Protocol-relative URLs leave the server, so the rule rejects them; a
// legacy stored one renders inert.
func TestRenderStatic_ProtocolRelativeURL(t *testing.T) {
	config := json.RawMessage(`{
		"url": "//example.com/path",
		"title": "Protocol Relative"
	}`)

	renderer := NewStaticRenderer()
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, `href="#"`) || strings.Contains(html, "example.com") {
		t.Errorf("expected protocol-relative URL rendered as '#', got: %s", html)
	}
}

type urlCase struct {
	URL   string `json:"url"`
	Valid bool   `json:"valid"`
}

// loadURLCases reads the case table shared with the editor's url.test.ts.
func loadURLCases(t *testing.T) []urlCase {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "url_cases.json"))
	if err != nil {
		t.Fatalf("read url cases: %v", err)
	}
	var file struct {
		Cases []urlCase `json:"cases"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("parse url cases: %v", err)
	}
	if len(file.Cases) == 0 {
		t.Fatal("no url cases")
	}
	return file.Cases
}

func TestValidateURL_SharedCases(t *testing.T) {
	for _, tc := range loadURLCases(t) {
		t.Run(tc.URL, func(t *testing.T) {
			err := validateURL(tc.URL)
			if tc.Valid && err != nil {
				t.Errorf("validateURL(%q) = %v, want nil", tc.URL, err)
			}
			if !tc.Valid && err == nil {
				t.Errorf("validateURL(%q) = nil, want error", tc.URL)
			}
		})
	}
}

// Every URL the rule rejects must render as the inert "#", whatever the
// stored dashboard carries.
func TestRenderStatic_RejectedURLsRenderInert(t *testing.T) {
	renderer := NewStaticRenderer()
	for _, tc := range loadURLCases(t) {
		if tc.Valid {
			continue
		}
		t.Run(tc.URL, func(t *testing.T) {
			config, err := json.Marshal(bookmarkConfig{URL: tc.URL, Title: "T"})
			if err != nil {
				t.Fatal(err)
			}
			got, err := renderer(config, widgets.RenderContext{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(string(got), `href="#"`) {
				t.Errorf("want href=\"#\" for %q, got: %s", tc.URL, got)
			}
		})
	}
}

// html/template is a second layer under sanitizeURL: even if a dangerous
// URL reached the template, its URL filter replaces it.
func TestTemplate_FiltersDangerousSchemes(t *testing.T) {
	for _, url := range []string{"javascript:alert(1)", "vbscript:x", "data:text/html,x"} {
		t.Run(url, func(t *testing.T) {
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, bookmarkData{URL: url}); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(buf.String(), `href="#ZgotmplZ"`) {
				t.Errorf("template did not filter %q: %s", url, buf.String())
			}
		})
	}
}

func TestRenderStatic_RelativeURL(t *testing.T) {
	config := json.RawMessage(`{
		"url": "/dashboard/settings",
		"title": "Relative Path"
	}`)

	renderer := NewStaticRenderer()
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, `href="/dashboard/settings"`) {
		t.Errorf("expected relative URL preserved, got: %s", html)
	}
}

func TestRenderStatic_DangerousURL(t *testing.T) {
	config := json.RawMessage(`{
		"url": "javascript:alert('xss')",
		"title": "Dangerous"
	}`)

	renderer := NewStaticRenderer()
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, `href="#"`) {
		t.Errorf("expected dangerous URL sanitized to '#', got: %s", html)
	}
	if strings.Contains(html, "javascript:") {
		t.Errorf("dangerous javascript: scheme should be removed, got: %s", html)
	}
}

func TestRenderStatic_InvalidJSON(t *testing.T) {
	config := json.RawMessage(`{invalid json}`)

	renderer := NewStaticRenderer()
	_, err := renderer(config, widgets.RenderContext{})
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "bookmark config") {
		t.Errorf("expected 'bookmark config' in error, got: %v", err)
	}
}

func TestRenderStatic_EmptyConfig(t *testing.T) {
	config := json.RawMessage(``)

	renderer := NewStaticRenderer()
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, `href="#"`) {
		t.Errorf("expected href='#' for empty config, got: %s", html)
	}
}

func TestRenderStatic_HTTPUrl(t *testing.T) {
	config := json.RawMessage(`{
		"url": "http://example.com",
		"title": "HTTP Link"
	}`)

	renderer := NewStaticRenderer()
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, `href="http://example.com"`) {
		t.Errorf("expected http:// URL preserved, got: %s", html)
	}
}
