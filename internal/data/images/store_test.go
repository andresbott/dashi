package images

import (
	"errors"
	"strings"
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

func TestImages_SaveGetDelete(t *testing.T) {
	s := newTestStore(t)
	pngBytes := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	if err := s.Save("hero.png", pngBytes); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, mime, err := s.Get("hero.png")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if string(got) != string(pngBytes) {
		t.Fatalf("bytes mismatch")
	}
	if !strings.HasPrefix(mime, "image/png") {
		t.Fatalf("mime: got %q, want image/png...", mime)
	}

	if err := s.Delete("hero.png"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestImages_RejectsMd(t *testing.T) {
	s := newTestStore(t)
	if err := s.Save("a.md", []byte("x")); !errors.Is(err, data.ErrExtNotAllowed) {
		t.Fatalf("got %v, want ErrExtNotAllowed", err)
	}
}

func TestImages_GetMissing(t *testing.T) {
	s := newTestStore(t)
	_, _, err := s.Get("missing.png")
	if !errors.Is(err, data.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}
