package sysinfo

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/sysinfo"
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

// probe returns a fabricated probe result. The mountpoints are deliberately
// out of config order so a row-order assertion means something.
func probe() sysinfo.SystemInfo {
	return sysinfo.SystemInfo{
		Disks: []sysinfo.DiskInfo{
			{Mountpoint: "/", Total: 1024 * 1024 * 1024, Free: 512 * 1024 * 1024, UsedPct: 50},
			{Mountpoint: "/data", Total: 2 * 1024 * 1024 * 1024, Free: 1024 * 1024 * 1024, UsedPct: 50},
			{Mountpoint: "/boot", Total: 512 * 1024 * 1024, Free: 256 * 1024 * 1024, UsedPct: 50},
		},
	}
}

func mountpoints(rows []diskData) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.Mountpoint
	}
	return out
}

func TestImageDiskRowsFollowProbeOrder(t *testing.T) {
	// The image stack renders a PNG that deployed e-ink firmware already
	// displays; its row order must stay the probe's order, not the config's.
	cfg := sysinfoConfig{Disks: []string{"/boot", "/"}}
	got := mountpoints(buildSysinfoData(cfg, probe(), diskRowsMounted).Disks)

	want := []string{"/", "/boot"}
	if len(got) != len(want) {
		t.Fatalf("image rows = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("image rows = %v, want %v", got, want)
		}
	}
}

func TestImageOmitsUnmountedDiskWhileBrowserKeepsASlot(t *testing.T) {
	cfg := sysinfoConfig{Disks: []string{"/", "/nope"}}

	imageRows := buildSysinfoData(cfg, probe(), diskRowsMounted).Disks
	for _, r := range imageRows {
		if r.Mountpoint == "/nope" {
			t.Errorf("image renderer must not emit a row for an unmounted mountpoint: %v", mountpoints(imageRows))
		}
	}
	if len(imageRows) != 1 {
		t.Errorf("image rows = %v, want exactly [/]", mountpoints(imageRows))
	}

	// The browser stack does need the slot: sysinfo.js only writes into
	// existing elements, so a mountpoint appearing later must have a row.
	browserRows := buildSysinfoData(cfg, probe(), diskRowsConfigured).Disks
	found := false
	for _, r := range browserRows {
		if r.Mountpoint == "/nope" {
			found = true
			if r.TotalHuman != "" || r.FreeHuman != "" || r.UsedPct != 0 {
				t.Errorf("unmounted browser slot should be empty, got %+v", r)
			}
		}
	}
	if !found {
		t.Errorf("browser renderer must emit a slot per configured disk, got %v", mountpoints(browserRows))
	}
	if len(browserRows) != 2 {
		t.Errorf("browser rows = %v, want one per configured disk", mountpoints(browserRows))
	}
}

func TestBrowserDiskRowsFollowConfigOrder(t *testing.T) {
	cfg := sysinfoConfig{Disks: []string{"/boot", "/"}}
	got := mountpoints(buildSysinfoData(cfg, probe(), diskRowsConfigured).Disks)

	want := []string{"/boot", "/"}
	for i := range want {
		if i >= len(got) || got[i] != want[i] {
			t.Fatalf("browser rows = %v, want %v", got, want)
		}
	}
}

func TestHumanBytesParity(t *testing.T) {
	// Keep this table identical to the one in sysinfo.test.js.
	cases := []struct {
		in   uint64
		want string
	}{
		{0, "0 B"}, {512, "512 B"}, {1024, "1.0 KB"}, {1536, "1.5 KB"},
		{1048576, "1.0 MB"}, {1073741824, "1.0 GB"}, {1099511627776, "1.0 TB"},
	}
	for _, tc := range cases {
		if got := humanBytes(tc.in); got != tc.want {
			t.Errorf("humanBytes(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestHumanUptimeParity(t *testing.T) {
	// Keep this table identical to the one in sysinfo.test.js.
	cases := []struct {
		in   uint64
		want string
	}{
		{45, "0m"}, {90, "1m"}, {3600, "1h 0m"}, {3660, "1h 1m"}, {90000, "1d 1h 0m"},
	}
	for _, tc := range cases {
		if got := humanUptime(tc.in); got != tc.want {
			t.Errorf("humanUptime(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
