package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// A missing module route would not 404 on the editor: unknown GETs fall through
// to the SPA and answer 200. The 400 proves the weather handler itself replied.
func TestWidgetRoutesServedOnViewerAndEditor(t *testing.T) {
	viewer, _ := newTestViewerWithDashboards(t)
	editor, err := NewEditorFromDeps(sharedTestCfg, sharedTestDeps)
	if err != nil {
		t.Fatalf("NewEditorFromDeps: %v", err)
	}

	servers := map[string]http.Handler{"viewer": viewer, "editor": editor}
	for name, h := range servers {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v0/widgets/weather", nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}
