package swisstransport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tr "github.com/andresbott/dashi/internal/providers/swisstransport"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/andresbott/dashi/internal/widgetthemes"
)

func newTestClient(t *testing.T, response map[string]any) *tr.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(response)
	}))
	t.Cleanup(srv.Close)

	client := tr.NewClient(nil)
	client.SetBaseURL(srv.URL)
	return client
}

func TestRenderStatic_Configured(t *testing.T) {
	client := newTestClient(t, map[string]any{
		"station": map[string]any{"id": "8500073", "name": "Basel, Aeschenplatz"},
		"stationboard": []any{
			map[string]any{
				"category": "T", "number": "11", "to": "St-Louis Grenze",
				"stop": map[string]any{
					"departure": "2026-04-05T22:00:00+0200", "departureTimestamp": 1775419200,
					"delay": 0, "platform": "A",
					"prognosis": map[string]any{"departure": nil},
				},
			},
		},
	})

	renderer := NewStaticRenderer(client, widgetthemes.NewStore(""))
	config := json.RawMessage(`{"stationId": "8500073", "stationName": "Basel, Aeschenplatz", "limit": 5}`)

	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "11") {
		t.Errorf("expected line number in output, got: %s", html)
	}
	if !strings.Contains(html, "St-Louis Grenze") {
		t.Errorf("expected destination in output, got: %s", html)
	}
}

func TestRenderStatic_Unconfigured(t *testing.T) {
	client := newTestClient(t, map[string]any{
		"station": map[string]any{}, "stationboard": []any{},
	})

	renderer := NewStaticRenderer(client, widgetthemes.NewStore(""))
	config := json.RawMessage(`{}`)

	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "Set a stop") {
		t.Errorf("expected unconfigured message, got: %s", html)
	}
}

func TestRenderStatic_EmptyConfig(t *testing.T) {
	client := newTestClient(t, map[string]any{
		"station": map[string]any{}, "stationboard": []any{},
	})

	renderer := NewStaticRenderer(client, widgetthemes.NewStore(""))
	config := json.RawMessage(``)

	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "Set a stop") {
		t.Errorf("expected unconfigured message, got: %s", html)
	}
}

func TestRenderStatic_InvalidJSON(t *testing.T) {
	client := newTestClient(t, map[string]any{
		"station": map[string]any{}, "stationboard": []any{},
	})

	renderer := NewStaticRenderer(client, widgetthemes.NewStore(""))
	config := json.RawMessage(`{invalid json}`)

	_, err := renderer(config, widgets.RenderContext{})
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "transport config") {
		t.Errorf("expected 'transport config' in error, got: %v", err)
	}
}

func TestTransportUsesPaletteRoles(t *testing.T) {
	client := newTestClient(t, map[string]any{
		"station": map[string]any{"id": "8500073", "name": "Basel"},
		"stationboard": []any{
			map[string]any{
				"category": "T", "number": "11", "to": "St-Louis Grenze",
				"stop": map[string]any{
					"departure": "2026-04-05T22:00:00+0200", "departureTimestamp": 1775419200,
					"delay": 0, "platform": "A",
					"prognosis": map[string]any{"departure": nil},
				},
			},
		},
	})
	pal := themes.Palette{
		BG: "#ffffff", Surface: "#fafafa", Border: "#654321", Text: "#111111",
		TextMuted: "#123456", Primary: "#0000ff", OnPrimary: "#ffffff", Secondary: "#00ffff",
		Success: "#00aa00", Warning: "#aaaa00", Danger: "#aa0000",
	}
	renderer := NewStaticRenderer(client, widgetthemes.NewStore(""))
	got, err := renderer(json.RawMessage(`{"stationId":"8500073","stationName":"Basel"}`), widgets.RenderContext{Palette: pal})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	html := string(got)
	for _, want := range []string{"#123456", "#654321"} {
		if !strings.Contains(html, want) {
			t.Errorf("expected %s in output", want)
		}
	}
	for _, bad := range []string{"#999", "#888", "#eee", "ZgotmplZ"} {
		if strings.Contains(html, bad) {
			t.Errorf("output contains %s", bad)
		}
	}
}
