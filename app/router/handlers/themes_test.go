package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/andresbott/dashi/internal/themes"
	"github.com/gorilla/mux"
)

func TestThemeHandler_List(t *testing.T) {
	store := themes.NewStore("")
	handler := NewThemeHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/v0/themes", nil)
	rec := httptest.NewRecorder()
	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var result []themes.ThemeInfo
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	found := false
	for _, th := range result {
		if th.Name == "default" {
			found = true
		}
	}
	if !found {
		t.Error("expected default theme in list")
	}
}

func TestThemeHandler_GetIcon_Font(t *testing.T) {
	store := themes.NewStore("")
	handler := NewThemeHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/v0/themes/default/icons/clear-sky", nil)
	req = mux.SetURLVars(req, map[string]string{"name": "default", "icon": "clear-sky"})
	rec := httptest.NewRecorder()
	handler.GetIcon(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var result map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if result["class"] != "ti ti-sun" {
		t.Errorf("class = %q, want %q", result["class"], "ti ti-sun")
	}
}

func TestThemeHandler_GetIcon_Image(t *testing.T) {
	dir := t.TempDir()
	themeDir := filepath.Join(dir, "custom")
	iconsDir := filepath.Join(themeDir, "widgets", "weather", "icons")
	os.MkdirAll(iconsDir, 0o755)
	os.WriteFile(filepath.Join(themeDir, "theme.yaml"), []byte(`name: custom
description: test
type: image
`), 0o644)
	os.WriteFile(filepath.Join(iconsDir, "clear-sky.svg"), []byte(`<svg xmlns="http://www.w3.org/2000/svg"><circle r="10"/></svg>`), 0o644)

	store := themes.NewStore(dir)
	handler := NewThemeHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/v0/themes/custom/icons/clear-sky", nil)
	req = mux.SetURLVars(req, map[string]string{"name": "custom", "icon": "clear-sky"})
	rec := httptest.NewRecorder()
	handler.GetIcon(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	contentType := rec.Header().Get("Content-Type")
	if contentType != "image/svg+xml" {
		t.Errorf("content-type = %q, want image/svg+xml", contentType)
	}
}

func TestThemeHandler_GetIcon_NotFound(t *testing.T) {
	store := themes.NewStore("")
	handler := NewThemeHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/v0/themes/nonexistent/icons/clear-sky", nil)
	req = mux.SetURLVars(req, map[string]string{"name": "nonexistent", "icon": "clear-sky"})
	rec := httptest.NewRecorder()
	handler.GetIcon(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
