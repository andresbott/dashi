package bookmark

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRenderStatic_FullConfig(t *testing.T) {
	config := json.RawMessage(`{
		"url": "https://example.com",
		"icon": "ti-home",
		"title": "Example",
		"subtitle": "A test bookmark"
	}`)

	renderer := NewStaticRenderer()
	got, err := renderer(config)
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
	got, err := renderer(config)
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
