package clock

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/andresbott/dashi/internal/widgets"
)

func TestRenderStatic_24h(t *testing.T) {
	config := json.RawMessage(`{"hour12": false, "showSeconds": false, "showDate": false}`)
	now := time.Date(2026, 4, 3, 14, 30, 45, 0, time.UTC)

	renderer := NewStaticRenderer(func() time.Time { return now })
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "14:30") {
		t.Errorf("expected 24h time '14:30' in output, got: %s", html)
	}
	if strings.Contains(html, "45") {
		t.Errorf("should not contain seconds, got: %s", html)
	}
}

func TestRenderStatic_12h_WithSeconds(t *testing.T) {
	config := json.RawMessage(`{"hour12": true, "showSeconds": true, "showDate": false}`)
	now := time.Date(2026, 4, 3, 14, 30, 45, 0, time.UTC)

	renderer := NewStaticRenderer(func() time.Time { return now })
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "2:30:45 PM") {
		t.Errorf("expected 12h time with seconds in output, got: %s", html)
	}
}

func TestRenderStatic_WithDate(t *testing.T) {
	config := json.RawMessage(`{"hour12": false, "showSeconds": false, "showDate": true}`)
	now := time.Date(2026, 4, 3, 14, 30, 45, 0, time.UTC)

	renderer := NewStaticRenderer(func() time.Time { return now })
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "14:30") {
		t.Errorf("expected time in output, got: %s", html)
	}
	if !strings.Contains(html, "Friday, April 3, 2026") {
		t.Errorf("expected date in output, got: %s", html)
	}
}

func TestRenderStatic_EmptyConfig(t *testing.T) {
	renderer := NewStaticRenderer(func() time.Time {
		return time.Date(2026, 4, 3, 9, 5, 0, 0, time.UTC)
	})
	got, err := renderer(json.RawMessage(`{}`), widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "09:05") {
		t.Errorf("expected 24h default time, got: %s", html)
	}
}

func TestRenderStatic_WithFont(t *testing.T) {
	now := time.Date(2026, 4, 3, 14, 30, 45, 0, time.UTC)
	renderer := NewStaticRenderer(func() time.Time { return now })
	config := json.RawMessage(`{"hour12": false, "showSeconds": false, "showDate": false, "font": "Inter"}`)
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	html := string(got)
	if !strings.Contains(html, "font-family: 'Inter';") {
		t.Errorf("expected font-family in output, got: %s", html)
	}
}

func TestRenderStatic_NoFont(t *testing.T) {
	now := time.Date(2026, 4, 3, 14, 30, 45, 0, time.UTC)
	renderer := NewStaticRenderer(func() time.Time { return now })
	config := json.RawMessage(`{"hour12": false, "showSeconds": false, "showDate": false}`)
	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	html := string(got)
	if strings.Contains(html, "font-family") {
		t.Errorf("expected no font-family when font not set, got: %s", html)
	}
}
