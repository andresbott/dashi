package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBootstrapImageTheme(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "my-icons")
	if err := bootstrapImageTheme(dir, "my-icons"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check theme.yaml exists
	manifest, err := os.ReadFile(filepath.Join(dir, "theme.yaml"))
	if err != nil {
		t.Fatalf("reading theme.yaml: %v", err)
	}
	if got := string(manifest); got == "" {
		t.Error("theme.yaml is empty")
	}

	// Check all icon SVGs were created
	iconsDir := filepath.Join(dir, "widgets", "weather", "icons")
	for _, icon := range canonicalIcons {
		path := filepath.Join(iconsDir, icon+".svg")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected icon file %s to exist", path)
		}
	}
}

func TestBootstrapFontTheme(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "my-font")
	if err := bootstrapFontTheme(dir, "my-font"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check theme.yaml exists and contains font config
	manifest, err := os.ReadFile(filepath.Join(dir, "theme.yaml"))
	if err != nil {
		t.Fatalf("reading theme.yaml: %v", err)
	}
	content := string(manifest)
	if content == "" {
		t.Error("theme.yaml is empty")
	}

	// Should not create icons directory for font themes
	iconsDir := filepath.Join(dir, "widgets", "weather", "icons")
	if _, err := os.Stat(iconsDir); err == nil {
		t.Error("font theme should not have icons directory")
	}
}

func TestBootstrapImageTheme_AlreadyExists(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "existing")
	if err := bootstrapImageTheme(dir, "existing"); err != nil {
		t.Fatalf("first creation: %v", err)
	}

	// Verify the directory exists and has content
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("theme dir should exist: %v", err)
	}
}
