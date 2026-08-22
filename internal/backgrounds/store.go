package backgrounds

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/fs"
	"math/big"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"unicode"

	"github.com/andresbott/dashi/internal/data/images"
	"golang.org/x/text/unicode/norm"
)

const backgroundFile = "background.json"
const assetsDir = "assets"

// Store persists backgrounds as one folder per entity. The folder name is
// snake_case of the display name and an in-memory id -> folder index is
// built at startup, so renaming a background never moves it out from under
// the dashboards referencing it by id.
//
// Build exactly one Store per process (in newSharedDeps): a second instance
// would hold a stale index.
type Store struct {
	dir   string
	pool  *images.Store
	mu    sync.RWMutex
	index map[string]string // id -> folder
}

func NewStore(dir string, pool *images.Store) *Store {
	s := &Store{dir: dir, pool: pool, index: make(map[string]string)}
	s.buildIndex()
	return s
}

func (s *Store) buildIndex() {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(s.dir, e.Name(), backgroundFile))
		if err != nil {
			continue
		}
		var b Background
		if err := json.Unmarshal(raw, &b); err != nil {
			continue
		}
		if b.ID != "" {
			s.index[b.ID] = e.Name()
		}
	}
}

func (s *Store) bgDir(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	folder, ok := s.index[id]
	if !ok {
		return "", false
	}
	return filepath.Join(s.dir, folder), true
}

const idChars = "abcdefghijklmnopqrstuvwxyz0123456789"
const idLen = 6

func randomID() (string, error) {
	b := make([]byte, idLen)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(idChars))))
		if err != nil {
			return "", fmt.Errorf("generate random ID: %w", err)
		}
		b[i] = idChars[n.Int64()]
	}
	return string(b), nil
}

func isValidID(id string) bool {
	if id == "" {
		return false
	}
	for _, c := range id {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

func toSnakeCase(name string) string {
	s := strings.ToLower(norm.NFKD.String(name))
	var buf strings.Builder
	for _, r := range s {
		if r < unicode.MaxASCII {
			buf.WriteRune(r)
		}
	}
	s = nonAlphaNum.ReplaceAllString(buf.String(), "_")
	s = strings.Trim(s, "_")
	if s == "" {
		s = "background"
	}
	return s
}

func (s *Store) uniqueFolder(base string) string {
	candidate := base
	for i := 2; ; i++ {
		if _, err := os.Stat(filepath.Join(s.dir, candidate)); os.IsNotExist(err) {
			return candidate
		}
		candidate = fmt.Sprintf("%s_%d", base, i)
	}
}

func (s *Store) Create(b Background) (Background, error) {
	if err := b.Validate(); err != nil {
		return Background{}, err
	}
	if err := os.MkdirAll(s.dir, 0o750); err != nil {
		return Background{}, fmt.Errorf("create dir: %w", err)
	}
	if b.ID == "" {
		id, err := randomID()
		if err != nil {
			return Background{}, err
		}
		b.ID = id
	}

	s.mu.Lock()
	folder := s.uniqueFolder(toSnakeCase(b.Name))
	dirPath := filepath.Join(s.dir, folder)
	if err := os.MkdirAll(dirPath, 0o750); err != nil {
		s.mu.Unlock()
		return Background{}, fmt.Errorf("create background dir: %w", err)
	}
	s.index[b.ID] = folder
	s.mu.Unlock()

	if err := s.write(b); err != nil {
		s.rollbackCreate(b.ID, dirPath)
		return Background{}, err
	}
	return b, nil
}

// rollbackCreate undoes a partially completed Create: it removes the folder
// that was made for the background and drops its index entry, so a failed
// write cannot leave an id resolving to a directory with no background.json.
func (s *Store) rollbackCreate(id, dirPath string) {
	_ = os.RemoveAll(dirPath)
	s.mu.Lock()
	delete(s.index, id)
	s.mu.Unlock()
}

func (s *Store) write(b Background) error {
	dir, ok := s.bgDir(b.ID)
	if !ok {
		return ErrNotFound
	}
	raw, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal background: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, backgroundFile), raw, 0o600); err != nil {
		return fmt.Errorf("write background file: %w", err)
	}
	return nil
}

func (s *Store) Get(id string) (Background, error) {
	if !isValidID(id) {
		return Background{}, ErrInvalidID
	}
	dir, ok := s.bgDir(id)
	if !ok {
		return Background{}, ErrNotFound
	}
	raw, err := os.ReadFile(filepath.Join(dir, backgroundFile))
	if err != nil {
		if os.IsNotExist(err) {
			return Background{}, ErrNotFound
		}
		return Background{}, fmt.Errorf("read background %s: %w", id, err)
	}
	var b Background
	if err := json.Unmarshal(raw, &b); err != nil {
		return Background{}, fmt.Errorf("unmarshal background %s: %w", id, err)
	}
	return b, nil
}

// List returns one Meta per background, ordered case-insensitively by name.
// usage maps a background ID to the number of dashboards referencing it; a
// nil map yields zero counts.
func (s *Store) List(usage map[string]int) ([]Meta, error) {
	if err := os.MkdirAll(s.dir, 0o750); err != nil {
		return nil, fmt.Errorf("ensure dir: %w", err)
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}
	out := []Meta{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(s.dir, e.Name(), backgroundFile))
		if err != nil {
			continue
		}
		var b Background
		if err := json.Unmarshal(raw, &b); err != nil {
			continue
		}
		out = append(out, Meta{
			ID:         b.ID,
			Name:       b.Name,
			UsedBy:     usage[b.ID],
			PreviewCSS: string(BrowserCSS(b)),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

func (s *Store) Update(b Background) (Background, error) {
	if !isValidID(b.ID) {
		return Background{}, ErrInvalidID
	}
	if err := b.Validate(); err != nil {
		return Background{}, err
	}

	s.mu.Lock()
	folder, ok := s.index[b.ID]
	if !ok {
		s.mu.Unlock()
		return Background{}, ErrNotFound
	}
	oldDir := filepath.Join(s.dir, folder)
	if _, err := os.Stat(oldDir); err != nil {
		s.mu.Unlock()
		if os.IsNotExist(err) {
			return Background{}, ErrNotFound
		}
		return Background{}, fmt.Errorf("stat background %s: %w", b.ID, err)
	}
	if newFolder := toSnakeCase(b.Name); newFolder != folder {
		newPath := filepath.Join(s.dir, newFolder)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			if err := os.Rename(oldDir, newPath); err == nil {
				s.index[b.ID] = newFolder
			}
		}
	}
	s.mu.Unlock()

	if err := s.write(b); err != nil {
		return Background{}, err
	}
	return b, nil
}

func (s *Store) Delete(id string) error {
	if !isValidID(id) {
		return ErrInvalidID
	}
	dir, ok := s.bgDir(id)
	if !ok {
		return ErrNotFound
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("delete background %s: %w", id, err)
	}
	s.mu.Lock()
	delete(s.index, id)
	s.mu.Unlock()
	return nil
}

// validateAssetName accepts a path relative to the background's assets/
// folder. Validation lives here rather than in the handler, matching
// dashboard.Store.
func validateAssetName(name string) error {
	if err := ValidateRef("asset:" + name); err != nil {
		return err
	}
	if filepath.Base(filepath.Clean(name)) == backgroundFile {
		return fmt.Errorf("%w: %s is reserved", ErrInvalidConfig, backgroundFile)
	}
	return nil
}

func (s *Store) assetPath(id, name string) (string, error) {
	if !isValidID(id) {
		return "", ErrInvalidID
	}
	if err := validateAssetName(name); err != nil {
		return "", err
	}
	dir, ok := s.bgDir(id)
	if !ok {
		return "", ErrNotFound
	}
	return filepath.Join(dir, assetsDir, filepath.Clean(name)), nil
}

func (s *Store) SaveAsset(id, name string, data []byte) error {
	full, err := s.assetPath(id, name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o750); err != nil {
		return fmt.Errorf("create asset directory: %w", err)
	}
	if err := os.WriteFile(full, data, 0o600); err != nil {
		return fmt.Errorf("write asset: %w", err)
	}
	return nil
}

func (s *Store) GetAsset(id, name string) ([]byte, string, error) {
	full, err := s.assetPath(id, name)
	if err != nil {
		return nil, "", err
	}
	data, err := os.ReadFile(full)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", ErrNotFound
		}
		return nil, "", fmt.Errorf("read asset: %w", err)
	}
	mimeType := mime.TypeByExtension(filepath.Ext(name))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return data, mimeType, nil
}

func (s *Store) DeleteAsset(id, name string) error {
	full, err := s.assetPath(id, name)
	if err != nil {
		return err
	}
	if err := os.Remove(full); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return fmt.Errorf("delete asset: %w", err)
	}
	return nil
}

// ListAssets returns paths relative to the background's assets/ folder.
func (s *Store) ListAssets(id string) ([]string, error) {
	if !isValidID(id) {
		return nil, ErrInvalidID
	}
	dir, ok := s.bgDir(id)
	if !ok {
		return nil, ErrNotFound
	}
	root := filepath.Join(dir, assetsDir)
	assets := []string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil // no assets/ folder yet
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		assets = append(assets, rel)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk assets: %w", err)
	}
	sort.Strings(assets)
	return assets, nil
}

// LoadImage returns the bytes an image reference points at. The litehtml
// stack needs raw bytes because it fetches nothing over the network.
func (s *Store) LoadImage(id, ref string) ([]byte, error) {
	scheme, path, ok := strings.Cut(ref, ":")
	if !ok {
		return nil, fmt.Errorf("%w: reference %q has no scheme", ErrInvalidConfig, ref)
	}
	switch scheme {
	case "asset":
		data, _, err := s.GetAsset(id, path)
		return data, err
	case "shared":
		if s.pool == nil {
			return nil, fmt.Errorf("shared image pool not configured")
		}
		data, _, err := s.pool.Get(path)
		return data, err
	}
	return nil, fmt.Errorf("%w: unsupported reference scheme %q", ErrInvalidConfig, scheme)
}
