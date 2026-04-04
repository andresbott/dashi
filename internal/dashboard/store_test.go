package dashboard

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore_Create(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	d := Dashboard{
		Name: "Test Dashboard",
		Icon: "ti-dashboard",
		Container: Container{
			MaxWidth:        "100%",
			VerticalAlign:   "top",
			HorizontalAlign: "center",
		},
		Rows: []Row{},
	}

	created, err := store.Create(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if created.Name != "Test Dashboard" {
		t.Fatalf("expected name 'Test Dashboard', got %q", created.Name)
	}

	// Directory should exist
	dashDir := filepath.Join(dir, created.ID)
	info, err := os.Stat(dashDir)
	if err != nil {
		t.Fatalf("expected directory %s to exist: %v", dashDir, err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %s to be a directory", dashDir)
	}

	// dashboard.json should exist inside
	fpath := filepath.Join(dashDir, "dashboard.json")
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		t.Fatalf("expected file %s to exist", fpath)
	}
}

func TestStore_Get(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	created, err := store.Create(Dashboard{Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Rows: []Row{}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := store.Get(created.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != "Test" {
		t.Fatalf("expected name 'Test', got %q", got.Name)
	}
}

func TestStore_Get_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, err := store.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent dashboard")
	}
}

func TestStore_List(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, _ = store.Create(Dashboard{Name: "A", Icon: "ti-a", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Rows: []Row{}})
	_, _ = store.Create(Dashboard{Name: "B", Icon: "ti-b", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Rows: []Row{
		{ID: "r1", Height: "100px", Width: "100%", Widgets: []Widget{{ID: "w1", Type: "placeholder", Title: "W", Width: 12}}},
	}})

	list, err := store.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 dashboards, got %d", len(list))
	}
	for _, m := range list {
		if m.ID == "" || m.Name == "" {
			t.Fatal("expected non-empty ID and Name in list")
		}
	}
}

func TestStore_Update(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	created, _ := store.Create(Dashboard{Name: "Old", Icon: "ti-old", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Rows: []Row{}})

	created.Name = "New"
	created.Rows = []Row{
		{ID: "r1", Height: "200px", Width: "100%", Widgets: []Widget{}},
	}

	updated, err := store.Update(created)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "New" {
		t.Fatalf("expected name 'New', got %q", updated.Name)
	}

	got, _ := store.Get(created.ID)
	if len(got.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(got.Rows))
	}
}

func TestStore_Update_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, err := store.Update(Dashboard{ID: "nonexistent", Name: "X", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}})
	if err == nil {
		t.Fatal("expected error updating nonexistent dashboard")
	}
}

func TestStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	created, _ := store.Create(Dashboard{Name: "Bye", Icon: "ti-bye", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Rows: []Row{}})

	err := store.Delete(created.ID)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err = store.Get(created.ID)
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestStore_Delete_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	err := store.Delete("nonexistent")
	if err == nil {
		t.Fatal("expected error deleting nonexistent dashboard")
	}
}

func TestIsAllowedAssetExt(t *testing.T) {
	allowed := []string{".png", ".jpg", ".jpeg", ".svg", ".webp", ".css"}
	for _, ext := range allowed {
		if !isAllowedAssetExt("file" + ext) {
			t.Errorf("expected %s to be allowed", ext)
		}
	}
	disallowed := []string{".json", ".exe", ".html", ".js", ".go", ""}
	for _, ext := range disallowed {
		if isAllowedAssetExt("file" + ext) {
			t.Errorf("expected %s to be disallowed", ext)
		}
	}
}

func TestValidateAssetPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"simple file", "logo.png", false},
		{"nested path", "icons/16x16/home.png", false},
		{"css file", "styles/custom.css", false},
		{"path traversal dotdot", "../etc/passwd", true},
		{"path traversal middle", "icons/../../etc/passwd", true},
		{"absolute path", "/etc/passwd", true},
		{"dashboard.json", "dashboard.json", true},
		{"nested dashboard.json", "sub/dashboard.json", true},
		{"disallowed extension", "script.js", true},
		{"empty path", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAssetPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAssetPath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestStore_SaveAndGetAsset(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, err := store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Rows: []Row{}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	content := []byte("PNG fake content")
	err = store.SaveAsset("test01", "logo.png", content)
	if err != nil {
		t.Fatalf("save asset: %v", err)
	}

	data, mimeType, err := store.GetAsset("test01", "logo.png")
	if err != nil {
		t.Fatalf("get asset: %v", err)
	}
	if string(data) != "PNG fake content" {
		t.Fatalf("expected content 'PNG fake content', got %q", string(data))
	}
	if mimeType != "image/png" {
		t.Fatalf("expected mime 'image/png', got %q", mimeType)
	}
}

func TestStore_SaveAsset_NestedPath(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Rows: []Row{}})

	err := store.SaveAsset("test01", "icons/16x16/home.png", []byte("icon"))
	if err != nil {
		t.Fatalf("save nested asset: %v", err)
	}

	data, _, err := store.GetAsset("test01", "icons/16x16/home.png")
	if err != nil {
		t.Fatalf("get nested asset: %v", err)
	}
	if string(data) != "icon" {
		t.Fatalf("expected 'icon', got %q", string(data))
	}
}

func TestStore_SaveAsset_Rejected(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Rows: []Row{}})

	err := store.SaveAsset("test01", "../escape.png", []byte("bad"))
	if err == nil {
		t.Fatal("expected error for path traversal")
	}

	err = store.SaveAsset("test01", "dashboard.json", []byte("bad"))
	if err == nil {
		t.Fatal("expected error for dashboard.json")
	}

	err = store.SaveAsset("test01", "script.js", []byte("bad"))
	if err == nil {
		t.Fatal("expected error for disallowed extension")
	}
}

func TestStore_GetAsset_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Rows: []Row{}})

	_, _, err := store.GetAsset("test01", "nonexistent.png")
	if err == nil {
		t.Fatal("expected error for nonexistent asset")
	}
}

func TestStore_DeleteAsset(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Rows: []Row{}})
	_ = store.SaveAsset("test01", "logo.png", []byte("content"))

	err := store.DeleteAsset("test01", "logo.png")
	if err != nil {
		t.Fatalf("delete asset: %v", err)
	}

	_, _, err = store.GetAsset("test01", "logo.png")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestStore_DeleteAsset_CleansEmptyDirs(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Rows: []Row{}})
	_ = store.SaveAsset("test01", "icons/16x16/home.png", []byte("icon"))

	err := store.DeleteAsset("test01", "icons/16x16/home.png")
	if err != nil {
		t.Fatalf("delete nested asset: %v", err)
	}

	iconsDir := filepath.Join(dir, "test01", "icons")
	if _, err := os.Stat(iconsDir); !os.IsNotExist(err) {
		t.Fatalf("expected icons directory to be removed, but it still exists")
	}
}

func TestStore_ListAssets(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Rows: []Row{}})
	_ = store.SaveAsset("test01", "logo.png", []byte("a"))
	_ = store.SaveAsset("test01", "custom.css", []byte("b"))
	_ = store.SaveAsset("test01", "icons/16x16/home.png", []byte("c"))

	assets, err := store.ListAssets("test01")
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(assets) != 3 {
		t.Fatalf("expected 3 assets, got %d: %v", len(assets), assets)
	}

	for _, a := range assets {
		if a == "dashboard.json" {
			t.Fatal("dashboard.json should not appear in asset list")
		}
	}
}

func TestStore_ListAssets_Empty(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Rows: []Row{}})

	assets, err := store.ListAssets("test01")
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(assets) != 0 {
		t.Fatalf("expected 0 assets, got %d", len(assets))
	}
}
