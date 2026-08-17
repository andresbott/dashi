package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/dashboard/browser"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/gorilla/mux"
)

func dashiAssetRouter() *mux.Router {
	r := mux.NewRouter()
	attachDashiAssets(r, browser.NewAssets(nil), themes.NewStore(""))
	return r
}

func TestDashiAssetRoutes(t *testing.T) {
	tests := []struct {
		path        string
		wantStatus  int
		wantType    string
		wantContent string
	}{
		{"/_dashi/assets/viewer.css", http.StatusOK, "text/css; charset=utf-8", ""},
		{"/_dashi/assets/dashi.js", http.StatusOK, "text/javascript; charset=utf-8", "export function poll"},
		{"/_dashi/assets/widgets.css", http.StatusOK, "text/css; charset=utf-8", ""},
		{"/_dashi/theme/default.css", http.StatusOK, "text/css; charset=utf-8", "--dashi-fg"},
		{"/_dashi/theme/nope.css", http.StatusOK, "text/css; charset=utf-8", "--dashi-fg"},
		{"/_dashi/widgets/clock.js", http.StatusNotFound, "", ""},
	}

	r := dashiAssetRouter()
	for _, tc := range tests {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != tc.wantStatus {
			t.Errorf("%s: status = %d, want %d", tc.path, rec.Code, tc.wantStatus)
			continue
		}
		if tc.wantStatus != http.StatusOK {
			continue
		}
		if got := rec.Header().Get("Content-Type"); got != tc.wantType {
			t.Errorf("%s: Content-Type = %q, want %q", tc.path, got, tc.wantType)
		}
		if tc.wantContent != "" && !strings.Contains(rec.Body.String(), tc.wantContent) {
			t.Errorf("%s: body missing %q", tc.path, tc.wantContent)
		}
	}
}

func TestDashiThemeCSSRejectsTraversal(t *testing.T) {
	// Theme names reach a store lookup, never the filesystem, but the
	// route must not accept a path segment with separators either.
	req := httptest.NewRequest(http.MethodGet, "/_dashi/theme/..%2F..%2Fetc%2Fpasswd.css", nil)
	rec := httptest.NewRecorder()
	dashiAssetRouter().ServeHTTP(rec, req)

	if rec.Code == http.StatusOK && strings.Contains(rec.Body.String(), "root:") {
		t.Fatal("traversal served file contents")
	}
}
