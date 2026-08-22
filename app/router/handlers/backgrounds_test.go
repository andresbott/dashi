package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/andresbott/dashi/internal/backgrounds"
	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/data/images"
	"github.com/gorilla/mux"
)

func newBgTestRouter(t *testing.T) (*mux.Router, *backgrounds.Store, *dashboard.Store) {
	t.Helper()
	dir := t.TempDir()
	pool, err := images.NewStore(filepath.Join(dir, "pool"))
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	bgStore := backgrounds.NewStore(filepath.Join(dir, "backgrounds"), pool)
	dashStore := dashboard.NewStore(filepath.Join(dir, "dashboards"))

	h := NewBackgroundHandler(bgStore, dashStore, slog.Default())
	r := mux.NewRouter()
	r.Path("/backgrounds").Methods(http.MethodGet).HandlerFunc(h.List)
	r.Path("/backgrounds").Methods(http.MethodPost).HandlerFunc(h.Create)
	r.Path("/backgrounds/{id}").Methods(http.MethodGet).HandlerFunc(h.Get)
	r.Path("/backgrounds/{id}").Methods(http.MethodPut).HandlerFunc(h.Update)
	r.Path("/backgrounds/{id}").Methods(http.MethodDelete).HandlerFunc(h.Delete)
	r.Path("/backgrounds/{id}/assets").Methods(http.MethodGet).HandlerFunc(h.ListAssets)
	r.Path("/backgrounds/{id}/assets/{path:.*}").Methods(http.MethodGet).HandlerFunc(h.GetAsset)
	r.Path("/backgrounds/{id}/assets/{path:.*}").Methods(http.MethodPost).HandlerFunc(h.SaveAsset)
	r.Path("/backgrounds/{id}/assets/{path:.*}").Methods(http.MethodDelete).HandlerFunc(h.DeleteAsset)
	return r, bgStore, dashStore
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

func TestBackgroundCreateAndGet(t *testing.T) {
	r, _, _ := newBgTestRouter(t)

	rec := doJSON(t, r, http.MethodPost, "/backgrounds", backgrounds.Background{
		Name:  "Paper",
		Color: &backgrounds.Color{Light: "#ffffff"},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("create must return 200 (a middleware corrupts 201 bodies), got %d", rec.Code)
	}
	var envelope map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &envelope)
	if _, isError := envelope["error"]; isError {
		t.Fatalf("response body must be the background, not an error envelope: %s", rec.Body)
	}
	var created backgrounds.Background
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected an assigned id")
	}

	rec = doJSON(t, r, http.MethodGet, "/backgrounds/"+created.ID, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("get: got %d", rec.Code)
	}
}

func TestBackgroundCreateRejectsInvalidConfig(t *testing.T) {
	r, _, _ := newBgTestRouter(t)
	rec := doJSON(t, r, http.MethodPost, "/backgrounds", map[string]any{
		"name":     "Bad",
		"color":    map[string]string{"light": "#fff"},
		"gradient": map[string]any{"direction": "to right", "light": []string{"#000", "#fff"}},
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body %s", rec.Code, rec.Body)
	}
}

func TestBackgroundGetUnknownIs404(t *testing.T) {
	r, _, _ := newBgTestRouter(t)
	if rec := doJSON(t, r, http.MethodGet, "/backgrounds/zzzzzz", nil); rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestBackgroundListCountsUsage(t *testing.T) {
	r, bgStore, dashStore := newBgTestRouter(t)
	bg, err := bgStore.Create(backgrounds.Background{Name: "Shared"})
	if err != nil {
		t.Fatalf("create bg: %v", err)
	}
	for _, name := range []string{"One", "Two"} {
		if _, err := dashStore.Create(dashboard.Dashboard{Name: name, BackgroundID: bg.ID}); err != nil {
			t.Fatalf("create dashboard: %v", err)
		}
	}
	if _, err := dashStore.Create(dashboard.Dashboard{Name: "Three"}); err != nil {
		t.Fatalf("create dashboard: %v", err)
	}

	rec := doJSON(t, r, http.MethodGet, "/backgrounds", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: got %d", rec.Code)
	}
	var resp struct {
		Items []backgrounds.Meta `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 background, got %d", len(resp.Items))
	}
	if resp.Items[0].UsedBy != 2 {
		t.Fatalf("expected UsedBy 2, got %d", resp.Items[0].UsedBy)
	}
}

func TestBackgroundUpdateAndDelete(t *testing.T) {
	r, bgStore, _ := newBgTestRouter(t)
	bg, _ := bgStore.Create(backgrounds.Background{Name: "Before"})

	bg.Name = "After"
	if rec := doJSON(t, r, http.MethodPut, "/backgrounds/"+bg.ID, bg); rec.Code != http.StatusOK {
		t.Fatalf("update: got %d body %s", rec.Code, rec.Body)
	}
	got, err := bgStore.Get(bg.ID)
	if err != nil || got.Name != "After" {
		t.Fatalf("expected the rename to persist, got %+v %v", got, err)
	}

	if rec := doJSON(t, r, http.MethodDelete, "/backgrounds/"+bg.ID, nil); rec.Code != http.StatusNoContent {
		t.Fatalf("delete: got %d", rec.Code)
	}
}

func TestBackgroundDeleteIsAllowedWhileReferenced(t *testing.T) {
	r, bgStore, dashStore := newBgTestRouter(t)
	bg, _ := bgStore.Create(backgrounds.Background{Name: "Doomed"})
	if _, err := dashStore.Create(dashboard.Dashboard{Name: "User", BackgroundID: bg.ID}); err != nil {
		t.Fatalf("create dashboard: %v", err)
	}
	if rec := doJSON(t, r, http.MethodDelete, "/backgrounds/"+bg.ID, nil); rec.Code != http.StatusNoContent {
		t.Fatalf("a referenced background must still be deletable, got %d", rec.Code)
	}
}

func TestBackgroundAssetUploadAndFetch(t *testing.T) {
	r, bgStore, _ := newBgTestRouter(t)
	bg, _ := bgStore.Create(backgrounds.Background{Name: "Assets"})

	req := httptest.NewRequest(http.MethodPost, "/backgrounds/"+bg.ID+"/assets/pic.png", bytes.NewReader([]byte("bytes")))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload: got %d body %s", rec.Code, rec.Body)
	}

	rec = doJSON(t, r, http.MethodGet, "/backgrounds/"+bg.ID+"/assets/pic.png", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("fetch: got %d", rec.Code)
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("expected nosniff on served asset bytes")
	}
	if got := rec.Body.String(); got != "bytes" {
		t.Errorf("got %q", got)
	}

	rec = doJSON(t, r, http.MethodGet, "/backgrounds/"+bg.ID+"/assets", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list assets: got %d", rec.Code)
	}
}

func TestBackgroundAssetRejectsTraversal(t *testing.T) {
	r, bgStore, _ := newBgTestRouter(t)
	bg, _ := bgStore.Create(backgrounds.Background{Name: "Assets"})

	req := httptest.NewRequest(http.MethodPost, "/backgrounds/"+bg.ID+"/assets/..%2F..%2Fescape.png", bytes.NewReader([]byte("x")))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code == http.StatusCreated {
		t.Fatal("traversal upload must be rejected")
	}
}
