package router

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/image/font/gofont/gomono"
)

// zipFontPack builds a widget-theme pack whose only display font is gomono,
// registered under family.
func zipFontPack(t *testing.T, name, family string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	files := map[string][]byte{
		"widget_theme.yaml": []byte("name: \"" + name + "\"\nfonts:\n  - name: \"" + family + "\"\n    file: \"face.ttf\"\n"),
		"face.ttf":          gomono.TTF,
	}
	for path, data := range files {
		w, err := zw.Create(path)
		if err != nil {
			t.Fatalf("zip create %s: %v", path, err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatalf("zip write %s: %v", path, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

// TestUploadedWidgetThemeFontRendersWithoutRestart pins the fix for fonts
// registered only at startup: a pack uploaded through the editor API must
// paint in the next PNG render from the same deps, and deleting it must
// drop the font again.
func TestUploadedWidgetThemeFontRendersWithoutRestart(t *testing.T) {
	// newSharedDeps registers a global Prometheus collector, so it can only
	// run once per test binary: reuse the shared deps. The pack is deleted
	// again below, leaving them as they were.
	newTestViewerWithDashboards(t)
	deps := sharedTestDeps
	editor, err := NewEditorFromDeps(sharedTestCfg, deps)
	if err != nil {
		t.Fatalf("NewEditorFromDeps: %v", err)
	}

	// "Pack Face" contains no "mono", so until the pack exists it falls
	// back to the embedded Inter.
	html := `<!DOCTYPE html><html><body style="margin:0;background:white;">` +
		`<span style="font-family: 'Pack Face'; font-size: 40px; color: black;">WWWW iiii</span></body></html>`
	render := func() []byte {
		t.Helper()
		png, err := deps.imageRenderer.Render(html, 300, 80)
		if err != nil {
			t.Fatalf("render: %v", err)
		}
		return png
	}
	do := func(method, path string, body []byte) int {
		t.Helper()
		req := httptest.NewRequest(method, path, bytes.NewReader(body))
		rec := httptest.NewRecorder()
		editor.ServeHTTP(rec, req)
		return rec.Code
	}

	fallback := render()
	t.Cleanup(func() { _ = deps.widgetThemeStore.Delete("packed") }) // if the test stops early

	if code := do(http.MethodPost, "/api/v0/widget-themes/upload", zipFontPack(t, "packed", "Pack Face")); code != http.StatusCreated {
		t.Fatalf("upload status = %d, want %d", code, http.StatusCreated)
	}
	if bytes.Equal(render(), fallback) {
		t.Fatal("uploaded display font not used by the next render")
	}

	if code := do(http.MethodDelete, "/api/v0/widget-themes/packed", nil); code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", code, http.StatusNoContent)
	}
	if !bytes.Equal(render(), fallback) {
		t.Error("deleted pack's font is still used")
	}
}
