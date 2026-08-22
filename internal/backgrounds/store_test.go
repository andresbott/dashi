package backgrounds

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/data/images"
)

func newTestStore(t *testing.T) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	pool, err := images.NewStore(filepath.Join(dir, "pool"))
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	return NewStore(filepath.Join(dir, "backgrounds"), pool), dir
}

func TestCreateAssignsIDAndSnakeCaseFolder(t *testing.T) {
	s, dir := newTestStore(t)
	got, err := s.Create(Background{Name: "My Sunset!"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if len(got.ID) != 6 {
		t.Fatalf("expected a 6-char id, got %q", got.ID)
	}
	for _, c := range got.ID {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') {
			t.Fatalf("id %q is not lowercase alphanumeric", got.ID)
		}
	}
	path := filepath.Join(dir, "backgrounds", "my_sunset", "background.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func TestCreateRejectsInvalidConfig(t *testing.T) {
	s, _ := newTestStore(t)
	_, err := s.Create(Background{
		Name:     "Bad",
		Color:    &Color{Light: "#fff"},
		Gradient: &Gradient{Direction: "to right", Light: []string{"#000", "#fff"}},
	})
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestGetRoundTrips(t *testing.T) {
	s, _ := newTestStore(t)
	created, err := s.Create(Background{
		Name:     "Dusk",
		Gradient: &Gradient{Direction: "135deg", Light: []string{"#000000", "#ffffff"}},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := s.Get(created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Dusk" || got.Gradient == nil || got.Gradient.Direction != "135deg" {
		t.Fatalf("round trip lost data: %+v", got)
	}
}

func TestGetUnknownIDReturnsNotFound(t *testing.T) {
	s, _ := newTestStore(t)
	if _, err := s.Get("zzzzzz"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetInvalidIDReturnsInvalidID(t *testing.T) {
	s, _ := newTestStore(t)
	if _, err := s.Get("../etc"); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}
}

func TestUpdateRenamesFolderButKeepsID(t *testing.T) {
	s, dir := newTestStore(t)
	created, _ := s.Create(Background{Name: "Before"})
	created.Name = "After"
	if _, err := s.Update(created); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "backgrounds", "after")); err != nil {
		t.Fatalf("expected the folder to be renamed: %v", err)
	}
	got, err := s.Get(created.ID)
	if err != nil {
		t.Fatalf("the id must still resolve after a rename: %v", err)
	}
	if got.Name != "After" {
		t.Fatalf("got name %q", got.Name)
	}
}

func TestUpdateUnknownIDReturnsNotFound(t *testing.T) {
	s, _ := newTestStore(t)
	if _, err := s.Update(Background{ID: "zzzzzz", Name: "Ghost"}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteRemovesFolderAndIndexEntry(t *testing.T) {
	s, dir := newTestStore(t)
	created, _ := s.Create(Background{Name: "Gone"})
	if err := s.Delete(created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "backgrounds", "gone")); !os.IsNotExist(err) {
		t.Fatalf("expected the folder to be gone, got %v", err)
	}
	if _, err := s.Get(created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestListSortsByNameAndCarriesUsageAndPreview(t *testing.T) {
	s, _ := newTestStore(t)
	b, _ := s.Create(Background{Name: "Zebra", Color: &Color{Light: "#ffffff"}})
	_, _ = s.Create(Background{Name: "apple"})

	metas, err := s.List(map[string]int{b.ID: 3})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(metas) != 2 {
		t.Fatalf("expected 2 backgrounds, got %d", len(metas))
	}
	if metas[0].Name != "apple" {
		t.Fatalf("expected case-insensitive name order, got %q first", metas[0].Name)
	}
	var zebra Meta
	for _, m := range metas {
		if m.ID == b.ID {
			zebra = m
		}
	}
	if zebra.UsedBy != 3 {
		t.Errorf("expected UsedBy 3, got %d", zebra.UsedBy)
	}
	if !strings.Contains(zebra.PreviewCSS, "#ffffff") {
		t.Errorf("expected preview CSS, got %q", zebra.PreviewCSS)
	}
}

func TestIndexIsRebuiltOnStartup(t *testing.T) {
	dir := t.TempDir()
	pool, _ := images.NewStore(filepath.Join(dir, "pool"))
	bgDir := filepath.Join(dir, "backgrounds")

	first := NewStore(bgDir, pool)
	created, _ := first.Create(Background{Name: "Persisted"})

	second := NewStore(bgDir, pool)
	if _, err := second.Get(created.ID); err != nil {
		t.Fatalf("a fresh store must find the background by id: %v", err)
	}
}

func TestAssetLifecycle(t *testing.T) {
	s, _ := newTestStore(t)
	created, _ := s.Create(Background{Name: "Assets"})

	if err := s.SaveAsset(created.ID, "pic.png", []byte("png-bytes")); err != nil {
		t.Fatalf("save asset: %v", err)
	}
	list, err := s.ListAssets(created.ID)
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(list) != 1 || list[0] != "pic.png" {
		t.Fatalf("got %v", list)
	}
	data, mime, err := s.GetAsset(created.ID, "pic.png")
	if err != nil {
		t.Fatalf("get asset: %v", err)
	}
	if string(data) != "png-bytes" {
		t.Errorf("got %q", data)
	}
	if mime != "image/png" {
		t.Errorf("got mime %q", mime)
	}
	if err := s.DeleteAsset(created.ID, "pic.png"); err != nil {
		t.Fatalf("delete asset: %v", err)
	}
	if list, _ := s.ListAssets(created.ID); len(list) != 0 {
		t.Fatalf("expected no assets, got %v", list)
	}
}

func TestAssetRejectsTraversalAndBadExtension(t *testing.T) {
	s, _ := newTestStore(t)
	created, _ := s.Create(Background{Name: "Assets"})

	cases := []string{"../escape.png", "/abs.png", "script.js", "background.json", ""}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			if err := s.SaveAsset(created.ID, name, []byte("x")); err == nil {
				t.Fatalf("expected %q to be rejected", name)
			}
		})
	}
}

func TestListAssetsExcludesBackgroundJSON(t *testing.T) {
	s, _ := newTestStore(t)
	created, _ := s.Create(Background{Name: "Assets"})
	list, err := s.ListAssets(created.ID)
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	for _, a := range list {
		if a == "background.json" {
			t.Fatal("background.json must not appear as an asset")
		}
	}
}

func TestLoadImageFromAssetAndPool(t *testing.T) {
	dir := t.TempDir()
	pool, _ := images.NewStore(filepath.Join(dir, "pool"))
	if err := pool.Save("shared.png", []byte("pool-bytes")); err != nil {
		t.Fatalf("seed pool: %v", err)
	}
	s := NewStore(filepath.Join(dir, "backgrounds"), pool)
	created, _ := s.Create(Background{Name: "Loader"})
	_ = s.SaveAsset(created.ID, "own.png", []byte("own-bytes"))

	own, err := s.LoadImage(created.ID, "asset:own.png")
	if err != nil || string(own) != "own-bytes" {
		t.Fatalf("asset ref: %q, %v", own, err)
	}
	shared, err := s.LoadImage(created.ID, "shared:shared.png")
	if err != nil || string(shared) != "pool-bytes" {
		t.Fatalf("shared ref: %q, %v", shared, err)
	}
	if _, err := s.LoadImage(created.ID, "theme:default/x.png"); err == nil {
		t.Fatal("theme refs must be rejected")
	}
}

func TestRollbackCreateRemovesFolderAndIndexEntry(t *testing.T) {
	s, _ := newTestStore(t)

	// Create a background successfully so folder and index entry exist
	created, err := s.Create(Background{Name: "ToBeRolledBack"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Get the actual folder path from bgDir
	folderPath, ok := s.bgDir(created.ID)
	if !ok {
		t.Fatalf("background %s not in index", created.ID)
	}

	// Establish precondition: folder exists on disk
	if _, err := os.Stat(folderPath); err != nil {
		t.Fatalf("expected folder to exist before rollback: %v", err)
	}

	// Establish precondition: Get succeeds
	if _, err := s.Get(created.ID); err != nil {
		t.Fatalf("expected Get to succeed before rollback: %v", err)
	}

	// Call rollbackCreate
	s.rollbackCreate(created.ID, folderPath)

	// Assert folder is gone
	if _, err := os.Stat(folderPath); !os.IsNotExist(err) {
		t.Fatalf("expected folder to be removed, got %v", err)
	}

	// Assert Get now returns ErrNotFound
	if _, err := s.Get(created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after rollback, got %v", err)
	}
}
