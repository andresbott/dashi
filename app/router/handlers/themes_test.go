package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/gorilla/mux"
)

func newThemeTestRouter(t *testing.T) (*mux.Router, *themes.Store, *dashboard.Store) {
	t.Helper()
	dir := t.TempDir()
	themeStore := themes.NewStore(filepath.Join(dir, "themes"))
	dashStore := dashboard.NewStore(filepath.Join(dir, "dashboards"))

	h := NewThemeHandler(themeStore, dashStore, slog.Default())
	r := mux.NewRouter()
	r.Path("/themes").Methods(http.MethodGet).HandlerFunc(h.List)
	r.Path("/themes").Methods(http.MethodPost).HandlerFunc(h.Create)
	r.Path("/themes/{id}").Methods(http.MethodGet).HandlerFunc(h.Get)
	r.Path("/themes/{id}").Methods(http.MethodPut).HandlerFunc(h.Update)
	r.Path("/themes/{id}").Methods(http.MethodDelete).HandlerFunc(h.Delete)
	r.Path("/themes/{id}/copy").Methods(http.MethodPost).HandlerFunc(h.Copy)
	r.Path("/themes/{id}/assets").Methods(http.MethodGet).HandlerFunc(h.ListAssets)
	r.Path("/themes/{id}/assets/{path:.*}").Methods(http.MethodGet).HandlerFunc(h.GetAsset)
	r.Path("/themes/{id}/assets/{path:.*}").Methods(http.MethodPost).HandlerFunc(h.SaveAsset)
	r.Path("/themes/{id}/assets/{path:.*}").Methods(http.MethodDelete).HandlerFunc(h.DeleteAsset)
	return r, themeStore, dashStore
}

func doJSON(t *testing.T, r *mux.Router, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		rdr = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, rdr)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// validTheme returns a theme that passes validation.
func validTheme(name string) themes.Theme {
	return themes.Theme{Name: name, Colors: themes.DefaultTheme().Colors}
}

func mustCreateTheme(t *testing.T, s *themes.Store, name string) themes.Theme {
	t.Helper()
	created, err := s.Create(validTheme(name))
	if err != nil {
		t.Fatalf("create theme %q: %v", name, err)
	}
	return created
}

func TestThemeCreateAndGet(t *testing.T) {
	r, _, _ := newThemeTestRouter(t)

	rec := doJSON(t, r, http.MethodPost, "/themes", validTheme("Paper"))
	if rec.Code != http.StatusOK {
		t.Fatalf("create must return 200 (a middleware corrupts 201 bodies), got %d body %s", rec.Code, rec.Body)
	}
	var envelope map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &envelope)
	if _, isError := envelope["error"]; isError {
		t.Fatalf("response body must be the theme, not an error envelope: %s", rec.Body)
	}
	var created themes.Theme
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected an assigned id")
	}

	rec = doJSON(t, r, http.MethodGet, "/themes/"+created.ID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get: got %d", rec.Code)
	}
}

func TestThemeCreateIgnoresClientID(t *testing.T) {
	r, _, _ := newThemeTestRouter(t)
	th := validTheme("Sneaky")
	th.ID = themes.DefaultID
	rec := doJSON(t, r, http.MethodPost, "/themes", th)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d body %s", rec.Code, rec.Body)
	}
	var created themes.Theme
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	if created.ID == themes.DefaultID {
		t.Fatal("the store must assign the id, not the client")
	}
}

func TestThemeCreateRejectsInvalidConfig(t *testing.T) {
	r, _, _ := newThemeTestRouter(t)
	rec := doJSON(t, r, http.MethodPost, "/themes", map[string]any{
		"name":   "Bad",
		"colors": map[string]any{"light": map[string]string{"bg": "#fff"}},
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body %s", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), "invalid theme configuration") {
		t.Errorf("expected the validation message, got %s", rec.Body)
	}
}

func TestThemeCreateRejectsMalformedBody(t *testing.T) {
	r, _, _ := newThemeTestRouter(t)
	req := httptest.NewRequest(http.MethodPost, "/themes", strings.NewReader("{"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestThemeGetUnknownIs404(t *testing.T) {
	r, _, _ := newThemeTestRouter(t)
	if rec := doJSON(t, r, http.MethodGet, "/themes/zzzzzz", nil); rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestThemeGetInvalidIDIs400(t *testing.T) {
	r, _, _ := newThemeTestRouter(t)
	if rec := doJSON(t, r, http.MethodGet, "/themes/BAD-ID", nil); rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestThemeGetDefault(t *testing.T) {
	r, _, _ := newThemeTestRouter(t)
	rec := doJSON(t, r, http.MethodGet, "/themes/default", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got themes.Theme
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if got != themes.DefaultTheme() {
		t.Fatalf("got %+v", got)
	}
}

func TestThemeGetFlagsBuiltins(t *testing.T) {
	r, themeStore, _ := newThemeTestRouter(t)
	stored := mustCreateTheme(t, themeStore, "Mine")
	cases := map[string]bool{themes.DefaultID: true, themes.EmberID: true, stored.ID: false}
	for id, want := range cases {
		rec := doJSON(t, r, http.MethodGet, "/themes/"+id, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", id, rec.Code)
		}
		var got struct {
			ID      string `json:"id"`
			Builtin bool   `json:"builtin"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &got)
		if got.ID != id || got.Builtin != want {
			t.Errorf("%s: got id %q builtin %v, want builtin %v", id, got.ID, got.Builtin, want)
		}
	}
}

func TestListUsedBy(t *testing.T) {
	r, themeStore, dashStore := newThemeTestRouter(t)
	th := mustCreateTheme(t, themeStore, "Shared")
	for _, name := range []string{"One", "Two"} {
		if _, err := dashStore.Create(dashboard.Dashboard{Name: name, ThemeID: th.ID}); err != nil {
			t.Fatalf("create dashboard: %v", err)
		}
	}
	// No theme and an explicit "default" both render the builtin; a dangling
	// ID counts toward nothing.
	for _, d := range []dashboard.Dashboard{
		{Name: "Three"},
		{Name: "Four", ThemeID: themes.DefaultID},
		{Name: "Five", ThemeID: "zzzzzz"},
	} {
		if _, err := dashStore.Create(d); err != nil {
			t.Fatalf("create dashboard: %v", err)
		}
	}

	rec := doJSON(t, r, http.MethodGet, "/themes", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: got %d", rec.Code)
	}
	var resp struct {
		Items []themes.Meta `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Items) != 3 {
		t.Fatalf("expected the theme plus the 2 builtins, got %d", len(resp.Items))
	}
	var found bool
	for _, m := range resp.Items {
		if m.ID == th.ID {
			found = true
			if m.UsedBy != 2 {
				t.Fatalf("expected usedBy 2, got %d", m.UsedBy)
			}
		}
		if m.ID == themes.DefaultID && m.UsedBy != 2 {
			t.Errorf("expected the builtin usedBy 2 (empty + explicit), got %d", m.UsedBy)
		}
	}
	if !found {
		t.Fatalf("theme %s missing from %+v", th.ID, resp.Items)
	}
}

func TestThemeUpdateAndDelete(t *testing.T) {
	r, themeStore, _ := newThemeTestRouter(t)
	th := mustCreateTheme(t, themeStore, "Before")

	th.Name = "After"
	if rec := doJSON(t, r, http.MethodPut, "/themes/"+th.ID, th); rec.Code != http.StatusOK {
		t.Fatalf("update: got %d body %s", rec.Code, rec.Body)
	}
	got, err := themeStore.Get(th.ID)
	if err != nil || got.Name != "After" {
		t.Fatalf("expected the rename to persist, got %+v %v", got, err)
	}

	if rec := doJSON(t, r, http.MethodDelete, "/themes/"+th.ID, nil); rec.Code != http.StatusNoContent {
		t.Fatalf("delete: got %d", rec.Code)
	}
}

func TestUpdateBuiltinForbidden(t *testing.T) {
	r, _, _ := newThemeTestRouter(t)
	th := themes.DefaultTheme()
	th.Name = "Hijacked"
	if rec := doJSON(t, r, http.MethodPut, "/themes/default", th); rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body %s", rec.Code, rec.Body)
	}
	if rec := doJSON(t, r, http.MethodDelete, "/themes/default", nil); rec.Code != http.StatusForbidden {
		t.Fatalf("delete: expected 403, got %d", rec.Code)
	}
	req := httptest.NewRequest(http.MethodPost, "/themes/default/assets/a.png", bytes.NewReader([]byte("x")))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("asset upload: expected 403, got %d", rec.Code)
	}
}

func TestThemeDeleteIsAllowedWhileReferenced(t *testing.T) {
	r, themeStore, dashStore := newThemeTestRouter(t)
	th := mustCreateTheme(t, themeStore, "Doomed")
	if _, err := dashStore.Create(dashboard.Dashboard{Name: "User", ThemeID: th.ID}); err != nil {
		t.Fatalf("create dashboard: %v", err)
	}
	if rec := doJSON(t, r, http.MethodDelete, "/themes/"+th.ID, nil); rec.Code != http.StatusNoContent {
		t.Fatalf("a referenced theme must still be deletable, got %d", rec.Code)
	}
}

func TestCopyEndpoint(t *testing.T) {
	r, themeStore, _ := newThemeTestRouter(t)
	th := mustCreateTheme(t, themeStore, "Ocean")

	rec := doJSON(t, r, http.MethodPost, "/themes/"+th.ID+"/copy", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("copy must return 200, got %d body %s", rec.Code, rec.Body)
	}
	var cp themes.Theme
	if err := json.Unmarshal(rec.Body.Bytes(), &cp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !strings.HasPrefix(cp.Name, "Copy of ") || cp.Name != "Copy of Ocean" {
		t.Errorf("name = %q", cp.Name)
	}
	if cp.ID == "" || cp.ID == th.ID {
		t.Errorf("expected a new id, got %q", cp.ID)
	}

	rec = doJSON(t, r, http.MethodPost, "/themes/default/copy", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("copying the builtin: got %d", rec.Code)
	}
	if rec := doJSON(t, r, http.MethodPost, "/themes/zzzzzz/copy", nil); rec.Code != http.StatusNotFound {
		t.Fatalf("copying an unknown theme: expected 404, got %d", rec.Code)
	}
}

func TestThemeAssetUploadAndFetch(t *testing.T) {
	r, themeStore, _ := newThemeTestRouter(t)
	th := mustCreateTheme(t, themeStore, "Assets")

	req := httptest.NewRequest(http.MethodPost, "/themes/"+th.ID+"/assets/pic.png", bytes.NewReader([]byte("bytes")))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload: got %d body %s", rec.Code, rec.Body)
	}

	rec = doJSON(t, r, http.MethodGet, "/themes/"+th.ID+"/assets/pic.png", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("fetch: got %d", rec.Code)
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("expected nosniff on served asset bytes")
	}
	if got := rec.Body.String(); got != "bytes" {
		t.Errorf("got %q", got)
	}

	rec = doJSON(t, r, http.MethodGet, "/themes/"+th.ID+"/assets", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list assets: got %d", rec.Code)
	}

	if rec := doJSON(t, r, http.MethodDelete, "/themes/"+th.ID+"/assets/pic.png", nil); rec.Code != http.StatusNoContent {
		t.Fatalf("delete asset: got %d", rec.Code)
	}
}

func TestThemeAssetRejectsTraversal(t *testing.T) {
	r, themeStore, _ := newThemeTestRouter(t)
	th := mustCreateTheme(t, themeStore, "Assets")

	req := httptest.NewRequest(http.MethodPost, "/themes/"+th.ID+"/assets/..%2F..%2Fescape.png", bytes.NewReader([]byte("x")))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code == http.StatusCreated {
		t.Fatal("traversal upload must be rejected")
	}
}
