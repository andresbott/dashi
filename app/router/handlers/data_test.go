package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/data/backgrounds"
	"github.com/andresbott/dashi/internal/data/images"
	"github.com/andresbott/dashi/internal/data/notes"
	"github.com/gorilla/mux"
)

func newDataHandler(t *testing.T) (*DataHandler, string) {
	t.Helper()
	dir := t.TempDir()
	ns, err := notes.NewStore(dir + "/notes")
	if err != nil {
		t.Fatal(err)
	}
	is, err := images.NewStore(dir + "/images")
	if err != nil {
		t.Fatal(err)
	}
	bs, err := backgrounds.NewStore(dir + "/backgrounds")
	if err != nil {
		t.Fatal(err)
	}
	return NewDataHandler(ns, is, bs, nil), dir
}

func doReq(t *testing.T, h *DataHandler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()
	r := mux.NewRouter()
	h.RegisterRead(r)
	h.RegisterWrite(r)
	req := httptest.NewRequest(method, path, body)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestDataHandler_ListNotes_Empty(t *testing.T) {
	h, _ := newDataHandler(t)
	w := doReq(t, h, http.MethodGet, "/data/notes", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	var items []map[string]any
	if err := json.NewDecoder(w.Body).Decode(&items); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestDataHandler_NotesRoundTrip(t *testing.T) {
	h, _ := newDataHandler(t)

	// Save
	w := doReq(t, h, http.MethodPost, "/data/notes/a.md", strings.NewReader("# hi"))
	if w.Code != http.StatusCreated {
		t.Fatalf("save status %d: %s", w.Code, w.Body.String())
	}

	// Get rendered
	w = doReq(t, h, http.MethodGet, "/data/notes/a.md", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get status %d: %s", w.Code, w.Body.String())
	}
	var rendered struct {
		HTML string `json:"html"`
	}
	if err := json.NewDecoder(w.Body).Decode(&rendered); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !strings.Contains(rendered.HTML, "<h1>hi</h1>") {
		t.Fatalf("unexpected html: %q", rendered.HTML)
	}

	// Get raw
	w = doReq(t, h, http.MethodGet, "/data/notes/a.md/raw", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("raw status %d", w.Code)
	}
	if w.Body.String() != "# hi" {
		t.Fatalf("raw body: got %q, want %q", w.Body.String(), "# hi")
	}

	// Delete
	w = doReq(t, h, http.MethodDelete, "/data/notes/a.md", nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete status %d", w.Code)
	}
}

func TestDataHandler_InvalidName_400(t *testing.T) {
	h, _ := newDataHandler(t)
	w := doReq(t, h, http.MethodPost, "/data/notes/bad.txt", strings.NewReader("x"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDataHandler_NotFound_404(t *testing.T) {
	h, _ := newDataHandler(t)
	w := doReq(t, h, http.MethodGet, "/data/notes/missing.md", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDataHandler_ImagesRoundTrip(t *testing.T) {
	h, _ := newDataHandler(t)

	body := strings.NewReader("\x89PNG\r\n\x1a\n")
	w := doReq(t, h, http.MethodPost, "/data/images/hero.png", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("save status %d: %s", w.Code, w.Body.String())
	}

	w = doReq(t, h, http.MethodGet, "/data/images/hero.png", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("get status %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "image/png") {
		t.Fatalf("content-type: got %q, want image/png...", ct)
	}
}
