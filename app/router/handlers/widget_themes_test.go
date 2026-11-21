package handlers

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"log/slog"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/widgetthemes"
	"github.com/gorilla/mux"
)

// buildThemeZip builds an in-memory zip containing a widget_theme.yaml plus any
// additional files. Used by the CRUD handler tests.
func buildThemeZip(t *testing.T, yaml string, extra map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	if yaml != "" {
		w, err := zw.Create("widget_theme.yaml")
		if err != nil {
			t.Fatalf("create widget_theme.yaml: %v", err)
		}
		if _, err := w.Write([]byte(yaml)); err != nil {
			t.Fatalf("write widget_theme.yaml: %v", err)
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

func TestWidgetThemeHandler_List(t *testing.T) {
	store := widgetthemes.NewStore("")
	handler := NewWidgetThemeHandler(store, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/api/v0/widget-themes", nil)
	rec := httptest.NewRecorder()
	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var result []widgetthemes.Info
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

func TestWidgetThemeHandler_GetIcon_Font(t *testing.T) {
	store := widgetthemes.NewStore("")
	handler := NewWidgetThemeHandler(store, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/api/v0/widget-themes/default/icons/clear-sky", nil)
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

func TestWidgetThemeHandler_GetIcon_Image(t *testing.T) {
	dir := t.TempDir()
	themeDir := filepath.Join(dir, "custom")
	iconsDir := filepath.Join(themeDir, "widgets", "weather", "icons")
	if err := os.MkdirAll(iconsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(themeDir, "widget_theme.yaml"), []byte(`name: custom
description: test
icons:
  type: image
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(iconsDir, "clear-sky.svg"), []byte(`<svg xmlns="http://www.w3.org/2000/svg"><circle r="10"/></svg>`), 0o644); err != nil {
		t.Fatal(err)
	}

	store := widgetthemes.NewStore(dir)
	handler := NewWidgetThemeHandler(store, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/api/v0/widget-themes/custom/icons/clear-sky", nil)
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

func TestWidgetThemeHandler_GetIcon_NotFound(t *testing.T) {
	store := widgetthemes.NewStore("")
	handler := NewWidgetThemeHandler(store, slog.Default())

	// An unknown theme falls back to the default, so only an icon the
	// default lacks is a 404.
	req := httptest.NewRequest(http.MethodGet, "/api/v0/widget-themes/nonexistent/icons/no-such-icon", nil)
	req = mux.SetURLVars(req, map[string]string{"name": "nonexistent", "icon": "no-such-icon"})
	rec := httptest.NewRecorder()
	handler.GetIcon(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestWidgetThemeHandler_GetIcon_UnknownThemeFallsBack(t *testing.T) {
	store := widgetthemes.NewStore("")
	handler := NewWidgetThemeHandler(store, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/api/v0/widget-themes/nonexistent/icons/clear-sky", nil)
	req = mux.SetURLVars(req, map[string]string{"name": "nonexistent", "icon": "clear-sky"})
	rec := httptest.NewRecorder()
	handler.GetIcon(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestWidgetThemeHandler_Upload_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := widgetthemes.NewStore(dir)
	handler := NewWidgetThemeHandler(store, slog.Default())

	zipBytes := buildThemeZip(t, `name: "sunny"
description: "bright"
`, map[string]string{"readme.txt": "hi"})

	req := httptest.NewRequest(http.MethodPost, "/api/v0/widget-themes/upload", bytes.NewReader(zipBytes))
	rec := httptest.NewRecorder()
	handler.Upload(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("upload status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var info widgetthemes.Info
	if err := json.NewDecoder(rec.Body).Decode(&info); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}
	if info.Name != "sunny" {
		t.Errorf("upload response mismatch: %+v", info)
	}

	// List should include it with its description populated.
	req = httptest.NewRequest(http.MethodGet, "/api/v0/widget-themes", nil)
	rec = httptest.NewRecorder()
	handler.List(rec, req)
	var list []widgetthemes.Info
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	var got *widgetthemes.Info
	for i := range list {
		if list[i].Name == "sunny" {
			got = &list[i]
		}
	}
	if got == nil {
		t.Fatalf("sunny missing from list: %+v", list)
	}
	if got.Description != "bright" {
		t.Errorf("list description = %q, want %q", got.Description, "bright")
	}

	// Download and verify we get a valid zip back.
	req = httptest.NewRequest(http.MethodGet, "/api/v0/widget-themes/sunny/download", nil)
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
		if f.Name == "widget_theme.yaml" {
			hasManifest = true
		}
	}
	if !hasManifest {
		t.Error("downloaded zip missing widget_theme.yaml")
	}

	// Delete, then confirm 404 on download.
	req = httptest.NewRequest(http.MethodDelete, "/api/v0/widget-themes/sunny", nil)
	req = mux.SetURLVars(req, map[string]string{"name": "sunny"})
	rec = httptest.NewRecorder()
	handler.Delete(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204; body=%s", rec.Code, rec.Body.String())
	}
}

// TestWidgetThemeHandlerDownloadEscapesFilename pins M6: a pack name may
// contain a double quote, which must not break out of the filename
// parameter of Content-Disposition.
func TestWidgetThemeHandlerDownloadEscapesFilename(t *testing.T) {
	store := widgetthemes.NewStore(t.TempDir())
	handler := NewWidgetThemeHandler(store, slog.Default())
	name := `say "hi"; x=1`
	if _, err := store.Upload(buildThemeZip(t, "name: '"+name+"'\n", nil)); err != nil {
		t.Fatalf("upload: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v0/widget-themes/x/download", nil)
	req = mux.SetURLVars(req, map[string]string{"name": name})
	rec := httptest.NewRecorder()
	handler.Download(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("download status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	cd := rec.Header().Get("Content-Disposition")
	disp, params, err := mime.ParseMediaType(cd)
	if err != nil {
		t.Fatalf("unparseable Content-Disposition %q: %v", cd, err)
	}
	if disp != "attachment" || params["filename"] != name+".zip" || len(params) != 1 {
		t.Fatalf("Content-Disposition %q parsed as %q %v", cd, disp, params)
	}
}

func TestWidgetThemeHandler_Upload_Conflict(t *testing.T) {
	store := widgetthemes.NewStore(t.TempDir())
	handler := NewWidgetThemeHandler(store, slog.Default())

	// Upload with name "default" — the embedded builtin — must collide.
	zipBytes := buildThemeZip(t, `name: "default"
`, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v0/widget-themes/upload", bytes.NewReader(zipBytes))
	rec := httptest.NewRecorder()
	handler.Upload(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rec.Code)
	}
}

// TestWidgetThemeHandlerUploadRejectsLegacyManifest pins Review Focus 4 at
// the HTTP layer: an old-format pack is a 400 naming widget_theme.yaml.
func TestWidgetThemeHandlerUploadRejectsLegacyManifest(t *testing.T) {
	store := widgetthemes.NewStore(t.TempDir())
	handler := NewWidgetThemeHandler(store, slog.Default())

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("theme.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("name: \"old\"\ntype: style\n")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v0/widget-themes/upload", bytes.NewReader(buf.Bytes()))
	rec := httptest.NewRecorder()
	handler.Upload(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "widget_theme.yaml") {
		t.Errorf("body %q does not name widget_theme.yaml", rec.Body.String())
	}
}

func TestWidgetThemeHandler_Delete_Builtin(t *testing.T) {
	store := widgetthemes.NewStore(t.TempDir())
	handler := NewWidgetThemeHandler(store, slog.Default())

	req := httptest.NewRequest(http.MethodDelete, "/api/v0/widget-themes/default", nil)
	req = mux.SetURLVars(req, map[string]string{"name": "default"})
	rec := httptest.NewRecorder()
	handler.Delete(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestWidgetThemeHandler_GetFont_ServesDisplayAndIconFonts(t *testing.T) {
	store := widgetthemes.NewStore("")
	handler := NewWidgetThemeHandler(store, slog.Default())

	// The browser stack's @font-face rules (widgetthemes.Store.FontFacesCSS) point at
	// this route for both the display font and the icon font, so both names
	// must resolve or font-icon themes render tofu in a browser.
	for _, fontName := range []string{"Inter", widgetthemes.IconFontFamily("default")} {
		req := httptest.NewRequest(http.MethodGet, "/api/v0/widget-themes/default/fonts/"+fontName, nil)
		req = mux.SetURLVars(req, map[string]string{"name": "default", "font": fontName})
		rec := httptest.NewRecorder()
		handler.GetFont(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("font %q: status = %d, want 200 (body %s)", fontName, rec.Code, rec.Body.String())
			continue
		}
		if rec.Body.Len() == 0 {
			t.Errorf("font %q: empty body", fontName)
		}
		if got := rec.Header().Get("Content-Type"); got != "font/ttf" {
			t.Errorf("font %q: Content-Type = %q, want font/ttf", fontName, got)
		}
	}
}

func TestWidgetThemeHandler_GetFont_UnknownFontIs404(t *testing.T) {
	handler := NewWidgetThemeHandler(widgetthemes.NewStore(""), slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/api/v0/widget-themes/default/fonts/NoSuchFont", nil)
	req = mux.SetURLVars(req, map[string]string{"name": "default", "font": "NoSuchFont"})
	rec := httptest.NewRecorder()
	handler.GetFont(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}
