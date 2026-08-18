// Package data provides the shared, dashboard-independent user-data
// layer. Content (notes, images, backgrounds) is stored under
// {dataDir}/data/{kind}/ as flat collections that any dashboard can
// reference by name.
//
// This package exports the shared types and errors used by the
// per-kind sub-packages (data/notes, data/images, data/backgrounds).
// The filesystem primitive FsStore is exported so the per-kind
// sub-packages can wrap it with typed APIs.
package data

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Item describes one stored entry as returned by Store.List.
type Item struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
}

// Sentinel errors. Kind stores return these directly (no wrapping)
// so handlers can map them to HTTP status codes via errors.Is.
var (
	ErrNotFound      = errors.New("data: not found")
	ErrInvalidName   = errors.New("data: invalid name")
	ErrExtNotAllowed = errors.New("data: file extension not allowed")
)

const maxNameLen = 255

// validateName enforces the flat-namespace, safe-name rules documented
// in the design spec. It accepts only bare filenames (no path
// components), rejects control chars and traversal, and requires an
// extension present in allowedExts (keys lowercased including the dot).
func validateName(name string, allowedExts map[string]bool) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("%w: empty", ErrInvalidName)
	}
	if len(trimmed) > maxNameLen {
		return fmt.Errorf("%w: name exceeds %d bytes", ErrInvalidName, maxNameLen)
	}
	// Normalize to NFC so visually identical unicode doesn't create
	// duplicate entries on disk.
	trimmed = norm.NFC.String(trimmed)

	if strings.ContainsAny(trimmed, `/\`) {
		return fmt.Errorf("%w: path separator not allowed", ErrInvalidName)
	}
	if strings.HasPrefix(trimmed, ".") {
		return fmt.Errorf("%w: leading dot not allowed", ErrInvalidName)
	}
	if strings.Contains(trimmed, "..") {
		return fmt.Errorf("%w: traversal not allowed", ErrInvalidName)
	}
	for _, r := range trimmed {
		if r == 0 || (unicode.IsControl(r)) {
			return fmt.Errorf("%w: control character not allowed", ErrInvalidName)
		}
	}

	ext := strings.ToLower(filepath.Ext(trimmed))
	if !allowedExts[ext] {
		return fmt.Errorf("%w: %q", ErrExtNotAllowed, ext)
	}
	return nil
}

// FsStore is the exported filesystem primitive shared by the per-kind
// sub-packages. All reads/writes go through os.Root to prevent
// symlink traversal (TOCTOU-safe).
type FsStore struct {
	dir         string
	allowedExts map[string]bool
	mu          sync.RWMutex
}

// NewFsStore creates dir if missing and returns a new store.
func NewFsStore(dir string, allowedExts map[string]bool) (*FsStore, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	return &FsStore{dir: dir, allowedExts: allowedExts}, nil
}

func (s *FsStore) openRoot() (*os.Root, error) {
	return os.OpenRoot(s.dir)
}

// Save writes data to name atomically (name.tmp + rename). Validates
// name against the allowed extension list. Overwrites an existing
// file of the same name.
func (s *FsStore) Save(name string, data []byte) error {
	if err := validateName(name, s.allowedExts); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	root, err := s.openRoot()
	if err != nil {
		return fmt.Errorf("open root: %w", err)
	}
	defer func() { _ = root.Close() }()

	tmp := name + ".tmp"
	f, err := root.Create(tmp)
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = root.Remove(tmp)
		return fmt.Errorf("write temp: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = root.Remove(tmp)
		return fmt.Errorf("close temp: %w", err)
	}

	if err := root.Rename(tmp, name); err != nil {
		_ = root.Remove(tmp)
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

// Get reads the full content of name. Returns ErrNotFound if the file
// does not exist.
func (s *FsStore) Get(name string) ([]byte, error) {
	if err := validateName(name, s.allowedExts); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	root, err := s.openRoot()
	if err != nil {
		return nil, fmt.Errorf("open root: %w", err)
	}
	defer func() { _ = root.Close() }()

	f, err := root.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("open: %w", err)
	}
	defer func() { _ = f.Close() }()
	return io.ReadAll(f)
}

// Delete removes name. Returns ErrNotFound if it does not exist.
func (s *FsStore) Delete(name string) error {
	if err := validateName(name, s.allowedExts); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	root, err := s.openRoot()
	if err != nil {
		return fmt.Errorf("open root: %w", err)
	}
	defer func() { _ = root.Close() }()

	if err := root.Remove(name); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return fmt.Errorf("remove: %w", err)
	}
	return nil
}

// List returns one Item per regular file directly in the store
// directory. Subdirectories and .tmp files are ignored. Files with
// disallowed extensions are ignored (defensive — should not exist).
func (s *FsStore) List() ([]Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Item{}, nil
		}
		return nil, fmt.Errorf("read dir: %w", err)
	}

	out := make([]Item, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".tmp") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(name))
		if !s.allowedExts[ext] {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, Item{
			Name:    name,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}
	// Stable alphabetical order makes tests and UIs predictable.
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
