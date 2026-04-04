package themes

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewStore_LoadsEmbeddedDefault(t *testing.T) {
	store := NewStore("")
	themes := store.List()

	found := false
	for _, th := range themes {
		if th.Name == "default" {
			found = true
			if th.Type != ThemeTypeFont {
				t.Errorf("default theme type = %q, want %q", th.Type, ThemeTypeFont)
			}
		}
	}
	if !found {
		t.Error("expected default theme to be loaded")
	}
}

func TestNewStore_LoadsUserTheme(t *testing.T) {
	dir := t.TempDir()
	themeDir := filepath.Join(dir, "my-icons")
	iconsDir := filepath.Join(themeDir, "widgets", "weather", "icons")
	if err := os.MkdirAll(iconsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	manifest := `name: "My Icons"
description: "Custom weather icons"
type: image
`
	if err := os.WriteFile(filepath.Join(themeDir, "theme.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(iconsDir, "clear-sky.svg"), []byte("<svg/>"), 0o644); err != nil {
		t.Fatal(err)
	}

	store := NewStore(dir)
	themes := store.List()

	found := false
	for _, th := range themes {
		if th.Name == "my-icons" {
			found = true
			if th.Type != ThemeTypeImage {
				t.Errorf("theme type = %q, want %q", th.Type, ThemeTypeImage)
			}
			if th.Description != "Custom weather icons" {
				t.Errorf("theme description = %q, want %q", th.Description, "Custom weather icons")
			}
		}
	}
	if !found {
		t.Error("expected my-icons theme to be loaded")
	}
}

func TestStore_UserThemeOverridesEmbedded(t *testing.T) {
	dir := t.TempDir()
	themeDir := filepath.Join(dir, "default")
	if err := os.MkdirAll(themeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	manifest := `name: "default"
description: "User override of default"
type: image
`
	if err := os.WriteFile(filepath.Join(themeDir, "theme.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	store := NewStore(dir)
	th, ok := store.Get("default")
	if !ok {
		t.Fatal("expected default theme")
	}
	if th.Description != "User override of default" {
		t.Errorf("description = %q, want user override", th.Description)
	}
}

func TestStore_ResolveIcon_FontTheme(t *testing.T) {
	store := NewStore("")
	result, err := store.ResolveIcon("default", "clear-sky")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type != ThemeTypeFont {
		t.Errorf("type = %q, want %q", result.Type, ThemeTypeFont)
	}
	if result.CSSClass != "ti ti-sun" {
		t.Errorf("css class = %q, want %q", result.CSSClass, "ti ti-sun")
	}
}

func TestStore_ResolveIcon_ImageTheme(t *testing.T) {
	dir := t.TempDir()
	themeDir := filepath.Join(dir, "custom")
	iconsDir := filepath.Join(themeDir, "widgets", "weather", "icons")
	if err := os.MkdirAll(iconsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	manifest := `name: "custom"
description: "Custom icons"
type: image
`
	if err := os.WriteFile(filepath.Join(themeDir, "theme.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(iconsDir, "clear-sky.svg"), []byte("<svg>sun</svg>"), 0o644); err != nil {
		t.Fatal(err)
	}

	store := NewStore(dir)
	result, err := store.ResolveIcon("custom", "clear-sky")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Type != ThemeTypeImage {
		t.Errorf("type = %q, want %q", result.Type, ThemeTypeImage)
	}
	if result.FilePath == "" {
		t.Error("expected file path to be set")
	}
}

func TestStore_ResolveIcon_UnknownTheme(t *testing.T) {
	store := NewStore("")
	_, err := store.ResolveIcon("nonexistent", "clear-sky")
	if err == nil {
		t.Error("expected error for unknown theme")
	}
}

func TestStore_ResolveIcon_UnknownIcon(t *testing.T) {
	store := NewStore("")
	_, err := store.ResolveIcon("default", "nonexistent-icon")
	if err == nil {
		t.Error("expected error for unknown icon in font theme")
	}
}
