package router

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/dashboard/browser"
	dashimage "github.com/andresbott/dashi/internal/dashboard/image"
	dashstatic "github.com/andresbott/dashi/internal/dashboard/static"
	"github.com/andresbott/dashi/internal/data/images"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/andresbott/dashi/internal/widgetthemes"
)

// The editor refuses a dither mode the renderer does not know, on create
// and update, with a 400 that names the bad value.
func TestEditorRejectsUnknownDither(t *testing.T) {
	_, ids := newTestViewerWithDashboards(t)
	editor, err := NewEditorFromDeps(sharedTestCfg, sharedTestDeps)
	if err != nil {
		t.Fatalf("NewEditorFromDeps: %v", err)
	}
	body := `{"name":"Dithered","type":"image","dither":"sparkly","pages":[{"name":"Main","rows":[]}]}`

	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, "/api/v0/dashboards"},
		{http.MethodPut, "/api/v0/dashboards/" + ids["image"]},
	} {
		t.Run(tc.method, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			editor.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body: %s", rec.Code, rec.Body)
			}
			if msg := errorMessage(t, rec.Body.Bytes()); !strings.Contains(msg, `invalid dither mode "sparkly"`) {
				t.Errorf("error %q does not name the bad mode", msg)
			}
		})
	}

	got, err := sharedTestDeps.dashStore.Get(ids["image"])
	if err != nil {
		t.Fatal(err)
	}
	if got.Dither != "" || got.Name != "Image" {
		t.Errorf("rejected update changed the dashboard: %+v", got)
	}
}

// newDitherMiddleware serves image dashboards whose single widget paints
// flat mid-tones and colours, which a processing preset maps differently
// from plain Floyd-Steinberg.
func newDitherMiddleware(t *testing.T, dashboards ...dashboard.Dashboard) http.Handler {
	t.Helper()
	store := dashboard.NewStore(t.TempDir())
	for _, d := range dashboards {
		if _, err := store.Create(d); err != nil {
			t.Fatalf("create dashboard: %v", err)
		}
	}
	reg := widgets.NewRegistry()
	reg.Register("swatch", func(json.RawMessage, widgets.RenderContext) (template.HTML, error) {
		return `<div style="background:#7f7f7f;height:30px"></div>` +
			`<div style="background:#3a6ea5;height:30px"></div>` +
			`<div style="background:#c08040;height:30px"></div>`, nil
	})
	bs, _ := images.NewStore(t.TempDir())
	mid := NewDashboardMiddleware(store,
		browser.NewRenderer(reg, browser.NewAssets(nil).JSTypes()),
		dashstatic.NewRenderer(reg), dashimage.NewRenderer(),
		themes.NewStore(t.TempDir(), bs), widgetthemes.NewStore(""))
	return mid(http.NotFoundHandler())
}

func ditherDashboard(id, dither string) dashboard.Dashboard {
	d := imageDashboard(id, dashboard.Page{Name: "Main", Rows: []dashboard.Row{{
		ID: "r1", Widgets: []dashboard.Widget{{ID: "w1", Type: "swatch", Width: 12, Config: json.RawMessage(`{}`)}},
	}}})
	d.Dither = dither
	return d
}

// The dashboard's dither field reaches every dithered format: a preset
// changes the bytes, plain png is untouched by it.
func TestImageDashboard_DitherFieldReachesEncoder(t *testing.T) {
	handler := newDitherMiddleware(t,
		ditherDashboard("plain", ""),
		ditherDashboard("gray", "grayscale"),
		ditherDashboard("vivid", "vivid"),
		ditherDashboard("auto", "auto"),
	)
	fetch := func(id, format string) []byte {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/"+id+"?format="+format+"&width=120&height=90", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s %s: status %d: %s", id, format, rec.Code, rec.Body)
		}
		return rec.Body.Bytes()
	}

	for _, format := range []string{"png-bw", "png-spectra6", "bw", "spectra6"} {
		t.Run(format, func(t *testing.T) {
			plain := fetch("plain", format)
			preset := "gray"
			if strings.Contains(format, "spectra6") {
				preset = "vivid"
			}
			if bytes.Equal(plain, fetch(preset, format)) {
				t.Errorf("%s dashboard dithered like plain", preset)
			}
			fetch("auto", format) // auto must render without error
		})
	}

	t.Run("png ignores dither", func(t *testing.T) {
		if !bytes.Equal(fetch("plain", "png"), fetch("gray", "png")) {
			t.Error("undithered png changed with the dither mode")
		}
	})
}
