package router

import (
	"bytes"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/dashboard"
	"golang.org/x/crypto/bcrypt"
)

// The editor's e-ink preview loads GET /{id}?format=…&width=…&height=… from
// the editor server itself. That must render the PNG from query params alone
// (an SPA fetch sends no display headers) and must not be blocked by the
// viewer's per-dashboard basic auth, which the editor does not mount.
func TestEditorServesEinkPreviewPNG(t *testing.T) {
	viewer, ids := newTestViewerWithDashboards(t)
	editor, err := NewEditorFromDeps(sharedTestCfg, sharedTestDeps)
	if err != nil {
		t.Fatalf("NewEditorFromDeps: %v", err)
	}

	// A protected image dashboard of its own, so the shared fixtures stay
	// unprotected for the other tests.
	protected, err := sharedTestDeps.dashStore.Create(dashboard.Dashboard{
		Name: "Protected Image", Type: "image", Dither: "grayscale",
		Pages: []dashboard.Page{{Name: "Main"}, {Name: "Second"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	if err := sharedTestDeps.dashStore.SetAuth(protected.ID, "viewer", string(hash)); err != nil {
		t.Fatal(err)
	}

	get := func(h http.Handler, path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	query := "?format=png-bw&width=320&height=200&rotation=90&page=1"

	t.Run("editor renders the PNG from query params, ignoring viewer auth", func(t *testing.T) {
		rec := get(editor, "/"+protected.ID+query)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
			t.Fatalf("Content-Type = %q, want image/png", ct)
		}
		img, err := png.Decode(bytes.NewReader(rec.Body.Bytes()))
		if err != nil {
			t.Fatalf("not a PNG: %v", err)
		}
		// Native panel size, whatever the rotation.
		if b := img.Bounds(); b.Dx() != 320 || b.Dy() != 200 {
			t.Errorf("image is %dx%d, want 320x200", b.Dx(), b.Dy())
		}
	})

	t.Run("viewer still requires the dashboard's credentials", func(t *testing.T) {
		if rec := get(viewer, "/"+protected.ID+query); rec.Code != http.StatusUnauthorized {
			t.Errorf("viewer status = %d, want 401", rec.Code)
		}
	})

	t.Run("bad display params answer 400 with the reason", func(t *testing.T) {
		rec := get(editor, "/"+ids["image"]+"?format=png-bw&width=0&height=200")
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "invalid X-Display-Width") {
			t.Errorf("body %q does not name the bad param", rec.Body)
		}
	})

	// Interactive dashboards have no PNG path: the display request falls
	// through to browser HTML, which is why the editor hides the button.
	t.Run("interactive dashboard answers HTML", func(t *testing.T) {
		rec := get(editor, "/"+ids["plain"]+"?format=png-bw&width=320&height=200")
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
			t.Errorf("Content-Type = %q, want text/html", ct)
		}
	})
}
