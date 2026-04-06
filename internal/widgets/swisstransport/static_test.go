package swisstransport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
	tr "github.com/andresbott/dashi/internal/swisstransport"
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
		"station":      map[string]any{"id": "8500073", "name": "Basel, Aeschenplatz"},
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

	renderer := NewStaticRenderer(client)
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

	renderer := NewStaticRenderer(client)
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

	renderer := NewStaticRenderer(client)
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

	renderer := NewStaticRenderer(client)
	config := json.RawMessage(`{invalid json}`)

	_, err := renderer(config, widgets.RenderContext{})
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "transport config") {
		t.Errorf("expected 'transport config' in error, got: %v", err)
	}
}
