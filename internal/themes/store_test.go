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
			if !th.HasIcons {
				t.Error("default theme should have icons")
			}
			if len(th.Fonts) != 1 {
				t.Errorf("default theme fonts count = %d, want 1", len(th.Fonts))
			}
			if len(th.Fonts) > 0 && th.Fonts[0].Name != "Go Mono" {
				t.Errorf("default theme font name = %q, want %q", th.Fonts[0].Name, "Go Mono")
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
			if !th.HasIcons {
				t.Error("theme should have icons")
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

func TestStore_ResolveIcon_FontThemeCodepoint(t *testing.T) {
	store := NewStore("")
	resolved, err := store.ResolveIcon("default", "clear-sky")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.CSSClass != "ti ti-sun" {
		t.Errorf("expected CSS class 'ti ti-sun', got %q", resolved.CSSClass)
	}
	if resolved.Codepoint != "eb30" {
		t.Errorf("expected codepoint 'eb30', got %q", resolved.Codepoint)
	}
}

func TestStore_ResolveIcon_DetailIcons(t *testing.T) {
	store := NewStore("")
	detailIcons := []string{"sunrise", "sunset", "wind", "humidity", "pressure", "uv-index", "visibility", "air-quality"}
	for _, name := range detailIcons {
		result, err := store.ResolveIcon("default", name)
		if err != nil {
			t.Errorf("failed to resolve detail icon %q: %v", name, err)
			continue
		}
		if result.Type != ThemeTypeFont {
			t.Errorf("detail icon %q: type = %q, want %q", name, result.Type, ThemeTypeFont)
		}
		if result.CSSClass == "" {
			t.Errorf("detail icon %q: expected non-empty CSS class", name)
		}
		if result.Codepoint == "" {
			t.Errorf("detail icon %q: expected non-empty codepoint", name)
		}
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

func TestStore_GetFontData_Default(t *testing.T) {
	store := NewStore("")
	data, err := store.GetFontData("default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty font data for default theme")
	}
	// TTF files start with 0x00 0x01 0x00 0x00
	if len(data) < 4 || data[0] != 0x00 || data[1] != 0x01 {
		t.Error("expected valid TTF header")
	}
}

func TestStore_GetFontData_NoFont(t *testing.T) {
	store := NewStore("")
	_, err := store.GetFontData("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent theme")
	}
}

func TestStore_GetDisplayFontData_Default(t *testing.T) {
	store := NewStore("")
	data, err := store.GetDisplayFontData("default", "Go Mono")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty font data for Go Mono")
	}
	// TTF files start with 0x00 0x01 0x00 0x00
	if len(data) < 4 || data[0] != 0x00 || data[1] != 0x01 {
		t.Error("expected valid TTF header")
	}
}

func TestStore_GetDisplayFontData_NotFound(t *testing.T) {
	store := NewStore("")
	_, err := store.GetDisplayFontData("default", "NonexistentFont")
	if err == nil {
		t.Error("expected error for nonexistent font")
	}
}

func TestStore_GetDisplayFontData_UserTheme(t *testing.T) {
	dir := t.TempDir()
	themeDir := filepath.Join(dir, "custom")
	if err := os.MkdirAll(themeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a minimal TTF file (fake but valid header)
	fakeTTF := []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if err := os.WriteFile(filepath.Join(themeDir, "myfont.ttf"), fakeTTF, 0o644); err != nil {
		t.Fatal(err)
	}

	manifest := `name: "custom"
description: "Custom theme with font"
fonts:
  - name: "My Font"
    file: "myfont.ttf"
`
	if err := os.WriteFile(filepath.Join(themeDir, "theme.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	store := NewStore(dir)
	data, err := store.GetDisplayFontData("custom", "My Font")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) != len(fakeTTF) {
		t.Errorf("expected font data length %d, got %d", len(fakeTTF), len(data))
	}
}

func TestStore_NewFormatThemeWithFontsAndIcons(t *testing.T) {
	dir := t.TempDir()
	themeDir := filepath.Join(dir, "newformat")
	if err := os.MkdirAll(themeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create fake font files
	fakeTTF := []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if err := os.WriteFile(filepath.Join(themeDir, "display.ttf"), fakeTTF, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(themeDir, "icons.ttf"), fakeTTF, 0o644); err != nil {
		t.Fatal(err)
	}

	manifest := `name: "newformat"
description: "Theme with new manifest format"
fonts:
  - name: "Display Font"
    file: "display.ttf"
  - name: "Another Font"
    file: "display.ttf"
icons:
  type: font
  classPrefix: "icon-"
  fontFile: "icons.ttf"
  icons:
    test-icon:
      class: "test"
      codepoint: "1234"
`
	if err := os.WriteFile(filepath.Join(themeDir, "theme.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	store := NewStore(dir)

	// Test theme info
	th, ok := store.Get("newformat")
	if !ok {
		t.Fatal("expected newformat theme to be loaded")
	}
	if !th.HasIcons {
		t.Error("theme should have icons")
	}
	if len(th.Fonts) != 2 {
		t.Errorf("expected 2 fonts, got %d", len(th.Fonts))
	}
	if len(th.Fonts) > 0 && th.Fonts[0].Name != "Display Font" {
		t.Errorf("first font name = %q, want %q", th.Fonts[0].Name, "Display Font")
	}

	// Test icon resolution
	resolved, err := store.ResolveIcon("newformat", "test-icon")
	if err != nil {
		t.Fatalf("unexpected error resolving icon: %v", err)
	}
	if resolved.Type != ThemeTypeFont {
		t.Errorf("icon type = %q, want %q", resolved.Type, ThemeTypeFont)
	}
	if resolved.CSSClass != "icon-test" {
		t.Errorf("css class = %q, want %q", resolved.CSSClass, "icon-test")
	}

	// Test display font retrieval
	data, err := store.GetDisplayFontData("newformat", "Display Font")
	if err != nil {
		t.Fatalf("unexpected error getting display font: %v", err)
	}
	if len(data) != len(fakeTTF) {
		t.Errorf("expected font data length %d, got %d", len(fakeTTF), len(data))
	}

	// Test icon font retrieval
	iconData, err := store.GetFontData("newformat")
	if err != nil {
		t.Fatalf("unexpected error getting icon font: %v", err)
	}
	if len(iconData) != len(fakeTTF) {
		t.Errorf("expected icon font data length %d, got %d", len(fakeTTF), len(iconData))
	}
}
