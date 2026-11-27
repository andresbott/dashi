package bookmarks

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andresbott/dashi/internal/data"
)

const sample = `bookmarks:
  - title: GitHub
    url: https://github.com
    icon: ti-brand-github
    subtitle: Code hosting
  - title: Notes
    url: /notes
`

func newTestStore(t *testing.T) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s, dir
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    int
		wantErr bool
	}{
		{name: "sample", raw: sample, want: 2},
		{name: "empty", raw: "", want: 0},
		{name: "only a comment", raw: "# nothing yet\n", want: 0},
		{name: "empty list", raw: "bookmarks: []\n", want: 0},
		{name: "not yaml", raw: "bookmarks: [", wantErr: true},
		{name: "unknown key", raw: "bookmarks:\n  - title: x\n    subtile: y\n", wantErr: true},
		{name: "top-level list", raw: "- title: x\n", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			items, err := Parse([]byte(tc.raw))
			if tc.wantErr {
				if !errors.Is(err, ErrInvalidFile) {
					t.Fatalf("err = %v, want ErrInvalidFile", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if items == nil || len(items) != tc.want {
				t.Fatalf("got %#v, want %d items", items, tc.want)
			}
		})
	}

	items, _ := Parse([]byte(sample))
	want := Item{URL: "https://github.com", Icon: "ti-brand-github", Title: "GitHub", Subtitle: "Code hosting"}
	if items[0] != want {
		t.Errorf("first item = %#v, want %#v", items[0], want)
	}
}

func TestStore_SaveGetListDelete(t *testing.T) {
	s, _ := newTestStore(t)
	if err := s.Save("links.yaml", sample); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := s.Get("links.yaml")
	if err != nil || got != sample {
		t.Fatalf("get = %q, %v", got, err)
	}
	list, err := s.List()
	if err != nil || len(list) != 1 || list[0].Name != "links.yaml" {
		t.Fatalf("list = %#v, %v", list, err)
	}
	if err := s.Delete("links.yaml"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.Get("links.yaml"); !errors.Is(err, data.ErrNotFound) {
		t.Fatalf("after delete: %v", err)
	}
	if err := s.Delete("links.yaml"); !errors.Is(err, data.ErrNotFound) {
		t.Fatalf("second delete: %v", err)
	}
}

func TestStore_SaveItemsRoundTrip(t *testing.T) {
	s, _ := newTestStore(t)
	items := []Item{
		{URL: "https://a.example", Icon: "ti-world", Title: "A: colon", Subtitle: "first"},
		{URL: "/b", Title: "B"},
	}
	if err := s.SaveItems("links.yaml", items); err != nil {
		t.Fatal(err)
	}
	got, err := s.Read("links.yaml")
	if err != nil || len(got) != 2 || got[0] != items[0] || got[1] != items[1] {
		t.Fatalf("read back %#v, %v", got, err)
	}
	if err := s.SaveItems("empty.yaml", nil); err != nil {
		t.Fatal(err)
	}
	if got, err := s.Read("empty.yaml"); err != nil || len(got) != 0 {
		t.Fatalf("empty: %#v, %v", got, err)
	}
	if err := s.SaveItems("bad.txt", items); !errors.Is(err, data.ErrExtNotAllowed) {
		t.Fatalf("bad name: %v", err)
	}
}

func TestStore_SaveRejects(t *testing.T) {
	s, _ := newTestStore(t)
	if err := s.Save("bad.yaml", "bookmarks: ["); !errors.Is(err, ErrInvalidFile) {
		t.Errorf("invalid yaml: err = %v", err)
	}
	if _, err := s.Get("bad.yaml"); !errors.Is(err, data.ErrNotFound) {
		t.Errorf("invalid content must not be written: %v", err)
	}
	if err := s.Save("links.md", sample); !errors.Is(err, data.ErrExtNotAllowed) {
		t.Errorf("wrong extension: err = %v", err)
	}
	if err := s.Save("links.yml", sample); err != nil {
		t.Errorf(".yml must be accepted: %v", err)
	}
}

// Files edited on disk are picked up once the cache entry is older than
// CacheTTL; saves through the store apply at once.
func TestStore_BookmarksCache(t *testing.T) {
	s, dir := newTestStore(t)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }

	if _, err := s.Bookmarks("links.yaml"); !errors.Is(err, data.ErrNotFound) {
		t.Fatalf("missing file: err = %v", err)
	}
	if err := s.Save("links.yaml", sample); err != nil {
		t.Fatal(err)
	}
	items, err := s.Bookmarks("links.yaml")
	if err != nil || len(items) != 2 {
		t.Fatalf("after save: %d items, %v", len(items), err)
	}

	// Edit on disk, behind the store's back.
	if err := os.WriteFile(filepath.Join(dir, "links.yaml"), []byte("bookmarks:\n  - title: Only\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	now = now.Add(CacheTTL - time.Second)
	if items, _ := s.Bookmarks("links.yaml"); len(items) != 2 {
		t.Fatalf("within the TTL the cached copy is served, got %d items", len(items))
	}
	now = now.Add(2 * time.Second)
	if items, _ := s.Bookmarks("links.yaml"); len(items) != 1 || items[0].Title != "Only" {
		t.Fatalf("after the TTL the file is re-read, got %#v", items)
	}

	// A broken file on disk reports ErrInvalidFile.
	if err := os.WriteFile(filepath.Join(dir, "links.yaml"), []byte("bookmarks: ["), 0o600); err != nil {
		t.Fatal(err)
	}
	now = now.Add(CacheTTL)
	if _, err := s.Bookmarks("links.yaml"); !errors.Is(err, ErrInvalidFile) {
		t.Fatalf("broken file: err = %v", err)
	}

	if err := s.Delete("links.yaml"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Bookmarks("links.yaml"); !errors.Is(err, data.ErrNotFound) {
		t.Fatalf("delete must drop the cached copy: err = %v", err)
	}
}

// Missing files and bad names are never cached: a file created on disk
// shows at once, and made-up names cannot grow the cache.
func TestStore_BookmarksCacheOnlyExistingFiles(t *testing.T) {
	s, dir := newTestStore(t)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }

	for _, name := range []string{"gone.yaml", "../x.yaml", "x.txt"} {
		if _, err := s.Bookmarks(name); err == nil {
			t.Fatalf("%s: want an error", name)
		}
	}
	if len(s.cache) != 0 {
		t.Fatalf("failed lookups must not be cached: %v", s.cache)
	}

	if err := os.WriteFile(filepath.Join(dir, "gone.yaml"), []byte(sample), 0o600); err != nil {
		t.Fatal(err)
	}
	if items, err := s.Bookmarks("gone.yaml"); err != nil || len(items) != 2 {
		t.Fatalf("a file created on disk shows at once: %d items, %v", len(items), err)
	}

	// Expired entries are dropped when a new file is loaded.
	if err := os.WriteFile(filepath.Join(dir, "other.yaml"), []byte(sample), 0o600); err != nil {
		t.Fatal(err)
	}
	now = now.Add(CacheTTL)
	if _, err := s.Bookmarks("other.yaml"); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.cache["gone.yaml"]; ok || len(s.cache) != 1 {
		t.Fatalf("expired entries must be pruned: %v", s.cache)
	}
}
