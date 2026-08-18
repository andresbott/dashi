package backgrounds

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

func TestBackgrounds_RoundTrip(t *testing.T) {
	s := newTestStore(t)
	if err := s.Save("sunset.jpg", []byte{0xFF, 0xD8, 0xFF}); err != nil {
		t.Fatalf("save: %v", err)
	}
	b, mime, err := s.Get("sunset.jpg")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(b) != 3 || mime != "image/jpeg" {
		t.Fatalf("mismatch: bytes=%v mime=%q", b, mime)
	}
}

func TestBackgrounds_RejectsMd(t *testing.T) {
	s := newTestStore(t)
	if err := s.Save("a.md", []byte("x")); !errors.Is(err, data.ErrExtNotAllowed) {
		t.Fatalf("got %v, want ErrExtNotAllowed", err)
	}
}

func TestBackgrounds_ListAndDelete(t *testing.T) {
	s := newTestStore(t)
	if err := s.Save("a.png", []byte{0x89, 'P', 'N', 'G'}); err != nil {
		t.Fatal(err)
	}
	items, err := s.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 || items[0].Name != "a.png" {
		t.Fatalf("unexpected items: %+v", items)
	}
	if err := s.Delete("a.png"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := s.Delete("a.png"); !errors.Is(err, data.ErrNotFound) {
		t.Fatalf("delete missing: got %v, want ErrNotFound", err)
	}
}
