package image

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/widgets"
)

func setupStore(t *testing.T) *dashboard.Store {
	t.Helper()
	dir := t.TempDir()
	store := dashboard.NewStore(dir)
	_, err := store.Create(dashboard.Dashboard{
		ID:        "test01",
		Name:      "Test",
		Icon:      "ti-home",
		Container: dashboard.Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"},
		Pages:     []dashboard.Page{},
	})
	if err != nil {
		t.Fatalf("create dashboard: %v", err)
	}
	return store
}

func TestRenderStatic_Cover(t *testing.T) {
	store := setupStore(t)
	imgData := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a}
	if err := store.SaveAsset("test01", "photo.png", imgData); err != nil {
		t.Fatalf("save asset: %v", err)
	}

	renderer := NewStaticRenderer(store)
	config := json.RawMessage(`{"image":"photo.png","fit":"cover"}`)
	got, err := renderer(config, widgets.RenderContext{DashboardID: "test01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	expected := base64.StdEncoding.EncodeToString(imgData)
	if !strings.Contains(html, expected) {
		t.Error("expected base64 image data in output")
	}
	if !strings.Contains(html, "object-fit:cover") {
		t.Errorf("expected object-fit:cover, got: %s", html)
	}
}

func TestRenderStatic_Contain(t *testing.T) {
	store := setupStore(t)
	imgData := []byte{0xff, 0xd8, 0xff}
	if err := store.SaveAsset("test01", "pic.jpg", imgData); err != nil {
		t.Fatalf("save asset: %v", err)
	}

	renderer := NewStaticRenderer(store)
	config := json.RawMessage(`{"image":"pic.jpg","fit":"contain"}`)
	got, err := renderer(config, widgets.RenderContext{DashboardID: "test01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "object-fit:contain") {
		t.Errorf("expected object-fit:contain, got: %s", html)
	}
	if !strings.Contains(html, "data:image/jpeg;base64,") {
		t.Errorf("expected jpeg data URI, got: %s", html)
	}
}

func TestRenderStatic_DefaultFit(t *testing.T) {
	store := setupStore(t)
	if err := store.SaveAsset("test01", "img.png", []byte{0x89, 0x50}); err != nil {
		t.Fatalf("save asset: %v", err)
	}

	renderer := NewStaticRenderer(store)
	config := json.RawMessage(`{"image":"img.png"}`)
	got, err := renderer(config, widgets.RenderContext{DashboardID: "test01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "object-fit:cover") {
		t.Errorf("expected default object-fit:cover, got: %s", html)
	}
}

func TestRenderStatic_EmptyImage(t *testing.T) {
	store := setupStore(t)

	renderer := NewStaticRenderer(store)
	config := json.RawMessage(`{}`)
	got, err := renderer(config, widgets.RenderContext{DashboardID: "test01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "widget-image-empty") {
		t.Errorf("expected empty placeholder, got: %s", html)
	}
}

func TestRenderStatic_AssetNotFound(t *testing.T) {
	store := setupStore(t)

	renderer := NewStaticRenderer(store)
	config := json.RawMessage(`{"image":"missing.png"}`)
	got, err := renderer(config, widgets.RenderContext{DashboardID: "test01"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "widget-image-empty") {
		t.Errorf("expected empty placeholder for missing asset, got: %s", html)
	}
}
