package sysinfo

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
)

func TestRenderBrowserEmitsOneSlotPerConfiguredDisk(t *testing.T) {
	m := NewModule(nil)
	cfg := json.RawMessage(`{"showMemory":true,"showUptime":true,"disks":["/","/home"]}`)

	out, err := m.RenderBrowser(cfg, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("RenderBrowser: %v", err)
	}
	html := string(out)

	// The slot count comes from config, not from data — that is why the
	// skeleton must be rendered server-side (spec: variable-length content).
	if !strings.Contains(html, `data-disk="/"`) {
		t.Errorf(`missing data-disk="/" in:\n%s`, html)
	}
	if !strings.Contains(html, `data-disk="/home"`) {
		t.Errorf(`missing data-disk="/home" in:\n%s`, html)
	}
	if !strings.Contains(html, `data-mem`) {
		t.Errorf("missing memory slot in:\n%s", html)
	}
	if !strings.Contains(html, `data-uptime`) {
		t.Errorf("missing uptime slot in:\n%s", html)
	}
}

func TestRenderBrowserOmitsDisabledSections(t *testing.T) {
	m := NewModule(nil)
	out, err := m.RenderBrowser(json.RawMessage(`{}`), widgets.RenderContext{})
	if err != nil {
		t.Fatalf("RenderBrowser: %v", err)
	}
	html := string(out)
	if strings.Contains(html, "data-mem") || strings.Contains(html, "data-uptime") {
		t.Errorf("memory and uptime must be absent when disabled:\n%s", html)
	}
}

func TestRenderBrowserDoesNotUseComputedTailwindClasses(t *testing.T) {
	// Tailwind extracts class names statically; a computed class silently
	// produces no CSS. Bar widths must therefore be inline styles.
	m := NewModule(nil)
	out, _ := m.RenderBrowser(json.RawMessage(`{"showMemory":true}`), widgets.RenderContext{})
	if strings.Contains(string(out), `class="w-[`) {
		t.Error("dynamic width must use an inline style, not a Tailwind arbitrary value")
	}
}

func TestBrowserAssetsArePresent(t *testing.T) {
	m := NewModule(nil)
	if !strings.Contains(string(m.CSS()), ".dashi-sysinfo") {
		t.Error("sysinfo.css must be scoped to .dashi-sysinfo")
	}
	js := string(m.JS())
	if !strings.Contains(js, "poll(") {
		t.Error("sysinfo.js must refresh through the shared poll() helper")
	}
	if strings.Contains(js, "innerHTML") || strings.Contains(js, "createElement") {
		t.Error("widget JS must not construct markup (decision D3)")
	}
}

func TestCSSIsolation(t *testing.T) {
	// Every selector must descend from the widget's root class to avoid
	// polluting other widgets. This enforces the isolation contract.
	m := NewModule(nil)
	css := string(m.CSS())
	if css == "" {
		return // no stylesheet, nothing to check
	}

	// Parse selectors naively — look for lines that start a rule block.
	// Valid: `.dashi-sysinfo__bar-fill`, `.dashi-sysinfo[data-dashi-error]`
	// Invalid: `.bar-fill`, `div`, `.something-else`
	for _, line := range strings.Split(css, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") {
			continue
		}
		if strings.Contains(line, "{") && !strings.HasPrefix(line, ".dashi-sysinfo") {
			t.Errorf("selector does not descend from .dashi-sysinfo: %s", line)
		}
	}
}
