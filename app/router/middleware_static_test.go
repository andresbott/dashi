package router

import (
	"bytes"
	"encoding/json"
	"html/template"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andresbott/dashi/internal/backgrounds"
	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/dashboard/browser"
	"github.com/andresbott/dashi/internal/data/images"
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

	bs, _ := images.NewStore(t.TempDir())
	bgStore := backgrounds.NewStore(t.TempDir(), bs)
	mid := NewDashboardMiddleware(store, browserRenderer, staticRenderer, imageRenderer, themes.NewStore(""), bgStore)
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

func TestResolveImageBackgroundWithGradientOnly(t *testing.T) {
	// When the background has a gradient but no image, resolveImageBackground
	// should return the gradient CSS and nil bytes.
	bgStore := backgrounds.NewStore(t.TempDir(), nil)
	bg := backgrounds.Background{
		ID:       "gradientonly",
		Name:     "Gradient Only",
		Gradient: &backgrounds.Gradient{Direction: "to right", Light: []string{"#000000", "#ffffff"}},
	}
	if _, err := bgStore.Create(bg); err != nil {
		t.Fatal(err)
	}

	dash := dashboard.Dashboard{BackgroundID: "gradientonly", ColorMode: "light"}
	css, imageData := resolveImageBackground(dash, bgStore)

	if css != "linear-gradient(to right,#000000,#ffffff)" {
		t.Errorf("expected gradient CSS, got %q", css)
	}
	if imageData != nil {
		t.Errorf("expected nil image data for gradient-only background, got %d bytes", len(imageData))
	}
}

func TestResolveImageBackgroundWithImagePresent(t *testing.T) {
	// When an image is present and loads successfully, resolveImageBackground
	// must return "transparent" (not empty string) to prevent the template's
	// else-branch from painting opaque theme colour over the pre-composited image.
	imageStore, err := images.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	bgStore := backgrounds.NewStore(t.TempDir(), imageStore)

	// Save a small test image
	testPNG := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG signature
	if err := imageStore.Save("test.png", testPNG); err != nil {
		t.Fatal(err)
	}

	bg := backgrounds.Background{
		ID:   "withimage",
		Name: "With Image",
		Image: &backgrounds.Image{
			Light: "shared:test.png", Fit: "cover", Position: "center", Repeat: "no-repeat",
		},
	}
	if _, err := bgStore.Create(bg); err != nil {
		t.Fatal(err)
	}

	dash := dashboard.Dashboard{BackgroundID: "withimage", ColorMode: "light"}
	css, imageData := resolveImageBackground(dash, bgStore)

	if css != "transparent" {
		t.Errorf("expected 'transparent' css to prevent theme-background paint-over, got %q", css)
	}
	if imageData == nil {
		t.Error("expected non-nil image data")
	} else if len(imageData) != len(testPNG) {
		t.Errorf("expected %d bytes, got %d", len(testPNG), len(imageData))
	}
}

func TestImageDashboardWithDanglingBackgroundRefStillRenders(t *testing.T) {
	// Create an image dashboard with a BackgroundID that references nothing.
	dash := imageDashboard("dangling-bg-test")
	dash.BackgroundID = "nonexistent-bg-id"

	handler := newTestMiddleware(t, dash)

	req := httptest.NewRequest(http.MethodGet, "/dangling-bg-test?format=png&width=200&height=100", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("a dangling background reference must not break rendering, got %d: %s", rec.Code, rec.Body)
	}
}

func TestImageDashboardBackgroundImageRendersOnCanvas(t *testing.T) {
	// Regression test for the bug where ImageCSS's empty string caused the
	// template to fall back to opaque theme background, covering the
	// pre-composited image. This test proves the full path: bgStore.LoadImage
	// is called, bytes reach RenderToImage, and the color survives onto pixels.

	// Create a distinctive solid-red 10x10 PNG
	redImg := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			redImg.SetRGBA(x, y, color.RGBA{R: 0xFF, G: 0x00, B: 0x00, A: 0xFF})
		}
	}
	var redBuf bytes.Buffer
	if err := png.Encode(&redBuf, redImg); err != nil {
		t.Fatal(err)
	}

	// Set up stores - use same imageStore for both saving and bgStore
	dashStore := dashboard.NewStore(t.TempDir())
	imageStore, err := images.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	bgStore := backgrounds.NewStore(t.TempDir(), imageStore)

	// Save the red image to the shared image store
	if err := imageStore.Save("red.png", redBuf.Bytes()); err != nil {
		t.Fatal(err)
	}

	// Create a background entity referencing it
	bg := backgrounds.Background{
		ID:   "redbg",
		Name: "Red Background",
		Image: &backgrounds.Image{
			Light:    "shared:red.png",
			Fit:      "cover",
			Position: "center",
			Repeat:   "no-repeat",
		},
	}
	if _, err := bgStore.Create(bg); err != nil {
		t.Fatal(err)
	}

	// Create an image dashboard with real content and the background
	dash := dashboard.Dashboard{
		ID:           "testredbg",
		Name:         "Red BG Test",
		Type:         "image",
		BackgroundID: "redbg",
		Container: dashboard.Container{
			MaxWidth:        "100%",
			VerticalAlign:   "top",
			HorizontalAlign: "center",
		},
		Pages: []dashboard.Page{
			{
				Name: "Main",
				Rows: []dashboard.Row{
					{
						ID:     "r1",
						Height: "200px",
						Width:  "100%",
						Widgets: []dashboard.Widget{
							{ID: "w1", Type: "test", Title: "Content", Width: 12, Config: json.RawMessage(`{}`)},
						},
					},
				},
			},
		},
	}
	if _, err := dashStore.Create(dash); err != nil {
		t.Fatal(err)
	}

	// Build the middleware stack
	reg := widgets.NewRegistry()
	reg.Register("test", func(config json.RawMessage, _ widgets.RenderContext) (template.HTML, error) {
		return template.HTML("<p>content</p>"), nil
	})
	staticRenderer := dashstatic.NewRenderer(reg)
	imageRenderer := dashimage.NewRenderer()
	browserAssets := browser.NewAssets(nil)
	browserRenderer := browser.NewRenderer(reg, browserAssets.JSTypes())
	spaHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("SPA"))
	})
	mid := NewDashboardMiddleware(dashStore, browserRenderer, staticRenderer, imageRenderer, themes.NewStore(""), bgStore)
	handler := mid(spaHandler)

	// Request the PNG
	req := httptest.NewRequest(http.MethodGet, "/testredbg?format=png&width=200&height=150", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body)
	}

	// Decode and check for red pixels
	img, err := png.Decode(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatalf("response is not a valid PNG: %v (body: %q)", err, rec.Body.String())
	}

	rgbaImg, ok := img.(*image.RGBA)
	if !ok {
		// Convert if needed
		bounds := img.Bounds()
		rgbaImg = image.NewRGBA(bounds)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				rgbaImg.Set(x, y, img.At(x, y))
			}
		}
	}

	// Assert that red pixels are present
	redCount := 0
	bounds := rgbaImg.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := rgbaImg.At(x, y).RGBA()
			// RGBA() returns 16-bit values, so full red is 0xFFFF, 0x0000, 0x0000
			if r > 0xE000 && g < 0x1000 && b < 0x1000 {
				redCount++
			}
		}
	}

	if redCount == 0 {
		t.Errorf("expected red pixels in the output, found none (image likely painted over by opaque theme background)")
	}
}
