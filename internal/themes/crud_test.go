package themes

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// buildThemeZip builds an in-memory zip containing a theme.yaml plus any
// additional files. Pass yaml = "" to skip the manifest (for
// missing-manifest tests).
func buildThemeZip(t *testing.T, yaml string, extra map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	if yaml != "" {
		w, err := zw.Create("theme.yaml")
		if err != nil {
			t.Fatalf("create theme.yaml entry: %v", err)
		}
		if _, err := w.Write([]byte(yaml)); err != nil {
			t.Fatalf("write theme.yaml: %v", err)
		}
	}
	for name, body := range extra {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create %q entry: %v", name, err)
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

func TestUpload_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	zipBytes := buildThemeZip(t, `name: "cool"
type: style
description: "a cool theme"
`, map[string]string{
		"readme.txt": "hi",
	})

	info, err := store.Upload(zipBytes)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if info.Name != "cool" || info.Type != ThemeKindStyle || info.Builtin {
		t.Fatalf("unexpected info: %+v", info)
	}

	// Directory should exist with manifest + extra file.
	themeDir := filepath.Join(dir, "cool")
	if _, err := os.Stat(filepath.Join(themeDir, "theme.yaml")); err != nil {
		t.Errorf("theme.yaml missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(themeDir, "readme.txt")); err != nil {
		t.Errorf("readme.txt missing: %v", err)
	}

	// List should include the new theme.
	found := false
	for _, th := range store.List() {
		if th.Name == "cool" {
			found = true
		}
	}
	if !found {
		t.Error("uploaded theme not in List()")
	}
}

func TestUpload_MissingManifest(t *testing.T) {
	store := NewStore(t.TempDir())
	zipBytes := buildThemeZip(t, "", map[string]string{"readme.txt": "hi"})
	_, err := store.Upload(zipBytes)
	if !errors.Is(err, ErrInvalidArchive) {
		t.Fatalf("got %v, want ErrInvalidArchive", err)
	}
}

func TestUpload_InvalidType(t *testing.T) {
	store := NewStore(t.TempDir())
	zipBytes := buildThemeZip(t, `name: "x"
type: widget
`, nil)
	_, err := store.Upload(zipBytes)
	if !errors.Is(err, ErrInvalidType) {
		t.Fatalf("got %v, want ErrInvalidType", err)
	}
}

func TestUpload_MissingType(t *testing.T) {
	store := NewStore(t.TempDir())
	zipBytes := buildThemeZip(t, `name: "x"
`, nil)
	_, err := store.Upload(zipBytes)
	if !errors.Is(err, ErrInvalidType) {
		t.Fatalf("got %v, want ErrInvalidType", err)
	}
}

func TestUpload_Conflict(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	zipBytes := buildThemeZip(t, `name: "dup"
type: style
`, nil)
	if _, err := store.Upload(zipBytes); err != nil {
		t.Fatalf("first upload: %v", err)
	}
	_, err := store.Upload(zipBytes)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("got %v, want ErrConflict", err)
	}
}

func TestUpload_ConflictWithBuiltin(t *testing.T) {
	store := NewStore(t.TempDir())
	zipBytes := buildThemeZip(t, `name: "default"
type: style
`, nil)
	_, err := store.Upload(zipBytes)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("got %v, want ErrConflict (builtin collision)", err)
	}
}

func TestUpload_PathTraversalInZip(t *testing.T) {
	store := NewStore(t.TempDir())
	zipBytes := buildThemeZip(t, `name: "evil"
type: style
`, map[string]string{
		"../outside.txt": "escape",
	})
	_, err := store.Upload(zipBytes)
	if !errors.Is(err, ErrInvalidArchive) {
		t.Fatalf("got %v, want ErrInvalidArchive", err)
	}
}

func TestUpload_InvalidName(t *testing.T) {
	store := NewStore(t.TempDir())
	zipBytes := buildThemeZip(t, `name: "bad/name"
type: style
`, nil)
	_, err := store.Upload(zipBytes)
	if !errors.Is(err, ErrInvalidName) {
		t.Fatalf("got %v, want ErrInvalidName", err)
	}
}

func TestUpload_TooLarge(t *testing.T) {
	store := NewStore(t.TempDir())
	big := make([]byte, MaxUploadSize+1)
	_, err := store.Upload(big)
	if !errors.Is(err, ErrInvalidArchive) {
		t.Fatalf("got %v, want ErrInvalidArchive", err)
	}
}

func TestDelete_RemovesTheme(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	zipBytes := buildThemeZip(t, `name: "gone"
type: style
`, nil)
	if _, err := store.Upload(zipBytes); err != nil {
		t.Fatalf("upload: %v", err)
	}
	if err := store.Delete("gone"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, ok := store.Get("gone"); ok {
		t.Error("theme still present after delete")
	}
	if _, err := os.Stat(filepath.Join(dir, "gone")); !os.IsNotExist(err) {
		t.Errorf("theme dir still on disk: err=%v", err)
	}
}

func TestDelete_Missing(t *testing.T) {
	store := NewStore(t.TempDir())
	if err := store.Delete("nope"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestDelete_Builtin(t *testing.T) {
	store := NewStore(t.TempDir())
	if err := store.Delete("default"); !errors.Is(err, ErrBuiltin) {
		t.Fatalf("got %v, want ErrBuiltin", err)
	}
}

func TestZip_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	orig := buildThemeZip(t, `name: "rt"
type: style
description: "round trip"
`, map[string]string{
		"sub/file.txt": "content",
	})
	if _, err := store.Upload(orig); err != nil {
		t.Fatalf("upload: %v", err)
	}

	data, err := store.Zip("rt")
	if err != nil {
		t.Fatalf("zip: %v", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("read zip: %v", err)
	}
	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	if !names["theme.yaml"] {
		t.Error("zip missing theme.yaml")
	}
	if !names["sub/file.txt"] {
		t.Errorf("zip missing sub/file.txt, got %v", names)
	}
}

func TestZip_Builtin(t *testing.T) {
	store := NewStore(t.TempDir())
	data, err := store.Zip("default")
	if err != nil {
		t.Fatalf("zip builtin: %v", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("read zip: %v", err)
	}
	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	for _, want := range []string{"theme.yaml", "Inter-Regular.ttf", "tabler-icons.ttf", "backgrounds/bg.jpg"} {
		if !names[want] {
			t.Errorf("zip missing %q, got %v", want, names)
		}
	}
}
