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
	"github.com/andresbott/dashi/internal/dashboard/browser"
	"github.com/andresbott/dashi/internal/data/backgrounds"
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
	browserAssets := browser.NewAssets(nil)
	browserRenderer := browser.NewRenderer(reg, browserAssets.JSTypes())

	spaHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("SPA"))
	})

	bs, _ := backgrounds.NewStore(t.TempDir())
	mid := NewDashboardMiddleware(store, browserRenderer, staticRenderer, imageRenderer, themes.NewStore(""), bs)
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
		t.Errorf("expected 200 (browser HTML), got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("expected text/html content type, got %s", ct)
	}
	// Image dashboards without display headers now render through the browser stack
	body := rec.Body.String()
	if body == "<p>test-content</p>" || body == "SPA" {
		t.Error("expected browser-rendered HTML, got static HTML or SPA")
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
	// After Task 7: non-image dashboards (including "interactive" type)
	// are rendered through the browser stack, not as SPA fallthrough.
	// Only image-type dashboards with display headers render as PNG.
	dash := dashboard.Dashboard{
		ID:   "interactive",
		Name: "Interactive",
		Type: "interactive",
		Container: dashboard.Container{
			MaxWidth:        "100%",
			VerticalAlign:   "top",
			HorizontalAlign: "center",
		},
		Pages: []dashboard.Page{{
			Name: "Main",
			Rows: []dashboard.Row{{
				ID: "r1", Height: "auto", Width: "100%",
				Widgets: []dashboard.Widget{{
					ID: "w1", Type: "test", Width: 12,
					Config: json.RawMessage(`{}`),
				}},
			}},
		}},
	}

	handler := newTestMiddleware(t, dash)

	req := httptest.NewRequest("GET", "/interactive", nil)
	req.Header.Set("X-Display-Format", "png")
	req.Header.Set("X-Display-Width", "296")
	req.Header.Set("X-Display-Height", "152")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Interactive dashboards now render as browser HTML even with display headers
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Errorf("expected text/html (browser rendering), got: %s", ct)
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

func TestLoadBackgroundImage_SharedPrefix(t *testing.T) {
	bs, err := backgrounds.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0xFF, 0xD8, 0xFF}
	if err := bs.Save("sunset.jpg", want); err != nil {
		t.Fatal(err)
	}

	data, name, err := loadBackgroundImage("shared:sunset.jpg", "anyid", nil, nil, bs)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if string(data) != string(want) {
		t.Fatalf("bytes mismatch")
	}
	if name != "sunset.jpg" {
		t.Fatalf("name: %q", name)
	}
}

func TestBuildBrowserBackground(t *testing.T) {
	// The browser stack references background images by URL. The litehtml
	// stack still inlines the bytes (buildBackground) because litehtml
	// fetches nothing over the network — see TestBuildBackgroundStillInlines.
	cases := []struct {
		name string
		bg   *dashboard.Background
		want string
	}{
		{"nil", nil, ""},
		{"none", &dashboard.Background{Type: "none", Value: "x"}, ""},
		{"empty value", &dashboard.Background{Type: "color", Value: ""}, ""},
		{"color", &dashboard.Background{Type: "color", Value: "#c0ffee"}, "#c0ffee"},
		{
			"gradient",
			&dashboard.Background{Type: "gradient", Value: "linear-gradient(to bottom, #fff, #000)"},
			"linear-gradient(to bottom, #fff, #000)",
		},
		{
			"theme image",
			&dashboard.Background{Type: "image", Value: "theme:default/bg.jpg"},
			"url('/api/v0/themes/default/backgrounds/bg.jpg') center/cover no-repeat",
		},
		{
			"dashboard asset",
			&dashboard.Background{Type: "image", Value: "dashboard:images/bg.png"},
			"url('/api/v0/dashboards/abc123/assets/images/bg.png') center/cover no-repeat",
		},
		{
			"shared background",
			&dashboard.Background{Type: "image", Value: "shared:sunset.jpg"},
			"url('/api/v0/data/backgrounds/sunset.jpg') center/cover no-repeat",
		},
		{"malformed theme ref", &dashboard.Background{Type: "image", Value: "theme:nofile"}, ""},
		{"unknown scheme", &dashboard.Background{Type: "image", Value: "ftp:bg.jpg"}, ""},
		{"unknown type", &dashboard.Background{Type: "weird", Value: "x"}, ""},
	}

	for _, tc := range cases {
		dash := dashboard.Dashboard{ID: "abc123", Background: tc.bg}
		if got := buildBrowserBackground(dash); got != tc.want {
			t.Errorf("%s: buildBrowserBackground = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestBuildBrowserBackgroundEscapesUnsafeCharacters(t *testing.T) {
	// The value lands in a double-quoted HTML attribute inside url('...'),
	// so quotes and angle brackets must not survive as literals.
	dash := dashboard.Dashboard{
		ID:         "abc123",
		Background: &dashboard.Background{Type: "image", Value: `dashboard:a'b"c<d>.png`},
	}
	got := buildBrowserBackground(dash)
	if strings.ContainsAny(got, `'"<>`[1:]) || strings.Count(got, "'") != 2 {
		t.Errorf("unsafe characters survived escaping: %q", got)
	}
	if !isSafeAttributeValueForTest(got) {
		t.Errorf("value would be rejected by the renderer's attribute check: %q", got)
	}
}

// isSafeAttributeValueForTest mirrors the browser renderer's check so this
// package can assert the values it produces will not be dropped there.
func isSafeAttributeValueForTest(s string) bool {
	for _, r := range s {
		switch r {
		case '"', '<', '>', '\\':
			return false
		}
		if r < 0x20 || r == 0x7F {
			return false
		}
	}
	return true
}

func TestBuildBackgroundStillInlinesForTheImageStack(t *testing.T) {
	// litehtml fetches nothing over the network, so the image stack's
	// contract — a data URI plus the raw bytes for the canvas — must not
	// change. This is a deployed-firmware contract.
	bs, err := backgrounds.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	raw := []byte{0xFF, 0xD8, 0xFF, 0xE0}
	if err := bs.Save("sunset.jpg", raw); err != nil {
		t.Fatal(err)
	}

	dash := dashboard.Dashboard{
		ID:         "abc123",
		Background: &dashboard.Background{Type: "image", Value: "shared:sunset.jpg"},
	}
	css, data := buildBackground(dash, nil, nil, bs)

	if !strings.HasPrefix(css, "url('data:image/jpeg;base64,") {
		t.Errorf("image stack must still receive a data URI, got %q", css)
	}
	if !strings.HasSuffix(css, "') center/cover no-repeat") {
		t.Errorf("image stack background lost its positioning, got %q", css)
	}
	if string(data) != string(raw) {
		t.Error("image stack must still receive the raw bytes for the canvas")
	}
}
