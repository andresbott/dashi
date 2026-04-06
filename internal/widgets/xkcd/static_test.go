package xkcd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/widgets"
	xkcdclient "github.com/andresbott/dashi/internal/xkcd"
)

func newTestClient(t *testing.T, response map[string]any) *xkcdclient.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(response)
	}))
	t.Cleanup(srv.Close)

	client := xkcdclient.NewClient(t.TempDir())
	client.SetBaseURL(srv.URL)
	return client
}

func TestRenderStatic_LatestMode(t *testing.T) {
	client := newTestClient(t, map[string]any{
		"num": 3228, "title": "Day Counter", "safe_title": "Day Counter",
		"img": "https://imgs.xkcd.com/comics/day_counter.png",
		"alt": "It has been ...", "day": "3", "month": "4", "year": "2026",
	})

	renderer := NewStaticRenderer(client)
	config := json.RawMessage(`{"mode": "latest"}`)

	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "Day Counter") {
		t.Errorf("expected title in output, got: %s", html)
	}
	if !strings.Contains(html, "day_counter.png") {
		t.Errorf("expected image URL in output, got: %s", html)
	}
	if !strings.Contains(html, "It has been") {
		t.Errorf("expected alt text in output, got: %s", html)
	}
}

func TestRenderStatic_RandomMode(t *testing.T) {
	client := newTestClient(t, map[string]any{
		"num": 614, "title": "Woodpecker", "safe_title": "Woodpecker",
		"img": "https://imgs.xkcd.com/comics/woodpecker.png",
		"alt": "If you don't have an emergency", "day": "9", "month": "7", "year": "2009",
	})

	renderer := NewStaticRenderer(client)
	config := json.RawMessage(`{"mode": "random"}`)

	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, `<img`) {
		t.Errorf("expected img tag in output, got: %s", html)
	}
}

func TestRenderStatic_DefaultMode(t *testing.T) {
	client := newTestClient(t, map[string]any{
		"num": 3228, "title": "Day Counter", "safe_title": "Day Counter",
		"img": "https://imgs.xkcd.com/comics/day_counter.png",
		"alt": "alt text", "day": "3", "month": "4", "year": "2026",
	})

	renderer := NewStaticRenderer(client)
	config := json.RawMessage(`{}`)

	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "Day Counter") {
		t.Errorf("expected title in output, got: %s", html)
	}
}

func TestRenderStatic_EmptyConfig(t *testing.T) {
	client := newTestClient(t, map[string]any{
		"num": 3228, "title": "Day Counter", "safe_title": "Day Counter",
		"img": "https://imgs.xkcd.com/comics/day_counter.png",
		"alt": "alt text", "day": "3", "month": "4", "year": "2026",
	})

	renderer := NewStaticRenderer(client)
	config := json.RawMessage(``)

	got, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := string(got)
	if !strings.Contains(html, "Day Counter") {
		t.Errorf("expected default latest comic, got: %s", html)
	}
}

func TestRenderStatic_InvalidJSON(t *testing.T) {
	client := newTestClient(t, map[string]any{
		"num": 1, "title": "t", "safe_title": "t", "img": "u", "alt": "a",
		"day": "1", "month": "1", "year": "2006",
	})

	renderer := NewStaticRenderer(client)
	config := json.RawMessage(`{invalid json}`)

	_, err := renderer(config, widgets.RenderContext{})
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "xkcd config") {
		t.Errorf("expected 'xkcd config' in error, got: %v", err)
	}
}
