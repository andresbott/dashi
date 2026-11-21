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

// The markdown widget emits GetHTML output as template.HTML (unescaped), so
// goldmark's safe defaults are the XSS boundary: raw HTML is dropped and
// dangerous link schemes lose their href. This pins those defaults.
func TestGetHTML_StripsUnsafeContent(t *testing.T) {
	s := newTestStore(t)
	md := "[click](javascript:alert(1))\n\n<script>alert(2)</script>\n\n<img src=x onerror=alert(3)>\n\n[ok](https://example.com)\n"
	if err := s.Save("x.md", md); err != nil {
		t.Fatal(err)
	}
	html, err := s.GetHTML("x.md")
	if err != nil {
		t.Fatalf("GetHTML: %v", err)
	}
	out := string(html)
	for _, bad := range []string{"javascript:", "<script", "onerror"} {
		if strings.Contains(out, bad) {
			t.Errorf("output contains %q: %s", bad, out)
		}
	}
	if !strings.Contains(out, `href="https://example.com"`) {
		t.Errorf("safe link missing: %s", out)
	}
}

func TestGetHTML_Missing(t *testing.T) {
	s := newTestStore(t)
	_, err := s.GetHTML("missing.md")
	if !errors.Is(err, data.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}
