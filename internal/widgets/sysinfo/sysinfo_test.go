package sysinfo

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
)

func TestRenderStatic_ShowMemoryAndUptime(t *testing.T) {
	config := json.RawMessage(`{
		"showMemory": true,
		"showUptime": true,
		"disks": ["/"]
	}`)

	renderer := NewStaticRenderer()
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "Memory") {
		t.Errorf("expected Memory in output, got: %s", html)
	}
	if !strings.Contains(html, "Uptime") {
		t.Errorf("expected Uptime in output, got: %s", html)
	}
}

func TestRenderStatic_EmptyConfig(t *testing.T) {
	config := json.RawMessage(`{}`)

	renderer := NewStaticRenderer()
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if strings.Contains(html, "Memory") {
		t.Errorf("should not contain Memory when not configured, got: %s", html)
	}
	if strings.Contains(html, "Uptime") {
		t.Errorf("should not contain Uptime when not configured, got: %s", html)
	}
}

func TestRenderStatic_InvalidJSON(t *testing.T) {
	config := json.RawMessage(`{invalid json}`)

	renderer := NewStaticRenderer()
	_, err := renderer(config, widgets.RenderContext{})
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "sysinfo config") {
		t.Errorf("expected 'sysinfo config' in error, got: %v", err)
	}
}

func TestRenderStatic_OnlyDisks(t *testing.T) {
	config := json.RawMessage(`{
		"showMemory": false,
		"showUptime": false,
		"disks": ["/"]
	}`)

	renderer := NewStaticRenderer()
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if strings.Contains(html, "Memory") {
		t.Errorf("should not contain Memory, got: %s", html)
	}
	if strings.Contains(html, "Uptime") {
		t.Errorf("should not contain Uptime, got: %s", html)
	}
}
