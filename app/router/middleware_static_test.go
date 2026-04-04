package router

import (
	"bytes"
	"encoding/json"
	"html/template"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/dashboard"
	dashimage "github.com/andresbott/dashi/internal/dashboard/image"
	dashstatic "github.com/andresbott/dashi/internal/dashboard/static"
	"github.com/andresbott/dashi/internal/widgets"
)

func TestStaticDashboardMiddleware_StaticDashboard(t *testing.T) {
	dir := t.TempDir()
	store := dashboard.NewStore(dir)

	_, err := store.Create(dashboard.Dashboard{
		ID:   "abc123",
		Name: "Static Test",
		Type: "static",
		Container: dashboard.Container{
			MaxWidth:        "100%",
			VerticalAlign:   "top",
			HorizontalAlign: "center",
		},
		Rows: []dashboard.Row{
			{
				ID:     "r1",
				Height: "auto",
				Width:  "100%",
				Widgets: []dashboard.Widget{
					{ID: "w1", Type: "test", Title: "W", Width: 12, Config: json.RawMessage(`{}`)},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("create dashboard: %v", err)
	}

	reg := widgets.NewRegistry()
	reg.Register("test", func(config json.RawMessage) (template.HTML, error) {
		return template.HTML("<p>static-content</p>"), nil
	})
	renderer := dashstatic.NewRenderer(reg)

	spaHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("SPA"))
	})

	imageRenderer := dashimage.NewRenderer()
	mid := NewStaticDashboardMiddleware(store, renderer, imageRenderer)
	handler := mid(spaHandler)

	req := httptest.NewRequest("GET", "/abc123", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "static-content") {
		t.Errorf("expected static HTML, got: %s", body)
	}
	if strings.Contains(body, "SPA") {
		t.Error("should not fall through to SPA")
	}
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("expected text/html content type, got: %s", ct)
	}
}

func TestStaticDashboardMiddleware_InteractiveDashboard(t *testing.T) {
	dir := t.TempDir()
	store := dashboard.NewStore(dir)

	_, err := store.Create(dashboard.Dashboard{
		ID:   "xyz789",
		Name: "Interactive Test",
		Type: "interactive",
		Container: dashboard.Container{
			MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center",
		},
		Rows: []dashboard.Row{},
	})
	if err != nil {
		t.Fatalf("create dashboard: %v", err)
	}

	reg := widgets.NewRegistry()
	renderer := dashstatic.NewRenderer(reg)

	spaHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("SPA"))
	})

	imageRenderer := dashimage.NewRenderer()
	mid := NewStaticDashboardMiddleware(store, renderer, imageRenderer)
	handler := mid(spaHandler)

	req := httptest.NewRequest("GET", "/xyz789", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "SPA" {
		t.Errorf("expected SPA fallthrough, got: %s", rec.Body.String())
	}
}

func TestStaticDashboardMiddleware_NonDashboardPath(t *testing.T) {
	dir := t.TempDir()
	store := dashboard.NewStore(dir)

	reg := widgets.NewRegistry()
	renderer := dashstatic.NewRenderer(reg)

	spaHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("SPA"))
	})

	imageRenderer := dashimage.NewRenderer()
	mid := NewStaticDashboardMiddleware(store, renderer, imageRenderer)
	handler := mid(spaHandler)

	req := httptest.NewRequest("GET", "/dashboards", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "SPA" {
		t.Errorf("expected SPA fallthrough for /dashboards, got: %s", rec.Body.String())
	}
}

func TestStaticDashboardMiddleware_APIPath(t *testing.T) {
	dir := t.TempDir()
	store := dashboard.NewStore(dir)

	reg := widgets.NewRegistry()
	renderer := dashstatic.NewRenderer(reg)

	spaHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("SPA"))
	})

	imageRenderer := dashimage.NewRenderer()
	mid := NewStaticDashboardMiddleware(store, renderer, imageRenderer)
	handler := mid(spaHandler)

	req := httptest.NewRequest("GET", "/api/v0/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "SPA" {
		t.Errorf("expected SPA fallthrough for API paths, got: %s", rec.Body.String())
	}
}

func TestStaticDashboardMiddleware_ImageDashboard(t *testing.T) {
	dir := t.TempDir()
	store := dashboard.NewStore(dir)

	_, err := store.Create(dashboard.Dashboard{
		ID:   "img123",
		Name: "Image Test",
		Type: "image",
		Container: dashboard.Container{
			MaxWidth:        "100%",
			VerticalAlign:   "top",
			HorizontalAlign: "center",
		},
		ImageConfig: &dashboard.ImageConfig{
			Width:  400,
			Height: 0,
		},
		Rows: []dashboard.Row{
			{
				ID:     "r1",
				Height: "auto",
				Width:  "100%",
				Widgets: []dashboard.Widget{
					{ID: "w1", Type: "test", Title: "W", Width: 12, Config: json.RawMessage(`{}`)},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("create dashboard: %v", err)
	}

	reg := widgets.NewRegistry()
	reg.Register("test", func(config json.RawMessage) (template.HTML, error) {
		return template.HTML("<p>image-content</p>"), nil
	})
	staticRenderer := dashstatic.NewRenderer(reg)
	imageRenderer := dashimage.NewRenderer()

	spaHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("SPA"))
	})

	mid := NewStaticDashboardMiddleware(store, staticRenderer, imageRenderer)
	handler := mid(spaHandler)

	req := httptest.NewRequest("GET", "/img123", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "image/png" {
		t.Errorf("expected image/png content type, got: %s", ct)
	}

	_, err = png.Decode(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatalf("response is not a valid PNG: %v", err)
	}
}
