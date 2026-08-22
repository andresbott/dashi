package router

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/andresbott/dashi/internal/backgrounds"
	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/data/images"
)

var (
	sharedHandler http.Handler
	sharedIDs     map[string]string
	sharedOnce    sync.Once
	sharedTempDir string
)

// newTestViewerWithDashboards returns a viewer handler with two dashboards
// (one plain, one image). The handler and dashboards are built once per test
// binary and reused across all tests to avoid prometheus metrics collision.
// All four tests using this helper are read-only GETs, so sharing is safe.
func newTestViewerWithDashboards(t *testing.T) (http.Handler, map[string]string) {
	t.Helper()

	sharedOnce.Do(func() {
		// Create temp dir that persists for the test binary's lifetime
		dir, err := os.MkdirTemp("", "dashi-test-*")
		if err != nil {
			panic("failed to create temp dir: " + err.Error())
		}
		sharedTempDir = dir

		store := dashboard.NewStore(filepath.Join(dir, "dashboards"))

		// Create background entities using the backgrounds Store so they are
		// properly indexed and persisted.
		sharedBgImages, err := images.NewStore(filepath.Join(dir, "data", "shared-background-images"))
		if err != nil {
			panic("create shared bg images store: " + err.Error())
		}
		tempBgStore := backgrounds.NewStore(filepath.Join(dir, "data", "backgrounds"), sharedBgImages)

		themeImageBG, err := tempBgStore.Create(backgrounds.Background{
			Name: "Shared Image",
			Image: &backgrounds.Image{
				Light:    "shared:bg.jpg",
				Fit:      "cover",
				Position: "center",
				Repeat:   "no-repeat",
			},
		})
		if err != nil {
			panic("create theme-image-bg: " + err.Error())
		}

		gradientBG, err := tempBgStore.Create(backgrounds.Background{
			Name: "Gradient",
			Gradient: &backgrounds.Gradient{
				Direction: "to bottom",
				Light:     []string{"#fff", "#000"},
			},
		})
		if err != nil {
			panic("create gradient-bg: " + err.Error())
		}

		page := []dashboard.Page{{
			Name: "Main",
			Rows: []dashboard.Row{{
				ID: "r1",
				Widgets: []dashboard.Widget{{
					ID: "w1", Type: "clock", Width: 12,
					Config: json.RawMessage(`{"showDate":true}`),
				}},
			}},
		}}

		plain, err := store.Create(dashboard.Dashboard{
			Name: "Plain", Default: true, Pages: page,
		})
		if err != nil {
			panic("create plain dashboard: " + err.Error())
		}
		image, err := store.Create(dashboard.Dashboard{
			Name: "Image", Type: "image", Pages: page,
		})
		if err != nil {
			panic("create image dashboard: " + err.Error())
		}
		imageBG, err := store.Create(dashboard.Dashboard{
			Name:         "ImageBG",
			Pages:        page,
			BackgroundID: themeImageBG.ID,
		})
		if err != nil {
			panic("create image-background dashboard: " + err.Error())
		}
		gradientBGDash, err := store.Create(dashboard.Dashboard{
			Name:         "GradientBG",
			Pages:        page,
			BackgroundID: gradientBG.ID,
		})
		if err != nil {
			panic("create gradient-background dashboard: " + err.Error())
		}
		multi, err := store.Create(dashboard.Dashboard{
			Name: "Multi",
			Pages: []dashboard.Page{
				{Name: "Overview", Rows: page[0].Rows},
				{Name: "Metrics", Rows: page[0].Rows},
				{Name: "", Rows: page[0].Rows},
			},
		})
		if err != nil {
			panic("create multi-page dashboard: " + err.Error())
		}

		cfg := Cfg{DataDir: dir, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
		deps, err := newSharedDeps(cfg)
		if err != nil {
			panic("newSharedDeps: " + err.Error())
		}
		h, err := NewViewerFromDeps(cfg, deps)
		if err != nil {
			panic("NewViewerFromDeps: " + err.Error())
		}

		sharedHandler = h
		sharedIDs = map[string]string{
			"plain":      plain.ID,
			"image":      image.ID,
			"imageBG":    imageBG.ID,
			"gradientBG": gradientBGDash.ID,
			"multi":      multi.ID,
		}
	})

	// The temp directory persists for the test binary's lifetime. The OS
	// will clean it up when the process exits. We do not use t.Cleanup here
	// because it would delete the shared directory after the first test
	// completes, breaking subsequent tests that rely on the same handler.

	return sharedHandler, sharedIDs
}

func TestBrowserDashboardIsServerRendered(t *testing.T) {
	h, ids := newTestViewerWithDashboards(t)

	req := httptest.NewRequest(http.MethodGet, "/"+ids["plain"], nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "/_dashi/assets/viewer.css") {
		t.Errorf("expected server-rendered page, got:\n%s", body)
	}
	if strings.Contains(body, "<div id=\"app\">") {
		t.Error("viewer served the Vue SPA instead of server-rendered HTML")
	}
}

func TestImageDashboardWithoutDisplayHeadersIsServerRenderedHTML(t *testing.T) {
	// The old serveImageHTMLPreview path is gone: an image dashboard
	// opened in a browser now gets the same browser stack as any other.
	h, ids := newTestViewerWithDashboards(t)

	req := httptest.NewRequest(http.MethodGet, "/"+ids["image"], nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "/_dashi/assets/viewer.css") {
		t.Error("image dashboard should render through the browser stack in a browser")
	}
}

func TestImageDashboardWithDisplayHeadersStillRendersPNG(t *testing.T) {
	h, ids := newTestViewerWithDashboards(t)

	req := httptest.NewRequest(http.MethodGet, "/"+ids["image"], nil)
	req.Header.Set("X-Display-Format", "png")
	req.Header.Set("X-Display-Width", "200")
	req.Header.Set("X-Display-Height", "100")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", got)
	}
}

func TestRootRedirectsToDefaultDashboard(t *testing.T) {
	h, ids := newTestViewerWithDashboards(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "/"+ids["plain"] {
		t.Errorf("Location = %q, want /%s", got, ids["plain"])
	}
}

func TestBrowserImageBackgroundIsAURLNotInlinedBytes(t *testing.T) {
	// A 2 MB background used to be base64-inlined into the <html> style
	// attribute of every single response, uncacheable — and, because the
	// shell compiled it to background-color, invisible anyway. Now it's
	// a <style> block with a URL pointing at the asset route.
	h, ids := newTestViewerWithDashboards(t)

	req := httptest.NewRequest(http.MethodGet, "/"+ids["imageBG"], nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()

	if strings.Contains(body, "base64,") || strings.Contains(body, "data:image") {
		t.Errorf("browser page inlined the background image bytes:\n%s", truncate(body))
	}
	// Background now lives in a <style> block, not an inline attribute.
	if strings.Contains(body, `style="--dashi-page-bg`) {
		t.Errorf("background leaked into an inline style attribute:\n%s", truncate(body))
	}
	// The URL appears in a <style> block. BrowserCSS emits both light and dark.
	// browserValue builds the shorthand with position/size/repeat.
	// The image is from the shared pool, so it references /api/v0/data/backgrounds/{name}.
	if !strings.Contains(body, `url('/api/v0/data/backgrounds/bg.jpg') center/cover no-repeat`) {
		t.Errorf("expected a background URL in the page:\n%s", truncate(body))
	}
	// We don't test the route here because the image file doesn't actually exist
	// in the shared pool - we only created the background entity, not the image file itself.
}

func TestBrowserGradientBackgroundReachesThePage(t *testing.T) {
	h, ids := newTestViewerWithDashboards(t)

	req := httptest.NewRequest(http.MethodGet, "/"+ids["gradientBG"], nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	body := rec.Body.String()
	// Background now lives in a <style> block. BrowserCSS always emits both
	// light and dark rules (gradient has no dark variant, so both are the same).
	// Note: backgrounds.resolve joins stops with "," (no space after comma).
	if !strings.Contains(body, "--dashi-page-bg:linear-gradient(to bottom,#fff,#000);") {
		t.Errorf("gradient background missing from the page:\n%s", truncate(body))
	}
	// Background must not be in an inline attribute.
	if strings.Contains(body, `style="--dashi-page-bg`) {
		t.Errorf("background leaked into an inline style attribute:\n%s", truncate(body))
	}
	// A background-color utility would silently discard a gradient.
	if strings.Contains(body, "bg-[var(--dashi-page-bg") {
		t.Errorf("page still uses a background-color utility for the page background:\n%s", truncate(body))
	}
}

func TestBrowserMultiPageDashboardRendersPageLinks(t *testing.T) {
	h, ids := newTestViewerWithDashboards(t)

	req := httptest.NewRequest(http.MethodGet, "/"+ids["multi"]+"?page=1", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`href="?page=0"`, `href="?page=1"`, `href="?page=2"`,
		`>Overview<`, `>Metrics<`, `>Page 3<`,
		`aria-current="page"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("page nav missing %q:\n%s", want, truncate(body))
		}
	}
}

func TestBrowserSinglePageDashboardHasNoPageNav(t *testing.T) {
	h, ids := newTestViewerWithDashboards(t)

	req := httptest.NewRequest(http.MethodGet, "/"+ids["plain"], nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if strings.Contains(rec.Body.String(), "dashi-pages") {
		t.Error("single-page dashboard should render no page navigation")
	}
}

func truncate(s string) string {
	const max = 2000
	if len(s) <= max {
		return s
	}
	return s[:max] + "…(truncated)"
}
