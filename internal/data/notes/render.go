package notes

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/yuin/goldmark"
)

// GetHTML loads the note and renders it to HTML via goldmark. Uses a
// single Markdown instance — goldmark.Markdown is safe for concurrent
// use by documentation, and creating one per call wastes the extension
// setup cost.
func (s *Store) GetHTML(name string) (template.HTML, error) {
	raw, err := s.Get(name)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := goldmarkInstance.Convert([]byte(raw), &buf); err != nil {
		return "", fmt.Errorf("notes: render: %w", err)
	}
	return template.HTML(buf.String()), nil //nolint:gosec // goldmark output is HTML-safe by design
}

var goldmarkInstance = goldmark.New()
