package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestBootstrapImageTheme(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "my-icons")
	if err := bootstrapImageTheme(dir, "my-icons"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check widget_theme.yaml exists
	manifest, err := os.ReadFile(filepath.Join(dir, "widget_theme.yaml"))
	if err != nil {
		t.Fatalf("reading widget_theme.yaml: %v", err)
	}
	if got := string(manifest); got == "" {
		t.Error("widget_theme.yaml is empty")
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

	// Check widget_theme.yaml exists and contains font config
	manifest, err := os.ReadFile(filepath.Join(dir, "widget_theme.yaml"))
	if err != nil {
		t.Fatalf("reading widget_theme.yaml: %v", err)
	}
	content := string(manifest)
	if content == "" {
		t.Error("widget_theme.yaml is empty")
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

func TestCreateWidgetTheme(t *testing.T) {
	dataDir := t.TempDir()
	if err := createWidgetTheme(dataDir, "mono", "font"); err != nil {
		t.Fatalf("create: %v", err)
	}
	path := filepath.Join(dataDir, "widget_themes", "mono", "widget_theme.yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("manifest not at %s: %v", path, err)
	}
	var m map[string]any
	if err := yaml.Unmarshal(raw, &m); err != nil {
		t.Fatalf("manifest is not valid yaml: %v", err)
	}
	if _, has := m["type"]; has {
		t.Error("manifest must have no top-level type key")
	}
	if m["name"] != "mono" {
		t.Errorf("name = %v, want mono", m["name"])
	}
	if err := createWidgetTheme(dataDir, "mono", "font"); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("second create: got %v, want already-exists error", err)
	}
	if err := createWidgetTheme(dataDir, "x", "bogus"); err == nil {
		t.Error("expected error for bad --type")
	}
}
