package router

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// End to end through the editor: newSharedDeps must register each module's
// ConfigValidator and hand the registry to the dashboard store, so an unsafe
// bookmark URL — top-level or nested in a stack — is refused with a 400 on
// create, update and zip import. Every request here is rejected, so the
// shared fixture's store is never modified.
func TestEditorRejectsUnsafeBookmarkURL(t *testing.T) {
	_, ids := newTestViewerWithDashboards(t)
	editor, err := NewEditorFromDeps(sharedTestCfg, sharedTestDeps)
	if err != nil {
		t.Fatalf("NewEditorFromDeps: %v", err)
	}

	bookmark := `{"id":"b1","type":"bookmark","width":3,"config":{"url":"javascript:alert(1)","title":"x"}}`
	stacked := `{"id":"s1","type":"stack","width":3,"config":{"widgets":[` +
		`{"id":"b2","type":"bookmark","config":{"url":"//evil.example","title":"x"}}]}}`
	dashJSON := func(widget string) string {
		return `{"name":"XSS","pages":[{"name":"Main","rows":[{"id":"r1","widgets":[` + widget + `]}]}]}`
	}

	tests := []struct {
		name    string
		method  string
		path    string
		body    []byte
		ctype   string
		wantMsg string
	}{
		{
			name: "create", method: http.MethodPost, path: "/api/v0/dashboards",
			body: []byte(dashJSON(bookmark)), ctype: "application/json",
			wantMsg: `bookmark widget "b1"`,
		},
		{
			name: "create with stacked bookmark", method: http.MethodPost, path: "/api/v0/dashboards",
			body: []byte(dashJSON(stacked)), ctype: "application/json",
			wantMsg: `stack child "b2" (bookmark)`,
		},
		{
			name: "update", method: http.MethodPut, path: "/api/v0/dashboards/" + ids["plain"],
			body: []byte(dashJSON(bookmark)), ctype: "application/json",
			wantMsg: `bookmark widget "b1"`,
		},
		{
			name: "zip import", method: http.MethodPost, path: "/api/v0/dashboards/upload",
			body: zipWithDashboard(t, dashJSON(bookmark)), ctype: "application/zip",
			wantMsg: `bookmark widget "b1"`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewReader(tc.body))
			req.Header.Set("Content-Type", tc.ctype)
			rec := httptest.NewRecorder()
			editor.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body: %s", rec.Code, rec.Body)
			}
			msg := errorMessage(t, rec.Body.Bytes())
			for _, want := range []string{tc.wantMsg, "url must start with http://, https:// or /"} {
				if !strings.Contains(msg, want) {
					t.Errorf("error %q does not mention %q", msg, want)
				}
			}
		})
	}

	// The rejected update left the stored dashboard alone.
	got, err := sharedTestDeps.dashStore.Get(ids["plain"])
	if err != nil {
		t.Fatalf("get plain: %v", err)
	}
	if got.Name != "Plain" {
		t.Errorf("plain dashboard name = %q after a rejected update, want %q", got.Name, "Plain")
	}
}

// errorMessage extracts the handler's message from an error response. The
// go-bumbu middleware (JsonErrors) wraps the handler's whole JSON body as
// the string value of its own {"error": ...}, so the message sits one level
// down.
func errorMessage(t *testing.T, body []byte) string {
	t.Helper()
	var outer struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &outer); err != nil {
		t.Fatalf("decode error body %q: %v", body, err)
	}
	var inner struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(outer.Error), &inner); err == nil && inner.Error != "" {
		return inner.Error
	}
	return outer.Error
}

func zipWithDashboard(t *testing.T, dashJSON string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	f, err := zw.Create("dashboard.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte(dashJSON)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
