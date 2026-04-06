package bookmark

import (
	"encoding/json"
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
	if strings.Contains(html, "bookmark-subtitle") {
		t.Errorf("should not contain subtitle span, got: %s", html)
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
	if !strings.Contains(html, `href="//example.com/path"`) {
		t.Errorf("expected protocol-relative URL preserved, got: %s", html)
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
