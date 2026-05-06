package notes

import (
	"errors"
	"testing"

	"github.com/andresbott/dashi/internal/data"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

func TestNotes_SaveGetDelete(t *testing.T) {
	s := newTestStore(t)
	if err := s.Save("todo.md", "# hello"); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := s.Get("todo.md")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got != "# hello" {
		t.Fatalf("got %q, want %q", got, "# hello")
	}

	if err := s.Delete("todo.md"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err = s.Get("todo.md")
	if !errors.Is(err, data.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestNotes_RejectsNonMd(t *testing.T) {
	s := newTestStore(t)
	if err := s.Save("todo.txt", "x"); !errors.Is(err, data.ErrExtNotAllowed) {
		t.Fatalf("got %v, want ErrExtNotAllowed", err)
	}
}

func TestNotes_List(t *testing.T) {
	s := newTestStore(t)
	_ = s.Save("b.md", "b")
	_ = s.Save("a.md", "aa")

	items, err := s.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 2 || items[0].Name != "a.md" || items[1].Name != "b.md" {
		t.Fatalf("unexpected items: %+v", items)
	}
}
