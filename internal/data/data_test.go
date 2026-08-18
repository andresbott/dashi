package data

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateName(t *testing.T) {
	allowed := map[string]bool{".md": true}

	cases := []struct {
		name string
		in   string
		want error
	}{
		{"empty", "", ErrInvalidName},
		{"whitespace only", "   ", ErrInvalidName},
		{"path separator /", "a/b.md", ErrInvalidName},
		{"path separator backslash", `a\b.md`, ErrInvalidName},
		{"traversal", "../x.md", ErrInvalidName},
		{"leading dot", ".hidden.md", ErrInvalidName},
		{"null byte", "a\x00b.md", ErrInvalidName},
		{"control char", "a\x01b.md", ErrInvalidName},
		{"bad ext", "notes.txt", ErrExtNotAllowed},
		{"no ext", "notes", ErrExtNotAllowed},
		{"uppercase ext accepted", "notes.MD", nil},
		{"overlong", strings.Repeat("a", 300) + ".md", ErrInvalidName},
		{"happy path", "todo.md", nil},
		{"unicode nfc ok", "café.md", nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := validateName(c.in, allowed)
			if c.want == nil && got != nil {
				t.Fatalf("unexpected error: %v", got)
			}
			if c.want != nil && !errors.Is(got, c.want) {
				t.Fatalf("got %v, want %v", got, c.want)
			}
		})
	}
}

func newTestStore(t *testing.T, allowed map[string]bool) *FsStore {
	t.Helper()
	dir := t.TempDir()
	s, err := NewFsStore(dir, allowed)
	if err != nil {
		t.Fatalf("NewFsStore: %v", err)
	}
	return s
}

func TestFsStore_SaveGetDelete(t *testing.T) {
	s := newTestStore(t, map[string]bool{".md": true})

	if err := s.Save("a.md", []byte("hello")); err != nil {
		t.Fatalf("save: %v", err)
	}

	data, err := s.Get("a.md")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("got %q, want %q", data, "hello")
	}

	if err := s.Delete("a.md"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err = s.Get("a.md")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("get after delete: got %v, want ErrNotFound", err)
	}
}

func TestFsStore_GetMissing(t *testing.T) {
	s := newTestStore(t, map[string]bool{".md": true})
	if _, err := s.Get("missing.md"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestFsStore_DeleteMissing(t *testing.T) {
	s := newTestStore(t, map[string]bool{".md": true})
	if err := s.Delete("missing.md"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestFsStore_SaveRejectsInvalidName(t *testing.T) {
	s := newTestStore(t, map[string]bool{".md": true})
	if err := s.Save("a/b.md", []byte("x")); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("got %v, want ErrInvalidName", err)
	}
	if err := s.Save("a.txt", []byte("x")); !errors.Is(err, ErrExtNotAllowed) {
		t.Fatalf("got %v, want ErrExtNotAllowed", err)
	}
}

func TestFsStore_SaveAtomic_NoTempLeft(t *testing.T) {
	s := newTestStore(t, map[string]bool{".md": true})
	if err := s.Save("a.md", []byte("x")); err != nil {
		t.Fatalf("save: %v", err)
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Fatalf("temp file left behind: %s", e.Name())
		}
	}
}

func TestFsStore_List(t *testing.T) {
	s := newTestStore(t, map[string]bool{".md": true})

	items, err := s.List()
	if err != nil {
		t.Fatalf("list empty: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}

	if err := s.Save("a.md", []byte("aa")); err != nil {
		t.Fatalf("save a: %v", err)
	}
	if err := s.Save("b.md", []byte("bbbb")); err != nil {
		t.Fatalf("save b: %v", err)
	}

	items, err = s.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	byName := map[string]Item{}
	for _, it := range items {
		byName[it.Name] = it
	}
	if byName["a.md"].Size != 2 {
		t.Fatalf("a.md size: got %d, want 2", byName["a.md"].Size)
	}
	if byName["b.md"].Size != 4 {
		t.Fatalf("b.md size: got %d, want 4", byName["b.md"].Size)
	}
	if byName["a.md"].ModTime.IsZero() {
		t.Fatalf("a.md modtime is zero")
	}
}

func TestFsStore_List_SkipsTempAndDirs(t *testing.T) {
	s := newTestStore(t, map[string]bool{".md": true})
	// Create a stray .tmp and a subdir manually; list must skip them.
	if err := os.WriteFile(filepath.Join(s.dir, "stale.tmp"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(s.dir, "sub"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := s.Save("real.md", []byte("x")); err != nil {
		t.Fatal(err)
	}
	items, err := s.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 || items[0].Name != "real.md" {
		t.Fatalf("unexpected items: %+v", items)
	}
}
