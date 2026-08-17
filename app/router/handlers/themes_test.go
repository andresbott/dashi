package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/andresbott/dashi/internal/themes"
	"github.com/gorilla/mux"
)

// buildThemeZip builds an in-memory zip containing a theme.yaml plus any
// additional files. Used by the CRUD handler tests.
func buildThemeZip(t *testing.T, yaml string, extra map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	if yaml != "" {
		w, err := zw.Create("theme.yaml")
		if err != nil {
			t.Fatalf("create theme.yaml: %v", err)
		}
		if _, err := w.Write([]byte(yaml)); err != nil {
			t.Fatalf("write theme.yaml: %v", err)
		}
	}
	for name, body := range extra {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create %q: %v", name, err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatalf("write %q: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func TestThemeHandler_List(t *testing.T) {
	store := themes.NewStore("")
	handler := NewThemeHandler(store, slog.Default())

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
	handler := NewThemeHandler(store, slog.Default())

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
	if err := os.MkdirAll(iconsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(themeDir, "theme.yaml"), []byte(`name: custom
type: icon
description: test
icons:
  type: image
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(iconsDir, "clear-sky.svg"), []byte(`<svg xmlns="http://www.w3.org/2000/svg"><circle r="10"/></svg>`), 0o644); err != nil {
		t.Fatal(err)
	}

	store := themes.NewStore(dir)
	handler := NewThemeHandler(store, slog.Default())

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
	handler := NewThemeHandler(store, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/api/v0/themes/nonexistent/icons/clear-sky", nil)
	req = mux.SetURLVars(req, map[string]string{"name": "nonexistent", "icon": "clear-sky"})
	rec := httptest.NewRecorder()
	handler.GetIcon(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestThemeHandler_Upload_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := themes.NewStore(dir)
	handler := NewThemeHandler(store, slog.Default())

	zipBytes := buildThemeZip(t, `name: "sunny"
type: style
description: "bright"
`, map[string]string{"readme.txt": "hi"})

	req := httptest.NewRequest(http.MethodPost, "/api/v0/themes/upload", bytes.NewReader(zipBytes))
	rec := httptest.NewRecorder()
	handler.Upload(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("upload status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var info themes.ThemeInfo
	if err := json.NewDecoder(rec.Body).Decode(&info); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}
	if info.Name != "sunny" || info.Type != themes.ThemeKindStyle {
		t.Errorf("upload response mismatch: %+v", info)
	}

	// List should include it with type populated.
	req = httptest.NewRequest(http.MethodGet, "/api/v0/themes", nil)
	rec = httptest.NewRecorder()
	handler.List(rec, req)
	var list []themes.ThemeInfo
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	var got *themes.ThemeInfo
	for i := range list {
		if list[i].Name == "sunny" {
			got = &list[i]
		}
	}
	if got == nil {
		t.Fatalf("sunny missing from list: %+v", list)
	}
	if got.Type != themes.ThemeKindStyle {
		t.Errorf("list type = %q, want %q", got.Type, themes.ThemeKindStyle)
	}

	// Download and verify we get a valid zip back.
	req = httptest.NewRequest(http.MethodGet, "/api/v0/themes/sunny/download", nil)
	req = mux.SetURLVars(req, map[string]string{"name": "sunny"})
	rec = httptest.NewRecorder()
	handler.Download(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("download status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/zip" {
		t.Errorf("content-type = %q, want application/zip", ct)
	}
	body := rec.Body.Bytes()
	zr, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatalf("download body not a valid zip: %v", err)
	}
	hasManifest := false
	for _, f := range zr.File {
		if f.Name == "theme.yaml" {
			hasManifest = true
		}
	}
	if !hasManifest {
		t.Error("downloaded zip missing theme.yaml")
	}

	// Delete, then confirm 404 on download.
	req = httptest.NewRequest(http.MethodDelete, "/api/v0/themes/sunny", nil)
	req = mux.SetURLVars(req, map[string]string{"name": "sunny"})
	rec = httptest.NewRecorder()
	handler.Delete(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
}

func TestThemeHandler_Upload_Conflict(t *testing.T) {
	store := themes.NewStore(t.TempDir())
	handler := NewThemeHandler(store, slog.Default())

	// Upload with name "default" — the embedded builtin — must collide.
	zipBytes := buildThemeZip(t, `name: "default"
type: style
`, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v0/themes/upload", bytes.NewReader(zipBytes))
	rec := httptest.NewRecorder()
	handler.Upload(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rec.Code)
	}
}

func TestThemeHandler_Upload_InvalidType(t *testing.T) {
	store := themes.NewStore(t.TempDir())
	handler := NewThemeHandler(store, slog.Default())

	zipBytes := buildThemeZip(t, `name: "x"
type: widget
`, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v0/themes/upload", bytes.NewReader(zipBytes))
	rec := httptest.NewRecorder()
	handler.Upload(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestThemeHandler_Delete_Builtin(t *testing.T) {
	store := themes.NewStore(t.TempDir())
	handler := NewThemeHandler(store, slog.Default())

	req := httptest.NewRequest(http.MethodDelete, "/api/v0/themes/default", nil)
	req = mux.SetURLVars(req, map[string]string{"name": "default"})
	rec := httptest.NewRecorder()
	handler.Delete(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}
