package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/gorilla/mux"
)

// setupDashboard creates a dashboard on disk in a temp dir and returns
// the store and dashboard ID for use in tests.
func setupDashboard(t *testing.T) (*dashboard.Store, string) {
	t.Helper()
	dir := t.TempDir()
	store := dashboard.NewStore(dir)
	d, err := store.Create(dashboard.Dashboard{Name: "Test"})
	if err != nil {
		t.Fatalf("create dashboard: %v", err)
	}
	return store, d.ID
}

func writeAsset(t *testing.T, store *dashboard.Store, id, assetPath string, data []byte) {
	t.Helper()
	if err := store.SaveAsset(id, assetPath, data); err != nil {
		t.Fatalf("save asset %s: %v", assetPath, err)
	}
}

func callListMarkdown(t *testing.T, h *MarkdownHandler, id string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v0/dashboards/"+id+"/markdown", nil)
	req = mux.SetURLVars(req, map[string]string{"id": id})
	rec := httptest.NewRecorder()
	h.ListMarkdown(rec, req)
	return rec
}

func TestMarkdownHandler_ListMarkdown_EmptyFolder(t *testing.T) {
	store, id := setupDashboard(t)
	h := NewMarkdownHandler(store, slog.Default())

	rec := callListMarkdown(t, h, id)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body struct {
		Files []string `json:"files"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Files == nil {
		t.Error("files is nil, want [] (empty slice)")
	}
	if len(body.Files) != 0 {
		t.Errorf("files = %v, want empty", body.Files)
	}
}

func TestMarkdownHandler_ListMarkdown_FiltersAndSorts(t *testing.T) {
	store, id := setupDashboard(t)
	writeAsset(t, store, id, "md/welcome.md", []byte("# hi"))
	writeAsset(t, store, id, "md/notes.md", []byte("# notes"))
	// a non-.md file under md/ — should be excluded
	writeAsset(t, store, id, "md/cover.png", []byte{0x89, 0x50, 0x4e, 0x47})
	// a .md file NOT under md/ — should be excluded
	writeAsset(t, store, id, "readme.md", []byte("root"))
	// an image in the root — should be excluded
	writeAsset(t, store, id, "bg.jpg", []byte{0xff, 0xd8, 0xff})

	h := NewMarkdownHandler(store, slog.Default())
	rec := callListMarkdown(t, h, id)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body struct {
		Files []string `json:"files"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := []string{"notes.md", "welcome.md"}
	if len(body.Files) != len(want) {
		t.Fatalf("files = %v, want %v", body.Files, want)
	}
	for i := range want {
		if body.Files[i] != want[i] {
			t.Errorf("files[%d] = %q, want %q", i, body.Files[i], want[i])
		}
	}
}

func TestMarkdownHandler_ListMarkdown_MissingDashboard(t *testing.T) {
	// Make a store whose directory exists but the dashboard ID does not
	store := dashboard.NewStore(t.TempDir())
	h := NewMarkdownHandler(store, slog.Default())

	// Use a syntactically valid ID that does not exist
	rec := callListMarkdown(t, h, "abc123")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestMarkdownHandler_ListMarkdown_InvalidID(t *testing.T) {
	store := dashboard.NewStore(t.TempDir())
	h := NewMarkdownHandler(store, slog.Default())

	// Capital letters are not valid IDs per the store's isValidID
	rec := callListMarkdown(t, h, "BAD!ID")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
