package clock

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/andresbott/dashi/internal/widgets"
)

func fixedNow() time.Time {
	return time.Date(2026, 8, 17, 14, 32, 5, 0, time.UTC)
}

func TestRenderBrowserPrefillsTimeAndCarriesConfigAsDataAttributes(t *testing.T) {
	m := NewModuleWithClock(fixedNow)
	out, err := m.RenderBrowser(json.RawMessage(`{"showSeconds":true,"showDate":true}`), widgets.RenderContext{})
	if err != nil {
		t.Fatalf("RenderBrowser: %v", err)
	}
	html := string(out)

	for _, want := range []string{
		`class="dashi-clock`,
		`data-hour12="0"`,
		`data-seconds="1"`,
		`14:32:05`,                        // prefilled by Go (decision D4)
		`dashi-clock__date`,
		`Monday, August 17, 2026`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q in:\n%s", want, html)
		}
	}
}

func TestRenderBrowserOmitsDateWhenDisabled(t *testing.T) {
	m := NewModuleWithClock(fixedNow)
	out, err := m.RenderBrowser(json.RawMessage(`{}`), widgets.RenderContext{})
	if err != nil {
		t.Fatalf("RenderBrowser: %v", err)
	}
	if strings.Contains(string(out), "dashi-clock__date") {
		t.Errorf("date element should be absent when showDate is false:\n%s", out)
	}
}

func TestRenderBrowser12HourFormat(t *testing.T) {
	m := NewModuleWithClock(fixedNow)
	out, _ := m.RenderBrowser(json.RawMessage(`{"hour12":true}`), widgets.RenderContext{})
	html := string(out)
	if !strings.Contains(html, `data-hour12="1"`) {
		t.Errorf("expected data-hour12=1:\n%s", html)
	}
	if !strings.Contains(html, "2:32 PM") {
		t.Errorf("expected 12-hour prefill:\n%s", html)
	}
}

func TestBrowserAssetsArePresent(t *testing.T) {
	m := NewModule()
	if m.CSS() != nil {
		t.Errorf("clock has no custom CSS; expected nil, got %d bytes", len(m.CSS()))
	}
	if !strings.Contains(string(m.JS()), "every(") {
		t.Error("clock.js must drive updates through the shared every() helper")
	}
}

