package dashboard

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
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
		Pages: []Page{},
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

	// Directory should exist with snake_case name
	dashDir := filepath.Join(dir, "test_dashboard")
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

	created, err := store.Create(Dashboard{Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})
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

	_, _ = store.Create(Dashboard{Name: "A", Icon: "ti-a", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})
	_, _ = store.Create(Dashboard{Name: "B", Icon: "ti-b", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{
		{Rows: []Row{
			{ID: "r1", Height: "100px", Width: "100%", Widgets: []Widget{{ID: "w1", Type: "placeholder", Title: "W", Width: 12}}},
		}},
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

	created, _ := store.Create(Dashboard{Name: "Old", Icon: "ti-old", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})

	created.Name = "New"
	created.Pages = []Page{
		{Rows: []Row{
			{ID: "r1", Height: "200px", Width: "100%", Widgets: []Widget{}},
		}},
	}

	updated, err := store.Update(created)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "New" {
		t.Fatalf("expected name 'New', got %q", updated.Name)
	}

	got, _ := store.Get(created.ID)
	if len(got.Pages) != 1 || len(got.Pages[0].Rows) != 1 {
		t.Fatalf("expected 1 page with 1 row, got %d pages", len(got.Pages))
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

	created, _ := store.Create(Dashboard{Name: "Bye", Icon: "ti-bye", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})

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

	_, err := store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})
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

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})

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

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})

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

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})

	_, _, err := store.GetAsset("test01", "nonexistent.png")
	if err == nil {
		t.Fatal("expected error for nonexistent asset")
	}
}

func TestStore_DeleteAsset(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})
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

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})
	_ = store.SaveAsset("test01", "icons/16x16/home.png", []byte("icon"))

	err := store.DeleteAsset("test01", "icons/16x16/home.png")
	if err != nil {
		t.Fatalf("delete nested asset: %v", err)
	}

	iconsDir := filepath.Join(dir, "test", "icons")
	if _, err := os.Stat(iconsDir); !os.IsNotExist(err) {
		t.Fatalf("expected icons directory to be removed, but it still exists")
	}
}

func TestStore_ListAssets(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})
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

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})

	assets, err := store.ListAssets("test01")
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(assets) != 0 {
		t.Fatalf("expected 0 assets, got %d", len(assets))
	}
}

func TestIsValidID(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		valid bool
	}{
		{"lowercase alphanumeric", "abc123", true},
		{"with preview suffix", "abc123-prev", true},
		{"uppercase letter", "Abc123", false},
		{"special char dash", "abc-123", false},
		{"special char underscore", "abc_123", false},
		{"empty string", "", false},
		{"only preview suffix", "-prev", false},
		{"space", "abc 123", false},
		{"numbers only", "123456", true},
		{"letters only", "abcdef", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidID(tt.id)
			if got != tt.valid {
				t.Errorf("isValidID(%q) = %v, want %v", tt.id, got, tt.valid)
			}
		})
	}
}

func TestIsPreviewID(t *testing.T) {
	tests := []struct {
		id      string
		preview bool
	}{
		{"abc123-prev", true},
		{"abc123", false},
		{"", false},
		{"prev", false},
		{"test-prev-prev", true},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got := isPreviewID(tt.id)
			if got != tt.preview {
				t.Errorf("isPreviewID(%q) = %v, want %v", tt.id, got, tt.preview)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello World", "hello_world"},
		{"Test Dashboard", "test_dashboard"},
		{"multiple   spaces", "multiple_spaces"},
		{"UPPERCASE", "uppercase"},
		{"CamelCase", "camelcase"},
		{"with-dashes", "with_dashes"},
		{"with_underscores", "with_underscores"},
		{"special!@#chars", "special_chars"},
		{"", "dashboard"},
		{"   ", "dashboard"},
		{"123numbers456", "123numbers456"},
		{"Café Münchën", "cafe_munchen"},
		{"日本語", "dashboard"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toSnakeCase(tt.input)
			if got != tt.want {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestStore_GetCustomCSS(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	created, _ := store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})

	css := store.GetCustomCSS(created.ID)
	if css != "" {
		t.Fatalf("expected empty CSS for new dashboard, got %q", css)
	}

	cssContent := []byte(".custom { color: red; }")
	dashDir := store.dashDir(created.ID)
	err := os.WriteFile(filepath.Join(dashDir, "custom.css"), cssContent, 0o600)
	if err != nil {
		t.Fatalf("write custom.css: %v", err)
	}

	css = store.GetCustomCSS(created.ID)
	if css != string(cssContent) {
		t.Fatalf("expected CSS %q, got %q", string(cssContent), css)
	}
}

func TestStore_GetCustomCSS_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	css := store.GetCustomCSS("nonexistent")
	if css != "" {
		t.Fatalf("expected empty CSS for nonexistent dashboard, got %q", css)
	}
}

func TestStore_DeletePreviews(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, _ = store.Create(Dashboard{ID: "regular1", Name: "Regular 1", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})
	_, _ = store.Create(Dashboard{ID: "regular2", Name: "Regular 2", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})
	_, _ = store.Create(Dashboard{ID: "abc123-prev", Name: "Preview 1", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})
	_, _ = store.Create(Dashboard{ID: "def456-prev", Name: "Preview 2", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})

	count, err := store.DeletePreviews()
	if err != nil {
		t.Fatalf("delete previews: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 previews deleted, got %d", count)
	}

	list, _ := store.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 regular dashboards remaining, got %d", len(list))
	}

	_, err = store.Get("abc123-prev")
	if err == nil {
		t.Fatal("expected preview dashboard to be deleted")
	}
}

func TestStore_DeletePreviews_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	count, err := store.DeletePreviews()
	if err != nil {
		t.Fatalf("delete previews on empty dir: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 previews deleted, got %d", count)
	}
}

func TestStore_DeletePreviews_NoPreviewsDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nonexistent")
	store := NewStore(dir)

	count, err := store.DeletePreviews()
	if err != nil {
		t.Fatalf("expected no error for nonexistent dir, got %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 previews deleted, got %d", count)
	}
}

func TestStore_Get_InvalidID(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, err := store.Get("INVALID-ID")
	if err != ErrInvalidID {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}
}

func TestStore_Update_InvalidID(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, err := store.Update(Dashboard{ID: "INVALID-ID", Name: "Test", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}})
	if err != ErrInvalidID {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}
}

func TestStore_Delete_InvalidID(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	err := store.Delete("INVALID-ID")
	if err != ErrInvalidID {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}
}

func TestStore_SaveAsset_InvalidID(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	err := store.SaveAsset("INVALID-ID", "test.png", []byte("data"))
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestStore_SaveAsset_DashboardNotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	err := store.SaveAsset("validid", "test.png", []byte("data"))
	if err == nil {
		t.Fatal("expected error for nonexistent dashboard")
	}
}

func TestStore_GetAsset_InvalidID(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, _, err := store.GetAsset("INVALID-ID", "test.png")
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestStore_DeleteAsset_InvalidID(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	err := store.DeleteAsset("INVALID-ID", "test.png")
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestStore_ListAssets_InvalidID(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, err := store.ListAssets("INVALID-ID")
	if err == nil {
		t.Fatal("expected error for invalid ID")
	}
}

func TestStore_ListAssets_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, err := store.ListAssets("validid")
	if err == nil {
		t.Fatal("expected error for nonexistent dashboard")
	}
}

func TestStore_UniqueFolder(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	created1, _ := store.Create(Dashboard{Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})
	created2, _ := store.Create(Dashboard{Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})

	dir1 := store.dashDir(created1.ID)
	dir2 := store.dashDir(created2.ID)

	if dir1 == dir2 {
		t.Fatal("expected unique directories for same-named dashboards")
	}

	base1 := filepath.Base(dir1)
	base2 := filepath.Base(dir2)

	if base1 != "test" {
		t.Errorf("expected first folder to be 'test', got %q", base1)
	}
	if base2 != "test_2" {
		t.Errorf("expected second folder to be 'test_2', got %q", base2)
	}
}

func TestStore_List_SkipsPreviews(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, _ = store.Create(Dashboard{ID: "regular", Name: "Regular", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})
	_, _ = store.Create(Dashboard{ID: "preview-prev", Name: "Preview", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})

	list, err := store.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("expected 1 dashboard (previews should be skipped), got %d", len(list))
	}

	if list[0].ID == "preview-prev" {
		t.Fatal("preview dashboard should not appear in list")
	}
}

func makeZip(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, data := range files {
		f, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip create %s: %v", name, err)
		}
		if _, err := f.Write(data); err != nil {
			t.Fatalf("zip write %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

func TestStore_ImportZip(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	dashJSON, _ := json.Marshal(Dashboard{
		ID:        "oldid1",
		Name:      "Imported",
		Icon:      "ti-home",
		Default:   true,
		Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"},
		Pages:     []Page{{Rows: []Row{}}},
	})

	zipData := makeZip(t, map[string][]byte{
		"dashboard.json": dashJSON,
		"bg.jpg":         []byte("fake jpg"),
		"icons/logo.png": []byte("fake png"),
	})

	created, err := store.ImportZip(zipData)
	if err != nil {
		t.Fatalf("import zip: %v", err)
	}
	if created.ID == "oldid1" {
		t.Fatal("expected a new ID, got the old one")
	}
	if created.Name != "Imported" {
		t.Fatalf("expected name 'Imported', got %q", created.Name)
	}
	if created.Default {
		t.Fatal("expected default to be false after import")
	}

	// Verify dashboard is readable
	got, err := store.Get(created.ID)
	if err != nil {
		t.Fatalf("get imported: %v", err)
	}
	if len(got.Pages) != 1 {
		t.Fatalf("expected 1 page, got %d", len(got.Pages))
	}

	// Verify assets were extracted
	assets, err := store.ListAssets(created.ID)
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(assets) != 2 {
		t.Fatalf("expected 2 assets, got %d: %v", len(assets), assets)
	}
}

func TestStore_ImportZip_NoDashboardJSON(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	zipData := makeZip(t, map[string][]byte{
		"bg.jpg": []byte("fake jpg"),
	})

	_, err := store.ImportZip(zipData)
	if err == nil {
		t.Fatal("expected error for zip without dashboard.json")
	}
}

func TestStore_ImportZip_EmptyName(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	dashJSON, _ := json.Marshal(Dashboard{
		Name:      "",
		Container: Container{MaxWidth: "100%"},
		Pages:     []Page{},
	})

	zipData := makeZip(t, map[string][]byte{
		"dashboard.json": dashJSON,
	})

	_, err := store.ImportZip(zipData)
	if err == nil {
		t.Fatal("expected error for empty dashboard name")
	}
}

func TestStore_ImportZip_InvalidZip(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, err := store.ImportZip([]byte("not a zip"))
	if err == nil {
		t.Fatal("expected error for invalid zip data")
	}
}

func TestStore_ImportZip_SkipsInvalidAssets(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	dashJSON, _ := json.Marshal(Dashboard{
		Name:      "Test",
		Container: Container{MaxWidth: "100%"},
		Pages:     []Page{},
	})

	zipData := makeZip(t, map[string][]byte{
		"dashboard.json": dashJSON,
		"valid.png":      []byte("png"),
		"script.js":      []byte("bad"),
		"../escape.png":  []byte("bad"),
	})

	created, err := store.ImportZip(zipData)
	if err != nil {
		t.Fatalf("import zip: %v", err)
	}

	assets, _ := store.ListAssets(created.ID)
	if len(assets) != 1 {
		t.Fatalf("expected 1 valid asset, got %d: %v", len(assets), assets)
	}
}

func TestStore_ExportZip(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	created, err := store.Create(Dashboard{
		ID:        "export01",
		Name:      "Exportable",
		Icon:      "ti-home",
		Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"},
		Pages:     []Page{{Rows: []Row{}}},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := store.SaveAsset(created.ID, "logo.png", []byte("pngbytes")); err != nil {
		t.Fatalf("save asset: %v", err)
	}
	if err := store.SaveAsset(created.ID, "icons/home.png", []byte("iconbytes")); err != nil {
		t.Fatalf("save nested asset: %v", err)
	}
	if err := store.SetAuth(created.ID, "viewer", "$2a$10$fakehash"); err != nil {
		t.Fatalf("set auth: %v", err)
	}

	var buf bytes.Buffer
	if err := store.ExportZip(created.ID, &buf); err != nil {
		t.Fatalf("export zip: %v", err)
	}

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}

	got := map[string]string{}
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s: %v", f.Name, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read %s: %v", f.Name, err)
		}
		got[filepath.ToSlash(f.Name)] = string(data)
	}

	if _, ok := got["dashboard.json"]; !ok {
		t.Fatal("expected dashboard.json in zip")
	}
	if got["logo.png"] != "pngbytes" {
		t.Fatalf("logo.png content mismatch: %q", got["logo.png"])
	}
	if got["icons/home.png"] != "iconbytes" {
		t.Fatalf("icons/home.png content mismatch: %q", got["icons/home.png"])
	}
	if _, ok := got["auth.json"]; ok {
		t.Fatal("auth.json must not be included in export")
	}
}

func TestStore_ExportZip_InvalidID(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	err := store.ExportZip("BAD_ID!", &bytes.Buffer{})
	if err != ErrInvalidID {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}
}

func TestStore_ExportZip_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	err := store.ExportZip("missing01", &bytes.Buffer{})
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStore_ExportZip_ImportRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	original, err := store.Create(Dashboard{
		ID:        "roundtrip",
		Name:      "Round Trip",
		Icon:      "ti-home",
		Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"},
		Pages:     []Page{{Rows: []Row{}}},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := store.SaveAsset(original.ID, "bg.jpg", []byte("jpegdata")); err != nil {
		t.Fatalf("save asset: %v", err)
	}

	var buf bytes.Buffer
	if err := store.ExportZip(original.ID, &buf); err != nil {
		t.Fatalf("export: %v", err)
	}

	imported, err := store.ImportZip(buf.Bytes())
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if imported.Name != original.Name {
		t.Fatalf("name mismatch: got %q want %q", imported.Name, original.Name)
	}
	if imported.ID == original.ID {
		t.Fatal("expected imported dashboard to get a new ID")
	}

	assets, err := store.ListAssets(imported.ID)
	if err != nil {
		t.Fatalf("list assets: %v", err)
	}
	if len(assets) != 1 || assets[0] != "bg.jpg" {
		t.Fatalf("expected [bg.jpg], got %v", assets)
	}
}
