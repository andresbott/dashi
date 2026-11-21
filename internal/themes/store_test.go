package themes

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
	return NewStore(filepath.Join(dir, "themes"), pool), dir
}

// mustCreate creates a valid theme named name and fails the test on error.
func mustCreate(t *testing.T, s *Store, name string) Theme {
	t.Helper()
	created, err := s.Create(testTheme(name))
	if err != nil {
		t.Fatalf("create %q: %v", name, err)
	}
	return created
}

func TestCreateAssignsIDAndSnakeCaseFolder(t *testing.T) {
	s, dir := newTestStore(t)
	got := mustCreate(t, s, "My Sunset!")
	if len(got.ID) != 6 {
		t.Fatalf("expected a 6-char id, got %q", got.ID)
	}
	for _, c := range got.ID {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') {
			t.Fatalf("id %q is not lowercase alphanumeric", got.ID)
		}
	}
	path := filepath.Join(dir, "themes", "my_sunset", "theme.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func TestCreateUnnamableFallsBackToThemeFolder(t *testing.T) {
	s, dir := newTestStore(t)
	mustCreate(t, s, "!!!")
	if _, err := os.Stat(filepath.Join(dir, "themes", "theme", "theme.json")); err != nil {
		t.Fatalf("expected the fallback folder name: %v", err)
	}
}

func TestCreateRejectsInvalidConfig(t *testing.T) {
	s, _ := newTestStore(t)
	th := testTheme("Bad")
	th.Background.Gradient = &Gradient{Direction: "sideways", Light: []string{"#000", "#fff"}}
	if _, err := s.Create(th); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestCreateRejectsIncompletePalette(t *testing.T) {
	s, _ := newTestStore(t)
	th := testTheme("Half")
	th.Colors.Dark = Palette{}
	if _, err := s.Create(th); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
}

// builtinIDs lists every builtin, so each builtin test runs for all of them.
var builtinIDs = []string{DefaultID, EmberID}

func TestCreateRejectsBuiltinID(t *testing.T) {
	s, _ := newTestStore(t)
	for _, id := range builtinIDs {
		th := testTheme("Imposter")
		th.ID = id
		if _, err := s.Create(th); !errors.Is(err, ErrBuiltin) {
			t.Fatalf("%s: expected ErrBuiltin, got %v", id, err)
		}
	}
}

func TestGetRoundTrips(t *testing.T) {
	s, _ := newTestStore(t)
	th := testTheme("Dusk")
	th.Background.Gradient = &Gradient{Direction: "135deg", Light: []string{"#000000", "#ffffff"}}
	th.Colors.Dark.Surface = "#23233d80"
	created, err := s.Create(th)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := s.Get(created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Dusk" || got.Background.Gradient == nil || got.Background.Gradient.Direction != "135deg" {
		t.Fatalf("round trip lost data: %+v", got)
	}
	if got.Colors != th.Colors {
		t.Fatalf("round trip lost colours: %+v", got.Colors)
	}
}

func TestGetDefaultIsBuiltin(t *testing.T) {
	s, _ := newTestStore(t)
	got, err := s.Get("default")
	if err != nil {
		t.Fatalf("get default: %v", err)
	}
	if got != DefaultTheme() {
		t.Fatalf("got %+v, want DefaultTheme()", got)
	}
}

func TestGetEmberIsBuiltin(t *testing.T) {
	s, _ := newTestStore(t)
	got, err := s.Get(EmberID)
	if err != nil {
		t.Fatalf("get ember: %v", err)
	}
	if got != EmberTheme() {
		t.Fatalf("got %+v, want EmberTheme()", got)
	}
	if s.Resolve(EmberID) != EmberTheme() {
		t.Fatal("Resolve must return the Ember builtin, not fall back to default")
	}
}

func TestBuiltinThemesAreValid(t *testing.T) {
	for _, id := range builtinIDs {
		th, ok := Builtin(id)
		if !ok {
			t.Fatalf("%s: not registered", id)
		}
		if th.ID != id {
			t.Errorf("%s: constructor returns ID %q", id, th.ID)
		}
		if err := th.Validate(); err != nil {
			t.Errorf("%s: %v", id, err)
		}
		// A 6-character ID could collide with a randomID.
		if len(id) == idLen {
			t.Errorf("%s: builtin IDs must not be %d characters long", id, idLen)
		}
	}
	if len(builtins) != len(builtinIDs) {
		t.Errorf("builtinIDs is out of date: %d registered, %d listed", len(builtins), len(builtinIDs))
	}
}

func TestResolveUnknownFallsBack(t *testing.T) {
	s, _ := newTestStore(t)
	for _, id := range []string{"zzzzzz", "", "../etc"} {
		if got := s.Resolve(id); got != DefaultTheme() {
			t.Errorf("Resolve(%q) = %+v, want DefaultTheme()", id, got)
		}
	}
}

func TestResolveReturnsStoredTheme(t *testing.T) {
	s, _ := newTestStore(t)
	created := mustCreate(t, s, "Stored")
	if got := s.Resolve(created.ID); got.ID != created.ID || got.Name != "Stored" {
		t.Fatalf("got %+v", got)
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

func TestBuiltinIsReadOnly(t *testing.T) {
	s, _ := newTestStore(t)
	for _, id := range builtinIDs {
		th, _ := Builtin(id)
		th.Name = "Hijacked"
		if _, err := s.Update(th); !errors.Is(err, ErrBuiltin) {
			t.Errorf("%s Update: expected ErrBuiltin, got %v", id, err)
		}
		if err := s.Delete(id); !errors.Is(err, ErrBuiltin) {
			t.Errorf("%s Delete: expected ErrBuiltin, got %v", id, err)
		}
		if err := s.SaveAsset(id, "a.png", []byte("x")); !errors.Is(err, ErrBuiltin) {
			t.Errorf("%s SaveAsset: expected ErrBuiltin, got %v", id, err)
		}
		if err := s.DeleteAsset(id, "a.png"); !errors.Is(err, ErrBuiltin) {
			t.Errorf("%s DeleteAsset: expected ErrBuiltin, got %v", id, err)
		}
	}
}

func TestBuiltinHasNoAssets(t *testing.T) {
	s, _ := newTestStore(t)
	for _, id := range builtinIDs {
		list, err := s.ListAssets(id)
		if err != nil {
			t.Fatalf("%s list assets: %v", id, err)
		}
		if list == nil || len(list) != 0 {
			t.Fatalf("%s: expected an empty, non-nil list, got %#v", id, list)
		}
	}
	if _, _, err := s.GetAsset(DefaultID, "a.png"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateRenamesFolderButKeepsID(t *testing.T) {
	s, dir := newTestStore(t)
	created := mustCreate(t, s, "Before")
	created.Name = "After"
	if _, err := s.Update(created); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "themes", "after")); err != nil {
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
	th := testTheme("Ghost")
	th.ID = "zzzzzz"
	if _, err := s.Update(th); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateRejectsIncompletePalette(t *testing.T) {
	s, _ := newTestStore(t)
	created := mustCreate(t, s, "Full")
	created.Colors.Light.Text = ""
	if _, err := s.Update(created); !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestDeleteRemovesFolderAndIndexEntry(t *testing.T) {
	s, dir := newTestStore(t)
	created := mustCreate(t, s, "Gone")
	if err := s.Delete(created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "themes", "gone")); !os.IsNotExist(err) {
		t.Fatalf("expected the folder to be gone, got %v", err)
	}
	if _, err := s.Get(created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestListIncludesBuiltins(t *testing.T) {
	s, _ := newTestStore(t)
	metas, err := s.List(nil)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(metas) != 2 {
		t.Fatalf("expected only the builtins, got %+v", metas)
	}
	// Sorted by name: Default, Ember.
	if metas[0].ID != DefaultID || !metas[0].Builtin {
		t.Fatalf("got %+v", metas[0])
	}
	if metas[0].Colors != DefaultTheme().Colors {
		t.Errorf("expected the default colours, got %+v", metas[0].Colors)
	}
	if !strings.Contains(metas[0].PreviewCSS, "--dashi-bg:#ffffff;") {
		t.Errorf("expected preview CSS, got %q", metas[0].PreviewCSS)
	}
	if metas[1].ID != EmberID || !metas[1].Builtin || metas[1].Colors != EmberTheme().Colors {
		t.Fatalf("got %+v", metas[1])
	}
}

func TestListSortsByNameAndCarriesUsageAndPreview(t *testing.T) {
	s, _ := newTestStore(t)
	zebraTheme := testTheme("Zebra")
	zebraTheme.Background.Gradient = &Gradient{Direction: "to right", Light: []string{"#123456", "#654321"}}
	b, err := s.Create(zebraTheme)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	mustCreate(t, s, "apple")

	metas, err := s.List(map[string]int{b.ID: 3, DefaultID: 2})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(metas) != 4 {
		t.Fatalf("expected 2 themes plus the 2 builtins, got %d", len(metas))
	}
	var names []string
	for _, m := range metas {
		names = append(names, m.Name)
	}
	if strings.Join(names, ",") != "apple,Default,Ember,Zebra" {
		t.Fatalf("expected case-insensitive name order, got %v", names)
	}
	if metas[0].Builtin || metas[3].Builtin {
		t.Error("stored themes must not be flagged builtin")
	}
	if metas[1].UsedBy != 2 {
		t.Errorf("expected the builtin UsedBy 2, got %d", metas[1].UsedBy)
	}
	zebra := metas[3]
	if zebra.UsedBy != 3 {
		t.Errorf("expected UsedBy 3, got %d", zebra.UsedBy)
	}
	if !strings.Contains(zebra.PreviewCSS, "#123456") {
		t.Errorf("expected preview CSS, got %q", zebra.PreviewCSS)
	}
	if zebra.Colors != zebraTheme.Colors {
		t.Errorf("expected the meta to carry colours, got %+v", zebra.Colors)
	}
}

func TestListSkipsStoredThemeClaimingBuiltinID(t *testing.T) {
	s, dir := newTestStore(t)
	folder := filepath.Join(dir, "themes", "imposter")
	if err := os.MkdirAll(folder, 0o750); err != nil {
		t.Fatal(err)
	}
	raw := `{"id":"ember","name":"Imposter","background":{},"colors":{}}`
	if err := os.WriteFile(filepath.Join(folder, "theme.json"), []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	metas, err := s.List(nil)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(metas) != 2 || metas[0].Name != "Default" || metas[1].Name != "Ember" {
		t.Fatalf("a stored theme must not shadow a builtin, got %+v", metas)
	}
}

func TestIndexIsRebuiltOnStartup(t *testing.T) {
	dir := t.TempDir()
	pool, _ := images.NewStore(filepath.Join(dir, "pool"))
	themesDir := filepath.Join(dir, "themes")

	first := NewStore(themesDir, pool)
	created := mustCreate(t, first, "Persisted")

	second := NewStore(themesDir, pool)
	if _, err := second.Get(created.ID); err != nil {
		t.Fatalf("a fresh store must find the theme by id: %v", err)
	}
}

func TestAssetLifecycle(t *testing.T) {
	s, _ := newTestStore(t)
	created := mustCreate(t, s, "Assets")

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
	created := mustCreate(t, s, "Assets")

	cases := []string{"../escape.png", "/abs.png", "script.js", "theme.json", ""}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			if err := s.SaveAsset(created.ID, name, []byte("x")); err == nil {
				t.Fatalf("expected %q to be rejected", name)
			}
		})
	}
}

func TestListAssetsExcludesThemeJSON(t *testing.T) {
	s, _ := newTestStore(t)
	created := mustCreate(t, s, "Assets")
	list, err := s.ListAssets(created.ID)
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	for _, a := range list {
		if a == "theme.json" {
			t.Fatal("theme.json must not appear as an asset")
		}
	}
}

func TestLoadImageFromAssetAndPool(t *testing.T) {
	dir := t.TempDir()
	pool, _ := images.NewStore(filepath.Join(dir, "pool"))
	if err := pool.Save("shared.png", []byte("pool-bytes")); err != nil {
		t.Fatalf("seed pool: %v", err)
	}
	s := NewStore(filepath.Join(dir, "themes"), pool)
	created := mustCreate(t, s, "Loader")
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

// createOceanWithAsset creates "Ocean" with a custom palette and an image
// asset at img/a.png referenced by its background.
func createOceanWithAsset(t *testing.T, s *Store) Theme {
	t.Helper()
	th := testTheme("Ocean")
	th.Colors.Light.Primary = "#0077be"
	th.Colors.Dark.Surface = "#00334e80"
	th.Background.Image = &Image{Light: "asset:img/a.png", Fit: "cover", Position: "center", Repeat: "no-repeat"}
	created, err := s.Create(th)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.SaveAsset(created.ID, "img/a.png", []byte("ocean-bytes")); err != nil {
		t.Fatalf("save asset: %v", err)
	}
	return created
}

func TestCopyTheme(t *testing.T) {
	s, _ := newTestStore(t)
	orig := createOceanWithAsset(t, s)

	cp, err := s.Copy(orig.ID)
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	if cp.ID == orig.ID || cp.ID == "" {
		t.Fatalf("expected a new id, got %q (original %q)", cp.ID, orig.ID)
	}
	if cp.Name != "Copy of Ocean" {
		t.Errorf("Name = %q", cp.Name)
	}
	if cp.Colors != orig.Colors {
		t.Errorf("colours differ: %+v vs %+v", cp.Colors, orig.Colors)
	}
	data, _, err := s.GetAsset(cp.ID, "img/a.png")
	if err != nil || string(data) != "ocean-bytes" {
		t.Fatalf("copied asset: %q, %v", data, err)
	}
	stored, err := s.Get(cp.ID)
	if err != nil {
		t.Fatalf("the copy must be persisted: %v", err)
	}
	if stored.Background.Image == nil || stored.Background.Image.Light != "asset:img/a.png" {
		t.Errorf("background not copied: %+v", stored.Background)
	}
}

func TestCopyEmber(t *testing.T) {
	s, _ := newTestStore(t)
	cp, err := s.Copy(EmberID)
	if err != nil {
		t.Fatalf("copy ember: %v", err)
	}
	if cp.ID == EmberID || cp.Name != "Copy of Ember" || cp.Colors != EmberTheme().Colors {
		t.Fatalf("got %+v", cp)
	}
	if _, err := s.Update(cp); err != nil {
		t.Errorf("the copy must be editable: %v", err)
	}
}

func TestCopyDefault(t *testing.T) {
	s, _ := newTestStore(t)
	cp, err := s.Copy(DefaultID)
	if err != nil {
		t.Fatalf("copy default: %v", err)
	}
	if cp.ID == DefaultID || cp.ID == "" {
		t.Fatalf("expected a new id, got %q", cp.ID)
	}
	if cp.Name != "Copy of Default" {
		t.Errorf("Name = %q", cp.Name)
	}
	if cp.Colors != DefaultTheme().Colors {
		t.Errorf("expected default colours, got %+v", cp.Colors)
	}
	// The copy is a regular, editable theme.
	if _, err := s.Update(cp); err != nil {
		t.Errorf("the copy must be editable: %v", err)
	}
	list, err := s.ListAssets(cp.ID)
	if err != nil || len(list) != 0 {
		t.Fatalf("expected no assets, got %v, %v", list, err)
	}
	metas, _ := s.List(nil)
	for _, m := range metas {
		if m.ID == cp.ID && m.Builtin {
			t.Error("the copy must not be flagged builtin")
		}
	}
}

func TestCopyThemeSurvivesOriginalDelete(t *testing.T) {
	// Review Focus 3: the copy owns its own asset files.
	s, _ := newTestStore(t)
	orig := createOceanWithAsset(t, s)
	cp, err := s.Copy(orig.ID)
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	if err := s.Delete(orig.ID); err != nil {
		t.Fatalf("delete original: %v", err)
	}
	data, err := s.LoadImage(cp.ID, "asset:img/a.png")
	if err != nil || string(data) != "ocean-bytes" {
		t.Fatalf("copy lost its image: %q, %v", data, err)
	}
}

func TestCopyUnknownReturnsNotFound(t *testing.T) {
	s, _ := newTestStore(t)
	if _, err := s.Copy("zzzzzz"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if _, err := s.Copy("../x"); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}
}

func TestRollbackCreateRemovesFolderAndIndexEntry(t *testing.T) {
	s, _ := newTestStore(t)
	created := mustCreate(t, s, "ToBeRolledBack")

	folderPath, ok := s.themeDir(created.ID)
	if !ok {
		t.Fatalf("theme %s not in index", created.ID)
	}
	if _, err := os.Stat(folderPath); err != nil {
		t.Fatalf("expected folder to exist before rollback: %v", err)
	}
	if _, err := s.Get(created.ID); err != nil {
		t.Fatalf("expected Get to succeed before rollback: %v", err)
	}

	s.rollbackCreate(created.ID, folderPath)

	if _, err := os.Stat(folderPath); !os.IsNotExist(err) {
		t.Fatalf("expected folder to be removed, got %v", err)
	}
	if _, err := s.Get(created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after rollback, got %v", err)
	}
}
