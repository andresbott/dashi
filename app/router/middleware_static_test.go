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
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/dashboard/browser"
	dashimage "github.com/andresbott/dashi/internal/dashboard/image"
	dashstatic "github.com/andresbott/dashi/internal/dashboard/static"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/andresbott/dashi/internal/widgetthemes"
)

// newTestMiddleware creates a test middleware stack with a store and test widget.
func newTestMiddleware(t *testing.T, dashboards ...dashboard.Dashboard) http.Handler {
	t.Helper()
	return newTestMiddlewareWithEditor(t, "", dashboards...)
}

// newTestMiddlewareWithEditor is newTestMiddleware with an editor public URL,
// the base of the viewer's Edit button.
func newTestMiddlewareWithEditor(t *testing.T, editorBase string, dashboards ...dashboard.Dashboard) http.Handler {
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

	themeStore := themes.NewStore(t.TempDir())
	mid := NewDashboardMiddleware(store, browserRenderer, staticRenderer, imageRenderer, themeStore, widgetthemes.NewStore(""), editorBase)
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

func TestEditLink(t *testing.T) {
	const base = "https://editor.example/ingress"
	tests := []struct {
		name string
		dash dashboard.Dashboard
		base string
		want string
	}{
		{
			name: "interactive dashboard with flag",
			dash: dashboard.Dashboard{ID: "abc123", Type: "interactive", ShowEditButton: true},
			base: base,
			want: base + "/dashboards/abc123/edit",
		},
		{
			name: "legacy empty type counts as interactive",
			dash: dashboard.Dashboard{ID: "abc123", ShowEditButton: true},
			base: base,
			want: base + "/dashboards/abc123/edit",
		},
		{
			name: "id is escaped to one path segment",
			dash: dashboard.Dashboard{ID: "a/b c", ShowEditButton: true},
			base: base,
			want: base + "/dashboards/a%2Fb%20c/edit",
		},
		{
			name: "flag off",
			dash: dashboard.Dashboard{ID: "abc123", Type: "interactive"},
			base: base,
			want: "",
		},
		{
			name: "image dashboard",
			dash: dashboard.Dashboard{ID: "abc123", Type: "image", ShowEditButton: true},
			base: base,
			want: "",
		},
		{
			name: "no editor public url",
			dash: dashboard.Dashboard{ID: "abc123", Type: "interactive", ShowEditButton: true},
			base: "",
			want: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := editLink(tc.dash, tc.base); got != tc.want {
				t.Errorf("editLink() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestInteractiveDashboard_EditButton(t *testing.T) {
	dash := imageDashboard("test")
	dash.Type = "interactive"
	dash.ShowEditButton = true
	handler := newTestMiddlewareWithEditor(t, "https://editor.example", dash)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if want := `href="https://editor.example/dashboards/test/edit"`; !strings.Contains(rec.Body.String(), want) {
		t.Errorf("expected edit button %s, body: %s", want, rec.Body.String())
	}
}

func TestImageDashboard_BrowserPreviewHasNoEditButton(t *testing.T) {
	dash := imageDashboard("test")
	dash.ShowEditButton = true
	handler := newTestMiddlewareWithEditor(t, "https://editor.example", dash)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "Edit dashboard") || strings.Contains(body, "/dashboards/test/edit") {
		t.Errorf("image dashboard must not get an edit button, body: %s", body)
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
			Rows:            []dashboard.Row{{ID: "r1", Height: "auto", Width: "100%", Widgets: []dashboard.Widget{{ID: "w1", Type: "test", Width: 12, Config: json.RawMessage(`{}`)}}}}},
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
			Rows:            []dashboard.Row{{ID: "r1", Height: "auto", Width: "100%", Widgets: []dashboard.Widget{{ID: "w1", Type: "test", Width: 12, Config: json.RawMessage(`{}`)}}}}},
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

func TestParseDisplayHeadersSizeBounds(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		headers map[string]string
		wantErr string // "" = accepted
	}{
		{name: "width at max via query", query: "?format=png&width=4096&height=480"},
		{name: "height at max via query", query: "?format=png&width=800&height=4096"},
		{name: "both at max via headers", headers: map[string]string{"X-Display-Width": "4096", "X-Display-Height": "4096"}},
		{name: "width 1 is the minimum", query: "?format=png&width=1&height=1"},
		{name: "width over max via query", query: "?format=png&width=4097&height=480", wantErr: "invalid X-Display-Width: 4097 (must be 1 to 4096)"},
		{name: "height over max via query", query: "?format=png&width=800&height=4097", wantErr: "invalid X-Display-Height: 4097 (must be 1 to 4096)"},
		{name: "width over max via header", headers: map[string]string{"X-Display-Width": "4097", "X-Display-Height": "480"}, wantErr: "invalid X-Display-Width: 4097 (must be 1 to 4096)"},
		{name: "height over max via header", headers: map[string]string{"X-Display-Width": "800", "X-Display-Height": "4097"}, wantErr: "invalid X-Display-Height: 4097 (must be 1 to 4096)"},
		{name: "header over max wins over valid query", query: "?format=png&width=800&height=480", headers: map[string]string{"X-Display-Width": "100000"}, wantErr: "invalid X-Display-Width: 100000 (must be 1 to 4096)"},
		{name: "zero width", query: "?format=png&width=0&height=480", wantErr: "invalid X-Display-Width: 0 (must be 1 to 4096)"},
		{name: "negative height", query: "?format=png&width=800&height=-5", wantErr: "invalid X-Display-Height: -5 (must be 1 to 4096)"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test"+tc.query, nil)
			if tc.query == "" {
				req.Header.Set("X-Display-Format", "png")
			}
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			_, err := parseDisplayHeaders(req)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

// An oversized display request must be refused before anything is rendered,
// with the limit in the 400 body so a client can tell what to fix.
func TestImageDashboard_OversizedDisplayAnswers400WithLimit(t *testing.T) {
	handler := newTestMiddleware(t, imageDashboard("test"))

	req := httptest.NewRequest("GET", "/test?format=png&width=800&height=100000", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("must be 1 to 4096")) {
		t.Errorf("body %q does not name the limit", rec.Body)
	}
}

func TestResolveImageBackgroundWithGradientOnly(t *testing.T) {
	// When the background has a gradient but no image, resolveImageBackground
	// should return the gradient CSS and nil bytes.
	themeStore := themes.NewStore(t.TempDir())
	theme := themes.Theme{
		ID:         "gradientonly",
		Name:       "Gradient Only",
		Colors:     themes.DefaultTheme().Colors,
		Background: themes.Background{Light: themes.ModeBackground{Gradient: &themes.Gradient{Direction: "to right", Stops: []string{"#000000", "#ffffff"}}}},
	}

	css, imageData := resolveImageBackground(theme, "light", themeStore)

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
	themeStore := themes.NewStore(t.TempDir())

	theme := themes.Theme{
		ID:     "withimage",
		Name:   "With Image",
		Colors: themes.DefaultTheme().Colors,
		Background: themes.Background{Light: themes.ModeBackground{Image: &themes.Image{
			Ref: "asset:test.png", Fit: "cover", Position: "center", Repeat: "no-repeat",
		}}},
	}
	if _, err := themeStore.Create(theme); err != nil {
		t.Fatal(err)
	}

	// Save a small test image into the theme's own assets
	testPNG := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG signature
	if err := themeStore.SaveAsset(theme.ID, "test.png", testPNG); err != nil {
		t.Fatal(err)
	}

	css, imageData := resolveImageBackground(theme, "light", themeStore)

	if css != "transparent" {
		t.Errorf("expected 'transparent' css to prevent theme-background paint-over, got %q", css)
	}
	if imageData == nil {
		t.Error("expected non-nil image data")
	} else if len(imageData) != len(testPNG) {
		t.Errorf("expected %d bytes, got %d", len(testPNG), len(imageData))
	}
}

// Review Focus 5: litehtml cannot render alpha, so a translucent surface must
// reach the image stack as an opaque 6-digit colour composited over bg.
func TestBuildRenderDataFlattensPalette(t *testing.T) {
	light := themes.DefaultPalette("light")
	light.BG = "#000000"
	light.Surface = "#ffffff80"
	theme := themes.Theme{
		ID:     "translucent",
		Name:   "Translucent",
		Colors: themes.Colors{Light: light, Dark: themes.DefaultPalette("dark")},
	}
	dash := imageDashboard("flat")
	dash.ColorMode = "light"

	got := buildRenderData(dash, 0, nil, dashboard.NewStore(t.TempDir()), theme, widgetthemes.NewStore(""))

	if got.Palette.Surface != "#808080" {
		t.Errorf("Palette.Surface = %q, want #808080 (#ffffff80 over #000000)", got.Palette.Surface)
	}
	if got.Palette.BG != "#000000" {
		t.Errorf("Palette.BG = %q, want #000000", got.Palette.BG)
	}
}

func TestBuildRenderDataCarriesWidgetThemeAndFont(t *testing.T) {
	dash := imageDashboard("wt")
	dash.WidgetTheme = "default"

	got := buildRenderData(dash, 0, nil, dashboard.NewStore(t.TempDir()), themes.DefaultTheme(), widgetthemes.NewStore(""))

	if got.WidgetTheme != "default" {
		t.Errorf("WidgetTheme = %q, want default", got.WidgetTheme)
	}
	if got.FontFamily != "Inter" {
		t.Errorf("FontFamily = %q, want Inter", got.FontFamily)
	}
}

func TestImageDashboardWithDanglingBackgroundRefStillRenders(t *testing.T) {
	// Create an image dashboard with a ThemeID that references nothing.
	dash := imageDashboard("dangling-bg-test")
	dash.ThemeID = "nonexistent-bg-id"

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
	// pre-composited image. This test proves the full path: themeStore.LoadImage
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

	dashStore := dashboard.NewStore(t.TempDir())
	themeStore := themes.NewStore(t.TempDir())

	// Create a theme whose background references its own red asset
	bg := themes.Theme{
		ID:     "redbg",
		Name:   "Red Background",
		Colors: themes.DefaultTheme().Colors,
		Background: themes.Background{Light: themes.ModeBackground{Image: &themes.Image{
			Ref:      "asset:red.png",
			Fit:      "cover",
			Position: "center",
			Repeat:   "no-repeat",
		}}},
	}
	if _, err := themeStore.Create(bg); err != nil {
		t.Fatal(err)
	}
	if err := themeStore.SaveAsset(bg.ID, "red.png", redBuf.Bytes()); err != nil {
		t.Fatal(err)
	}

	// Create an image dashboard with real content and the background
	dash := dashboard.Dashboard{
		ID:      "testredbg",
		Name:    "Red BG Test",
		Type:    "image",
		ThemeID: "redbg",
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
	mid := NewDashboardMiddleware(dashStore, browserRenderer, staticRenderer, imageRenderer, themeStore, widgetthemes.NewStore(""), "")
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
