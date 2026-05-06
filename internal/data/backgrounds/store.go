// Package backgrounds stores background images shared across all
// dashboards. API identical to images.Store; kept separate so future
// divergence (thumbnails, preview generation) doesn't force a rename.
package backgrounds

import (
	"mime"
	"path/filepath"

	"github.com/andresbott/dashi/internal/data"
)

var allowedExts = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
	".svg":  true,
}

type Store struct {
	fs *data.FsStore
}

func NewStore(dir string) (*Store, error) {
	fs, err := data.NewFsStore(dir, allowedExts)
	if err != nil {
		return nil, err
	}
	return &Store{fs: fs}, nil
}

func (s *Store) List() ([]data.Item, error) { return s.fs.List() }

func (s *Store) Get(name string) (content []byte, mimeType string, err error) {
	b, err := s.fs.Get(name)
	if err != nil {
		return nil, "", err
	}
	mimeType = mime.TypeByExtension(filepath.Ext(name))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return b, mimeType, nil
}

func (s *Store) Save(name string, content []byte) error { return s.fs.Save(name, content) }
func (s *Store) Delete(name string) error            { return s.fs.Delete(name) }
