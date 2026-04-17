package markdown

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/widgets"
)

// setupTestDashboard creates a temp dashboard store with a markdown file.
func setupTestDashboard(t *testing.T, filename, content string) (*dashboard.Store, string) {
	t.Helper()
	dir := t.TempDir()
	store := dashboard.NewStore(dir)

	dash, err := store.Create(dashboard.Dashboard{
		Name: "test-dash",
		Icon: "ti-test",
		Type: "interactive",
		Pages: []dashboard.Page{{
			Name: "main",
			Rows: nil,
		}},
		Container: dashboard.Container{
			MaxWidth: "1200px",
		},
	})
	if err != nil {
		t.Fatalf("create dashboard: %v", err)
	}

	// Write the markdown file via the store so it lands in the correct folder.
	if err := store.SaveAsset(dash.ID, "md/"+filename, []byte(content)); err != nil {
		t.Fatalf("save md asset: %v", err)
	}

	return store, dash.ID
}

func TestRenderStatic_BasicMarkdown(t *testing.T) {
	store, dashID := setupTestDashboard(t, "test.md", "# Hello\n\nThis is **bold** text.")

	renderer := NewStaticRenderer(store)
	config := json.RawMessage(`{"filename":"test.md"}`)
	ctx := widgets.RenderContext{DashboardID: dashID}

	got, err := renderer(config, ctx)
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
	store, dashID := setupTestDashboard(t, "exists.md", "content")

	renderer := NewStaticRenderer(store)
	config := json.RawMessage(`{"filename":"missing.md"}`)
	ctx := widgets.RenderContext{DashboardID: dashID}

	got, err := renderer(config, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "not found") {
		t.Errorf("expected 'not found' message, got: %s", html)
	}
}

func TestRenderStatic_EmptyConfig(t *testing.T) {
	store, dashID := setupTestDashboard(t, "test.md", "content")

	renderer := NewStaticRenderer(store)
	config := json.RawMessage(`{}`)
	ctx := widgets.RenderContext{DashboardID: dashID}

	got, err := renderer(config, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "not found") {
		t.Errorf("expected 'not found' message for empty filename, got: %s", html)
	}
}

func TestRenderStatic_InvalidJSON(t *testing.T) {
	store, dashID := setupTestDashboard(t, "test.md", "content")

	renderer := NewStaticRenderer(store)
	config := json.RawMessage(`not valid json at all`)
	ctx := widgets.RenderContext{DashboardID: dashID}

	_, err := renderer(config, ctx)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "markdown config") {
		t.Errorf("expected 'markdown config' in error, got: %v", err)
	}
}

func TestRenderStatic_ListsAndCodeBlocks(t *testing.T) {
	md := "- item one\n- item two\n\n```\ncode block\n```\n"
	store, dashID := setupTestDashboard(t, "lists.md", md)

	renderer := NewStaticRenderer(store)
	config := json.RawMessage(`{"filename":"lists.md"}`)
	ctx := widgets.RenderContext{DashboardID: dashID}

	got, err := renderer(config, ctx)
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
