// Package bookmarks stores YAML bookmark lists shared across all
// dashboards. A bookmarkgroup widget can reference one by name instead of
// keeping its bookmarks in the dashboard.
//
// File format:
//
//	bookmarks:
//	  - title: GitHub
//	    url: https://github.com
//	    icon: ti-brand-github
//	    subtitle: Code hosting
package bookmarks

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/andresbott/dashi/internal/data"
	"gopkg.in/yaml.v3"
)

var allowedExts = map[string]bool{".yaml": true, ".yml": true}

// CacheTTL is how long a parsed file is served from memory before it is
// read again, so edits made on disk show up within this time. Saves and
// deletes through the Store take effect at once.
const CacheTTL = 10 * time.Minute

// ErrInvalidFile is returned by Save (and Bookmarks) when the content is not
// a valid bookmark list.
var ErrInvalidFile = errors.New("bookmarks: invalid file")

// Item is one bookmark of a file, with the same fields as a bookmark kept
// in a dashboard.
type Item struct {
	URL      string `yaml:"url" json:"url"`
	Icon     string `yaml:"icon" json:"icon"`
	Title    string `yaml:"title" json:"title"`
	Subtitle string `yaml:"subtitle" json:"subtitle"`
}

type file struct {
	Bookmarks []Item `yaml:"bookmarks"`
}

type cached struct {
	items    []Item
	err      error
	loadedAt time.Time
}

// Store holds bookmark files in a flat directory. All names must end in
// .yaml or .yml.
type Store struct {
	fs  *data.FsStore
	now func() time.Time

	mu    sync.Mutex
	cache map[string]cached
}

// NewStore creates the directory if missing and returns a new store.
func NewStore(dir string) (*Store, error) {
	fs, err := data.NewFsStore(dir, allowedExts)
	if err != nil {
		return nil, err
	}
	return &Store{fs: fs, now: time.Now, cache: map[string]cached{}}, nil
}

// List returns the bookmark files in alphabetical order.
func (s *Store) List() ([]data.Item, error) { return s.fs.List() }

// Get returns the raw YAML content.
func (s *Store) Get(name string) (string, error) {
	b, err := s.fs.Get(name)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Save validates content as a bookmark list and writes or overwrites the
// file. Invalid content is rejected with ErrInvalidFile and nothing is
// written.
func (s *Store) Save(name, content string) error {
	if _, err := Parse([]byte(content)); err != nil {
		return err
	}
	if err := s.fs.Save(name, []byte(content)); err != nil {
		return err
	}
	s.forget(name)
	return nil
}

// SaveItems writes items as a bookmark file, replacing any existing
// content (comments in a hand-edited file are not kept).
func (s *Store) SaveItems(name string, items []Item) error {
	if items == nil {
		items = []Item{}
	}
	raw, err := yaml.Marshal(file{Bookmarks: items})
	if err != nil {
		return fmt.Errorf("bookmarks: encode: %w", err)
	}
	return s.Save(name, string(raw))
}

// Read returns the parsed bookmarks of name straight from disk, bypassing
// the cache: for the editor, which must not edit a stale copy.
func (s *Store) Read(name string) ([]Item, error) { return s.load(name) }

// Delete removes a file. Returns data.ErrNotFound if missing.
func (s *Store) Delete(name string) error {
	if err := s.fs.Delete(name); err != nil {
		return err
	}
	s.forget(name)
	return nil
}

// Bookmarks returns the parsed bookmarks of name, read from disk at most
// once per CacheTTL. Only files that exist are cached (a broken one with
// its ErrInvalidFile, so it is not re-parsed on every render): a missing
// file or a bad name is answered from disk each time, so a file created
// on disk shows at once and requests for made-up names cannot grow the
// cache, which the public viewer API would otherwise allow.
func (s *Store) Bookmarks(name string) ([]Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	if c, ok := s.cache[name]; ok && now.Sub(c.loadedAt) < CacheTTL {
		return c.items, c.err
	}
	items, err := s.load(name)
	if err != nil && !errors.Is(err, ErrInvalidFile) {
		delete(s.cache, name)
		return nil, err
	}
	s.pruneExpired(now)
	s.cache[name] = cached{items: items, err: err, loadedAt: now}
	return items, err
}

// pruneExpired drops entries older than CacheTTL, so files deleted on disk
// do not linger. The caller holds s.mu.
func (s *Store) pruneExpired(now time.Time) {
	for name, c := range s.cache {
		if now.Sub(c.loadedAt) >= CacheTTL {
			delete(s.cache, name)
		}
	}
}

func (s *Store) load(name string) ([]Item, error) {
	raw, err := s.fs.Get(name)
	if err != nil {
		return nil, err
	}
	return Parse(raw)
}

func (s *Store) forget(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cache, name)
}

// Parse decodes a bookmark file. Empty content is an empty list; unknown
// keys are rejected so a typo ("subtile") is reported instead of silently
// dropped.
func Parse(raw []byte) ([]Item, error) {
	var f file
	if len(raw) == 0 {
		return []Item{}, nil
	}
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	if err := dec.Decode(&f); err != nil {
		if errors.Is(err, io.EOF) {
			return []Item{}, nil
		}
		return nil, fmt.Errorf("%w: %s", ErrInvalidFile, err.Error())
	}
	if f.Bookmarks == nil {
		f.Bookmarks = []Item{}
	}
	return f.Bookmarks, nil
}
