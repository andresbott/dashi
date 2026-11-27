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

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/data/images"
	"github.com/andresbott/dashi/internal/themes"
)

var (
	sharedHandler  http.Handler
	sharedIDs      map[string]string
	sharedOnce     sync.Once
	sharedTempDir  string
	sharedTestCfg  Cfg
	sharedTestDeps *sharedDeps
)

// newTestViewerWithDashboards returns a viewer handler with a set of fixture
// dashboards (plain, image, themed, multi-page, search, ...) keyed by name in
// the returned map. The handler and dashboards are built once per test binary
// and reused across all tests to avoid prometheus metrics collision. Every
// test using this helper only issues read-only GETs, so sharing is safe.
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

		// Create theme entities using the themes Store so they are properly
		// indexed and persisted where newSharedDeps will look for them.
		sharedBgImages, err := images.NewStore(filepath.Join(dir, "data", "shared-background-images"))
		if err != nil {
			panic("create shared bg images store: " + err.Error())
		}
		tempThemeStore := themes.NewStore(filepath.Join(dir, "data", "themes"), sharedBgImages)

		themeImageBG, err := tempThemeStore.Create(themes.Theme{
			Name:   "Shared Image",
			Colors: themes.DefaultTheme().Colors,
			Background: themes.Background{Image: &themes.Image{
				Light:    "shared:bg.jpg",
				Fit:      "cover",
				Position: "center",
				Repeat:   "no-repeat",
			}},
		})
		if err != nil {
			panic("create theme-image-bg: " + err.Error())
		}

		gradientBG, err := tempThemeStore.Create(themes.Theme{
			Name:   "Gradient",
			Colors: themes.DefaultTheme().Colors,
			Background: themes.Background{Gradient: &themes.Gradient{
				Direction: "to bottom",
				Light:     []string{"#fff", "#000"},
			}},
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
			Name:    "ImageBG",
			Pages:   page,
			ThemeID: themeImageBG.ID,
		})
		if err != nil {
			panic("create image-background dashboard: " + err.Error())
		}
		gradientBGDash, err := store.Create(dashboard.Dashboard{
			Name:    "GradientBG",
			Pages:   page,
			ThemeID: gradientBG.ID,
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
		// References are not checked on save: the links carry the stored
		// values verbatim, and the asset routes resolve them.
		themed, err := store.Create(dashboard.Dashboard{
			Name: "Themed", Pages: page,
			ThemeID: "abc123", WidgetTheme: "mono",
		})
		if err != nil {
			panic("create themed dashboard: " + err.Error())
		}
		dangling, err := store.Create(dashboard.Dashboard{
			Name: "Dangling", Pages: page,
			ThemeID: "gone99", WidgetTheme: "nope",
		})
		if err != nil {
			panic("create dangling dashboard: " + err.Error())
		}

		search, err := store.Create(dashboard.Dashboard{
			Name: "Search",
			Pages: []dashboard.Page{{
				Name: "Main",
				Rows: []dashboard.Row{{
					ID: "r1",
					Widgets: []dashboard.Widget{{
						ID: "w1", Type: "search", Width: 12,
						Config: json.RawMessage(`{"engine":"duckduckgo","placeholder":"Find"}`),
					}},
				}},
			}},
		})
		if err != nil {
			panic("create search dashboard: " + err.Error())
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
		sharedTestCfg = cfg
		sharedTestDeps = deps
		sharedIDs = map[string]string{
			"plain":      plain.ID,
			"image":      image.ID,
			"imageBG":    imageBG.ID,
			"gradientBG": gradientBGDash.ID,
			"multi":      multi.ID,
			"themed":     themed.ID,
			"dangling":   dangling.ID,
			"search":     search.ID,
			// Theme entity IDs, for fetching the linked stylesheets.
			"imageTheme":    themeImageBG.ID,
			"gradientTheme": gradientBG.ID,
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
	// shell compiled it to background-color, invisible anyway. Now the page
	// links the theme stylesheet, which carries a URL to the asset route.
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
	if strings.Contains(body, "--dashi-page-bg") {
		t.Errorf("background leaked into the page instead of the theme stylesheet:\n%s", truncate(body))
	}
	link := "/_dashi/theme/" + ids["imageTheme"] + ".css"
	if !strings.Contains(body, `href="`+link+`"`) {
		t.Fatalf("page does not link %s:\n%s", link, truncate(body))
	}

	cssRec := httptest.NewRecorder()
	h.ServeHTTP(cssRec, httptest.NewRequest(http.MethodGet, link, nil))
	// browserValue builds the shorthand with position/size/repeat. The image
	// is from the shared pool, so it references /api/v0/data/backgrounds/{name}.
	if css := cssRec.Body.String(); !strings.Contains(css, `url('/api/v0/data/backgrounds/bg.jpg') center/cover no-repeat`) {
		t.Errorf("expected a background URL in the theme stylesheet:\n%s", css)
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
	link := "/_dashi/theme/" + ids["gradientTheme"] + ".css"
	if !strings.Contains(body, `href="`+link+`"`) {
		t.Fatalf("page does not link %s:\n%s", link, truncate(body))
	}
	cssRec := httptest.NewRecorder()
	h.ServeHTTP(cssRec, httptest.NewRequest(http.MethodGet, link, nil))
	// Note: themes.resolve joins stops with "," (no space after comma).
	if css := cssRec.Body.String(); !strings.Contains(css, "--dashi-page-bg:linear-gradient(to bottom,#fff,#000);") {
		t.Errorf("gradient background missing from the theme stylesheet:\n%s", css)
	}
	// Background must not be inlined into the page.
	if strings.Contains(body, "--dashi-page-bg:") {
		t.Errorf("background leaked into the page:\n%s", truncate(body))
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

func TestBrowserDashboardLinksBothStylesheets(t *testing.T) {
	h, ids := newTestViewerWithDashboards(t)

	req := httptest.NewRequest(http.MethodGet, "/"+ids["themed"], nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`href="/_dashi/theme/abc123.css"`,
		`href="/_dashi/widget-theme/mono.css"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("page missing %q:\n%s", want, truncate(body))
		}
	}
}

// Review Focus 2: a dashboard referencing a deleted theme and a widget theme
// that never existed still renders, and both stylesheets it links resolve.
func TestBrowserDashboardDanglingRefsUseDefaults(t *testing.T) {
	h, ids := newTestViewerWithDashboards(t)

	get := func(path string) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec
	}

	if rec := get("/" + ids["dangling"]); rec.Code != http.StatusOK {
		t.Fatalf("dashboard status = %d, want 200", rec.Code)
	}

	themeCSS := get("/_dashi/theme/gone99.css")
	if themeCSS.Code != http.StatusOK {
		t.Fatalf("theme css status = %d, want 200", themeCSS.Code)
	}
	if body := themeCSS.Body.String(); !strings.Contains(body, "--dashi-bg:#ffffff;") {
		t.Errorf("dangling theme must serve the default palette, got:\n%s", body)
	}

	if rec := get("/_dashi/widget-theme/nope.css"); rec.Code != http.StatusOK {
		t.Errorf("widget theme css status = %d, want 200", rec.Code)
	}
}

// The real viewer wiring (module list in main.go, page.html) renders the
// search widget as a working GET form, not the old empty image placeholder.
func TestBrowserSearchWidgetRendersAForm(t *testing.T) {
	h, ids := newTestViewerWithDashboards(t)

	req := httptest.NewRequest(http.MethodGet, "/"+ids["search"], nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`<form`,
		`action="https://duckduckgo.com/"`,
		`name="q"`,
		`placeholder="Find"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("search widget missing %q:\n%s", want, truncate(body))
		}
	}
	// "search-placeholder" matches both the image stack's fake input
	// (image.html) and the old empty <div class="widget-search-placeholder">,
	// so it fails if the browser path falls back to either.
	if strings.Contains(body, "search-placeholder") {
		t.Errorf("search widget renders the image-stack placeholder, not the browser form:\n%s", truncate(body))
	}
}

func truncate(s string) string {
	const max = 2000
	if len(s) <= max {
		return s
	}
	return s[:max] + "…(truncated)"
}
