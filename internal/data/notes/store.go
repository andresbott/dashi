// Package notes stores markdown documents shared across all dashboards.
package notes

import (
	"github.com/andresbott/dashi/internal/data"
)

var allowedExts = map[string]bool{".md": true}

// Store holds markdown notes in a flat directory. All names must end
// in .md. Content is handled as strings; rendering to HTML is provided
// by GetHTML in render.go.
type Store struct {
	fs *data.FsStore
}

// NewStore creates the directory if missing and returns a new store.
func NewStore(dir string) (*Store, error) {
	fs, err := data.NewFsStore(dir, allowedExts)
	if err != nil {
		return nil, err
	}
	return &Store{fs: fs}, nil
}

// List returns the notes in alphabetical order.
func (s *Store) List() ([]data.Item, error) { return s.fs.List() }

// Get returns the raw markdown content.
func (s *Store) Get(name string) (string, error) {
	b, err := s.fs.Get(name)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Save writes or overwrites a note.
func (s *Store) Save(name, content string) error {
	return s.fs.Save(name, []byte(content))
}

// Delete removes a note. Returns data.ErrNotFound if missing.
func (s *Store) Delete(name string) error { return s.fs.Delete(name) }
