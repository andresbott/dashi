package markdown

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/data/notes"
	"github.com/andresbott/dashi/internal/widgets"
)

// setupTestStore creates a temp notes store seeded with one file.
func setupTestStore(t *testing.T, filename, content string) *notes.Store {
	t.Helper()
	store, err := notes.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := store.Save(filename, content); err != nil {
		t.Fatalf("save: %v", err)
	}
	return store
}

func TestRenderStatic_BasicMarkdown(t *testing.T) {
	store := setupTestStore(t, "test.md", "# Hello\n\nThis is **bold** text.")

	renderer := NewStaticRenderer(store)
	config := json.RawMessage(`{"filename":"test.md"}`)

	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "<h1>Hello</h1>") {
		t.Errorf("expected <h1>Hello</h1>, got: %s", html)
	}
	if !strings.Contains(html, "<strong>bold</strong>") {
		t.Errorf("expected <strong>bold</strong>, got: %s", html)
	}
}

func TestRenderStatic_FileNotFound(t *testing.T) {
	store := setupTestStore(t, "exists.md", "content")

	renderer := NewStaticRenderer(store)
	config := json.RawMessage(`{"filename":"missing.md"}`)

	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "not found") {
		t.Errorf("expected 'not found' message, got: %s", html)
	}
}

func TestRenderStatic_EmptyConfig(t *testing.T) {
	store := setupTestStore(t, "test.md", "content")

	renderer := NewStaticRenderer(store)
	config := json.RawMessage(`{}`)

	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "not found") {
		t.Errorf("expected 'not found' message for empty filename, got: %s", html)
	}
}

func TestRenderStatic_InvalidJSON(t *testing.T) {
	store := setupTestStore(t, "test.md", "content")

	renderer := NewStaticRenderer(store)
	config := json.RawMessage(`not valid json at all`)

	_, err := renderer(config, widgets.RenderContext{})
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "markdown config") {
		t.Errorf("expected 'markdown config' in error, got: %v", err)
	}
}

func TestRenderStatic_ListsAndCodeBlocks(t *testing.T) {
	md := "- item one\n- item two\n\n```\ncode block\n```\n"
	store := setupTestStore(t, "lists.md", md)

	renderer := NewStaticRenderer(store)
	config := json.RawMessage(`{"filename":"lists.md"}`)

	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "<li>item one</li>") {
		t.Errorf("expected list item, got: %s", html)
	}
	if !strings.Contains(html, "<code>") {
		t.Errorf("expected code block, got: %s", html)
	}
}
