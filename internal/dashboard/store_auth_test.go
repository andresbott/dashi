package dashboard

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore_GetAuth_NoFile(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})

	auth, err := store.GetAuth("test01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if auth != nil {
		t.Fatal("expected nil auth for dashboard without auth.json")
	}
}

func TestStore_SetAuth(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})

	err := store.SetAuth("test01", "viewer", "$2a$10$fakehashvalue")
	if err != nil {
		t.Fatalf("set auth: %v", err)
	}

	auth, err := store.GetAuth("test01")
	if err != nil {
		t.Fatalf("get auth: %v", err)
	}
	if auth == nil {
		t.Fatal("expected non-nil auth")
	}
	if auth.Username != "viewer" {
		t.Fatalf("expected username 'viewer', got %q", auth.Username)
	}
	if auth.PasswordHash != "$2a$10$fakehashvalue" {
		t.Fatalf("expected password hash, got %q", auth.PasswordHash)
	}
}

func TestStore_SetAuth_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	err := store.SetAuth("nonexistent", "user", "hash")
	if err == nil {
		t.Fatal("expected error for nonexistent dashboard")
	}
}

func TestStore_SetAuth_InvalidID(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	err := store.SetAuth("INVALID-ID", "user", "hash")
	if err != ErrInvalidID {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}
}

func TestStore_DeleteAuth(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})
	_ = store.SetAuth("test01", "viewer", "$2a$10$fakehash")

	err := store.DeleteAuth("test01")
	if err != nil {
		t.Fatalf("delete auth: %v", err)
	}

	auth, err := store.GetAuth("test01")
	if err != nil {
		t.Fatalf("get auth after delete: %v", err)
	}
	if auth != nil {
		t.Fatal("expected nil auth after delete")
	}
}

func TestStore_DeleteAuth_NoFile(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})

	err := store.DeleteAuth("test01")
	if err != nil {
		t.Fatalf("expected no error deleting nonexistent auth: %v", err)
	}
}

func TestStore_DeleteAuth_InvalidID(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	err := store.DeleteAuth("INVALID-ID")
	if err != ErrInvalidID {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}
}

func TestStore_GetAuth_InvalidID(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, err := store.GetAuth("INVALID-ID")
	if err != ErrInvalidID {
		t.Fatalf("expected ErrInvalidID, got %v", err)
	}
}

func TestStore_SetAuth_Overwrite(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	_, _ = store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})

	_ = store.SetAuth("test01", "user1", "hash1")
	_ = store.SetAuth("test01", "user2", "hash2")

	auth, _ := store.GetAuth("test01")
	if auth.Username != "user2" {
		t.Fatalf("expected username 'user2', got %q", auth.Username)
	}
}

func TestStore_Delete_RemovesAuth(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	created, _ := store.Create(Dashboard{ID: "test01", Name: "Test", Icon: "ti-home", Container: Container{MaxWidth: "100%", VerticalAlign: "top", HorizontalAlign: "center"}, Pages: []Page{}})
	_ = store.SetAuth(created.ID, "viewer", "hash")

	err := store.Delete(created.ID)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}

	// Auth file should be gone along with the whole directory
	authPath := filepath.Join(dir, "test", authFile)
	if _, err := os.Stat(authPath); !os.IsNotExist(err) {
		t.Fatal("expected auth.json to be removed with dashboard")
	}
}
