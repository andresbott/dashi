package router

import (
	"bytes"
	"encoding/json"
	"html/template"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andresbott/dashi/internal/dashboard"
	dashimage "github.com/andresbott/dashi/internal/dashboard/image"
	dashstatic "github.com/andresbott/dashi/internal/dashboard/static"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/widgets"
)

// newTestMiddleware creates a test middleware stack with a store and test widget.
func newTestMiddleware(t *testing.T, dashboards ...dashboard.Dashboard) http.Handler {
	t.Helper()
	dir := t.TempDir()
	store := dashboard.NewStore(dir)

	for _, d := range dashboards {
		_, err := store.Create(d)
		if err != nil {
			t.Fatalf("create dashboard: %v", err)
		}
	}

	reg := widgets.NewRegistry()
	reg.Register("test", func(config json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
		return template.HTML("<p>test-content</p>"), nil
	})

	staticRenderer := dashstatic.NewRenderer(reg)
	imageRenderer := dashimage.NewRenderer()

	spaHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("SPA"))
	})

	mid := NewStaticDashboardMiddleware(store, staticRenderer, imageRenderer, themes.NewStore(""))
	return mid(spaHandler)
}

// imageDashboard creates a test image dashboard with sensible defaults.
func imageDashboard(id string, pages ...dashboard.Page) dashboard.Dashboard {
	if len(pages) == 0 {
		pages = []dashboard.Page{
			{
				Name: "Default",
				Rows: []dashboard.Row{
					{
						ID:     "r1",
						Height: "auto",
						Width:  "100%",
						Widgets: []dashboard.Widget{
							{ID: "w1", Type: "test", Title: "Test", Width: 12, Config: json.RawMessage(`{}`)},
						},
					},
				},
			},
		}
	}

	return dashboard.Dashboard{
		ID:   id,
		Name: "Test Image",
		Type: "image",
		Container: dashboard.Container{
			MaxWidth:        "100%",
			VerticalAlign:   "top",
			HorizontalAlign: "center",
		},
		Pages: pages,
	}
}

func TestImageDashboard_MissingFormatHeader(t *testing.T) {
	handler := newTestMiddleware(t, imageDashboard("test"))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Display-Width", "296")
	req.Header.Set("X-Display-Height", "152")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestImageDashboard_MissingWidthHeader(t *testing.T) {
	handler := newTestMiddleware(t, imageDashboard("test"))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Display-Format", "png")
	req.Header.Set("X-Display-Height", "152")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestImageDashboard_MissingHeightHeader(t *testing.T) {
	handler := newTestMiddleware(t, imageDashboard("test"))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Display-Format", "png")
	req.Header.Set("X-Display-Width", "296")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestImageDashboard_NoHeadersHTMLPreview(t *testing.T) {
	handler := newTestMiddleware(t, imageDashboard("test"))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 (HTML preview), got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("expected text/html content type, got %s", ct)
	}
}

func TestImageDashboard_InvalidFormat(t *testing.T) {
	handler := newTestMiddleware(t, imageDashboard("test"))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Display-Format", "invalid")
	req.Header.Set("X-Display-Width", "296")
	req.Header.Set("X-Display-Height", "152")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestImageDashboard_PNGFormat(t *testing.T) {
	handler := newTestMiddleware(t, imageDashboard("test"))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Display-Format", "png")
	req.Header.Set("X-Display-Width", "296")
	req.Header.Set("X-Display-Height", "152")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "image/png" {
		t.Errorf("expected image/png, got: %s", ct)
	}

	_, err := png.Decode(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatalf("response is not a valid PNG: %v", err)
	}
}

func TestImageDashboard_BWFormat(t *testing.T) {
	handler := newTestMiddleware(t, imageDashboard("test"))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Display-Format", "bw")
	req.Header.Set("X-Display-Width", "296")
	req.Header.Set("X-Display-Height", "152")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/octet-stream" {
		t.Errorf("expected application/octet-stream, got: %s", ct)
	}

	// Expected size: ceil(296/8) * 152 = 37 * 152 = 5624 bytes
	expectedSize := ((296 + 7) / 8) * 152
	if rec.Body.Len() != expectedSize {
		t.Errorf("expected body size %d, got %d", expectedSize, rec.Body.Len())
	}
}

func TestImageDashboard_Spectra6Format(t *testing.T) {
	handler := newTestMiddleware(t, imageDashboard("test"))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Display-Format", "spectra6")
	req.Header.Set("X-Display-Width", "100")
	req.Header.Set("X-Display-Height", "50")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "application/octet-stream" {
		t.Errorf("expected application/octet-stream, got: %s", ct)
	}

	// Expected size: (100 * 50) / 2 = 2500 bytes
	expectedSize := (100 * 50) / 2
	if rec.Body.Len() != expectedSize {
		t.Errorf("expected body size %d, got %d", expectedSize, rec.Body.Len())
	}
}

func TestImageDashboard_SwipeRightRedirect(t *testing.T) {
	pages := []dashboard.Page{
		{Name: "P0", Rows: []dashboard.Row{{ID: "r1", Height: "auto", Width: "100%", Widgets: []dashboard.Widget{{ID: "w1", Type: "test", Width: 12, Config: json.RawMessage(`{}`)}}}}},
		{Name: "P1", Rows: []dashboard.Row{{ID: "r2", Height: "auto", Width: "100%", Widgets: []dashboard.Widget{{ID: "w2", Type: "test", Width: 12, Config: json.RawMessage(`{}`)}}}}},
		{Name: "P2", Rows: []dashboard.Row{{ID: "r3", Height: "auto", Width: "100%", Widgets: []dashboard.Widget{{ID: "w3", Type: "test", Width: 12, Config: json.RawMessage(`{}`)}}}}},
	}

	handler := newTestMiddleware(t, imageDashboard("test", pages...))

	req := httptest.NewRequest("GET", "/test?page=1", nil)
	req.Header.Set("X-Display-Format", "png")
	req.Header.Set("X-Display-Width", "296")
	req.Header.Set("X-Display-Height", "152")
	req.Header.Set("X-Action", "swipe_right")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != "/test?page=2" {
		t.Errorf("expected redirect to /test?page=2, got: %s", location)
	}
}

func TestImageDashboard_SwipeRightWraps(t *testing.T) {
	pages := []dashboard.Page{
		{Name: "P0", Rows: []dashboard.Row{{ID: "r1", Height: "auto", Width: "100%", Widgets: []dashboard.Widget{{ID: "w1", Type: "test", Width: 12, Config: json.RawMessage(`{}`)}}}}},
		{Name: "P1", Rows: []dashboard.Row{{ID: "r2", Height: "auto", Width: "100%", Widgets: []dashboard.Widget{{ID: "w2", Type: "test", Width: 12, Config: json.RawMessage(`{}`)}}}}},
	}

	handler := newTestMiddleware(t, imageDashboard("test", pages...))

	req := httptest.NewRequest("GET", "/test?page=1", nil)
	req.Header.Set("X-Display-Format", "png")
	req.Header.Set("X-Display-Width", "296")
	req.Header.Set("X-Display-Height", "152")
	req.Header.Set("X-Action", "swipe_right")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != "/test?page=0" {
		t.Errorf("expected redirect to /test?page=0 (wrap), got: %s", location)
	}
}

func TestImageDashboard_SwipeLeftWraps(t *testing.T) {
	pages := []dashboard.Page{
		{Name: "P0", Rows: []dashboard.Row{{ID: "r1", Height: "auto", Width: "100%", Widgets: []dashboard.Widget{{ID: "w1", Type: "test", Width: 12, Config: json.RawMessage(`{}`)}}}}},
		{Name: "P1", Rows: []dashboard.Row{{ID: "r2", Height: "auto", Width: "100%", Widgets: []dashboard.Widget{{ID: "w2", Type: "test", Width: 12, Config: json.RawMessage(`{}`)}}}}},
		{Name: "P2", Rows: []dashboard.Row{{ID: "r3", Height: "auto", Width: "100%", Widgets: []dashboard.Widget{{ID: "w3", Type: "test", Width: 12, Config: json.RawMessage(`{}`)}}}}},
	}

	handler := newTestMiddleware(t, imageDashboard("test", pages...))

	req := httptest.NewRequest("GET", "/test?page=0", nil)
	req.Header.Set("X-Display-Format", "png")
	req.Header.Set("X-Display-Width", "296")
	req.Header.Set("X-Display-Height", "152")
	req.Header.Set("X-Action", "swipe_left")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != "/test?page=2" {
		t.Errorf("expected redirect to /test?page=2 (wrap), got: %s", location)
	}
}

func TestImageDashboard_RefreshInterval(t *testing.T) {
	pages := []dashboard.Page{
		{
			Name:            "P0",
			RefreshInterval: 300,
			Rows: []dashboard.Row{{ID: "r1", Height: "auto", Width: "100%", Widgets: []dashboard.Widget{{ID: "w1", Type: "test", Width: 12, Config: json.RawMessage(`{}`)}}}}},
	}

	handler := newTestMiddleware(t, imageDashboard("test", pages...))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Display-Format", "png")
	req.Header.Set("X-Display-Width", "296")
	req.Header.Set("X-Display-Height", "152")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	refreshInterval := rec.Header().Get("X-Refresh-Interval")
	if refreshInterval != "300" {
		t.Errorf("expected X-Refresh-Interval: 300, got: %s", refreshInterval)
	}
}

func TestImageDashboard_NoRefreshIntervalWhenZero(t *testing.T) {
	pages := []dashboard.Page{
		{
			Name:            "P0",
			RefreshInterval: 0,
			Rows: []dashboard.Row{{ID: "r1", Height: "auto", Width: "100%", Widgets: []dashboard.Widget{{ID: "w1", Type: "test", Width: 12, Config: json.RawMessage(`{}`)}}}}},
	}

	handler := newTestMiddleware(t, imageDashboard("test", pages...))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Display-Format", "png")
	req.Header.Set("X-Display-Width", "296")
	req.Header.Set("X-Display-Height", "152")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	refreshInterval := rec.Header().Get("X-Refresh-Interval")
	if refreshInterval != "" {
		t.Errorf("expected no X-Refresh-Interval header, got: %s", refreshInterval)
	}
}

func TestImageDashboard_InteractiveFallsThrough(t *testing.T) {
	dash := dashboard.Dashboard{
		ID:   "interactive",
		Name: "Interactive",
		Type: "interactive",
		Container: dashboard.Container{
			MaxWidth:        "100%",
			VerticalAlign:   "top",
			HorizontalAlign: "center",
		},
		Pages: []dashboard.Page{},
	}

	handler := newTestMiddleware(t, dash)

	req := httptest.NewRequest("GET", "/interactive", nil)
	req.Header.Set("X-Display-Format", "png")
	req.Header.Set("X-Display-Width", "296")
	req.Header.Set("X-Display-Height", "152")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "SPA" {
		t.Errorf("expected SPA fallthrough, got: %s", rec.Body.String())
	}
}

func TestImageDashboard_NonDashboardPathFallsThrough(t *testing.T) {
	handler := newTestMiddleware(t, imageDashboard("test"))

	req := httptest.NewRequest("GET", "/api/v0/health", nil)
	req.Header.Set("X-Display-Format", "png")
	req.Header.Set("X-Display-Width", "296")
	req.Header.Set("X-Display-Height", "152")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Body.String() != "SPA" {
		t.Errorf("expected SPA fallthrough, got: %s", rec.Body.String())
	}
}

func TestImageDashboard_QueryParams(t *testing.T) {
	handler := newTestMiddleware(t, imageDashboard("test"))

	req := httptest.NewRequest("GET", "/test?format=png&width=296&height=152", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "image/png" {
		t.Errorf("expected image/png, got: %s", ct)
	}

	_, err := png.Decode(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatalf("response is not a valid PNG: %v", err)
	}
}

func TestImageDashboard_QueryParamsAction(t *testing.T) {
	pages := []dashboard.Page{
		{Name: "P0", Rows: []dashboard.Row{{ID: "r1", Height: "auto", Width: "100%", Widgets: []dashboard.Widget{{ID: "w1", Type: "test", Width: 12, Config: json.RawMessage(`{}`)}}}}},
		{Name: "P1", Rows: []dashboard.Row{{ID: "r2", Height: "auto", Width: "100%", Widgets: []dashboard.Widget{{ID: "w2", Type: "test", Width: 12, Config: json.RawMessage(`{}`)}}}}},
	}

	handler := newTestMiddleware(t, imageDashboard("test", pages...))

	req := httptest.NewRequest("GET", "/test?page=0&format=png&width=296&height=152&action=swipe_right", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != "/test?page=1" {
		t.Errorf("expected redirect to /test?page=1, got: %s", location)
	}
}

func TestImageDashboard_HeadersOverrideQueryParams(t *testing.T) {
	handler := newTestMiddleware(t, imageDashboard("test"))

	req := httptest.NewRequest("GET", "/test?format=bw&width=100&height=50", nil)
	req.Header.Set("X-Display-Format", "png")
	req.Header.Set("X-Display-Width", "296")
	req.Header.Set("X-Display-Height", "152")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "image/png" {
		t.Errorf("expected image/png (header override), got: %s", ct)
	}
}

func TestImageDashboard_Rotation0(t *testing.T) {
	handler := newTestMiddleware(t, imageDashboard("test"))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Display-Format", "png")
	req.Header.Set("X-Display-Width", "800")
	req.Header.Set("X-Display-Height", "480")
	req.Header.Set("X-Display-Rotation", "0")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	img, err := png.Decode(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatalf("response is not a valid PNG: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 800 || bounds.Dy() != 480 {
		t.Errorf("expected 800x480 output, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestImageDashboard_Rotation90(t *testing.T) {
	handler := newTestMiddleware(t, imageDashboard("test"))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Display-Format", "png")
	req.Header.Set("X-Display-Width", "800")
	req.Header.Set("X-Display-Height", "480")
	req.Header.Set("X-Display-Rotation", "90")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	img, err := png.Decode(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatalf("response is not a valid PNG: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 800 || bounds.Dy() != 480 {
		t.Errorf("expected 800x480 output (native panel), got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestImageDashboard_Rotation180(t *testing.T) {
	handler := newTestMiddleware(t, imageDashboard("test"))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Display-Format", "png")
	req.Header.Set("X-Display-Width", "800")
	req.Header.Set("X-Display-Height", "480")
	req.Header.Set("X-Display-Rotation", "180")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	img, err := png.Decode(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatalf("response is not a valid PNG: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 800 || bounds.Dy() != 480 {
		t.Errorf("expected 800x480 output, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestImageDashboard_Rotation270(t *testing.T) {
	handler := newTestMiddleware(t, imageDashboard("test"))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Display-Format", "png")
	req.Header.Set("X-Display-Width", "800")
	req.Header.Set("X-Display-Height", "480")
	req.Header.Set("X-Display-Rotation", "270")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	img, err := png.Decode(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatalf("response is not a valid PNG: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 800 || bounds.Dy() != 480 {
		t.Errorf("expected 800x480 output (native panel), got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestImageDashboard_InvalidRotation(t *testing.T) {
	handler := newTestMiddleware(t, imageDashboard("test"))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Display-Format", "png")
	req.Header.Set("X-Display-Width", "800")
	req.Header.Set("X-Display-Height", "480")
	req.Header.Set("X-Display-Rotation", "45")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid rotation, got %d", rec.Code)
	}
}

func TestImageDashboard_RotationQueryParam(t *testing.T) {
	handler := newTestMiddleware(t, imageDashboard("test"))

	req := httptest.NewRequest("GET", "/test?format=png&width=800&height=480&rotation=90", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	img, err := png.Decode(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatalf("response is not a valid PNG: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 800 || bounds.Dy() != 480 {
		t.Errorf("expected 800x480 output (native panel), got %dx%d", bounds.Dx(), bounds.Dy())
	}
}
