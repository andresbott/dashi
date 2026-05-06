package notes

import (
	"errors"
	"strings"
	"testing"

	"github.com/andresbott/dashi/internal/data"
)

func TestGetHTML(t *testing.T) {
	s := newTestStore(t)
	if err := s.Save("a.md", "# hello\n"); err != nil {
		t.Fatal(err)
	}
	html, err := s.GetHTML("a.md")
	if err != nil {
		t.Fatalf("GetHTML: %v", err)
	}
	if !strings.Contains(string(html), "<h1>hello</h1>") {
		t.Fatalf("expected <h1>hello</h1>, got %q", html)
	}
}

func TestGetHTML_Missing(t *testing.T) {
	s := newTestStore(t)
	_, err := s.GetHTML("missing.md")
	if !errors.Is(err, data.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}
