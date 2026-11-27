package themes

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

	"golang.org/x/text/unicode/norm"
)

const themeFile = "theme.json"
const assetsDir = "assets"

// Store persists themes as one folder per entity. The folder name is
// snake_case of the display name and an in-memory id -> folder index is
// built at startup, so renaming a theme never moves it out from under the
// dashboards referencing it by id.
//
// The builtin themes (see builtins) are never written to disk: Get returns
// them from code and every write on one fails with ErrBuiltin.
//
// Build exactly one Store per process (in newSharedDeps): a second instance
// would hold a stale index.
type Store struct {
	dir   string
	mu    sync.RWMutex
	index map[string]string // id -> folder
}

func NewStore(dir string) *Store {
	s := &Store{dir: dir, index: make(map[string]string)}
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
		t, ok := readThemeFile(filepath.Join(s.dir, e.Name()))
		if !ok {
			continue
		}
		s.index[t.ID] = e.Name()
	}
}

// readThemeFile reads dir/theme.json. It reports false for a missing or
// unparsable file, an empty ID, or a stored file claiming a builtin ID,
// which must never shadow the builtin.
func readThemeFile(dir string) (Theme, bool) {
	raw, err := os.ReadFile(filepath.Join(dir, themeFile))
	if err != nil {
		return Theme{}, false
	}
	t, err := decodeTheme(raw)
	if err != nil {
		return Theme{}, false
	}
	if t.ID == "" || IsBuiltin(t.ID) {
		return Theme{}, false
	}
	return t, true
}

func (s *Store) themeDir(id string) (string, bool) {
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
		s = "theme"
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

// Create stores t. An empty ID is assigned a random one; a builtin ID is
// refused.
func (s *Store) Create(t Theme) (Theme, error) {
	if IsBuiltin(t.ID) {
		return Theme{}, ErrBuiltin
	}
	if err := t.Validate(); err != nil {
		return Theme{}, err
	}
	if err := os.MkdirAll(s.dir, 0o750); err != nil {
		return Theme{}, fmt.Errorf("create dir: %w", err)
	}
	if t.ID == "" {
		id, err := randomID()
		if err != nil {
			return Theme{}, err
		}
		t.ID = id
	}

	s.mu.Lock()
	folder := s.uniqueFolder(toSnakeCase(t.Name))
	dirPath := filepath.Join(s.dir, folder)
	if err := os.MkdirAll(dirPath, 0o750); err != nil {
		s.mu.Unlock()
		return Theme{}, fmt.Errorf("create theme dir: %w", err)
	}
	s.index[t.ID] = folder
	s.mu.Unlock()

	if err := s.write(t); err != nil {
		s.rollbackCreate(t.ID, dirPath)
		return Theme{}, err
	}
	return t, nil
}

// rollbackCreate undoes a partially completed Create: it removes the folder
// that was made for the theme and drops its index entry, so a failed write
// cannot leave an id resolving to a directory with no theme.json.
func (s *Store) rollbackCreate(id, dirPath string) {
	_ = os.RemoveAll(dirPath)
	s.mu.Lock()
	delete(s.index, id)
	s.mu.Unlock()
}

func (s *Store) write(t Theme) error {
	dir, ok := s.themeDir(t.ID)
	if !ok {
		return ErrNotFound
	}
	raw, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal theme: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, themeFile), raw, 0o600); err != nil {
		return fmt.Errorf("write theme file: %w", err)
	}
	return nil
}

// Get returns the theme with the given ID; a builtin ID returns the builtin.
func (s *Store) Get(id string) (Theme, error) {
	if t, ok := Builtin(id); ok {
		return t, nil
	}
	if !isValidID(id) {
		return Theme{}, ErrInvalidID
	}
	dir, ok := s.themeDir(id)
	if !ok {
		return Theme{}, ErrNotFound
	}
	raw, err := os.ReadFile(filepath.Join(dir, themeFile))
	if err != nil {
		if os.IsNotExist(err) {
			return Theme{}, ErrNotFound
		}
		return Theme{}, fmt.Errorf("read theme %s: %w", id, err)
	}
	t, err := decodeTheme(raw)
	if err != nil {
		return Theme{}, fmt.Errorf("unmarshal theme %s: %w", id, err)
	}
	return t, nil
}

// Resolve returns the theme a dashboard should render with: the stored
// theme, or DefaultTheme() for an empty, unknown, dangling or unreadable
// reference. It never fails, so a bad reference can never break a render.
func (s *Store) Resolve(id string) Theme {
	t, err := s.Get(id)
	if err != nil {
		return DefaultTheme()
	}
	return t
}

// List returns one Meta per theme, the builtins included and flagged, ordered
// case-insensitively by name. usage maps a theme ID to the number of
// dashboards referencing it; a nil map yields zero counts.
func (s *Store) List(usage map[string]int) ([]Meta, error) {
	if err := os.MkdirAll(s.dir, 0o750); err != nil {
		return nil, fmt.Errorf("ensure dir: %w", err)
	}
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}
	out := make([]Meta, 0, len(builtins)+len(entries))
	for _, fn := range builtins {
		out = append(out, newMeta(fn(), true, usage))
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		t, ok := readThemeFile(filepath.Join(s.dir, e.Name()))
		if !ok {
			continue
		}
		out = append(out, newMeta(t, false, usage))
	}
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}

func newMeta(t Theme, builtin bool, usage map[string]int) Meta {
	return Meta{
		ID:         t.ID,
		Name:       t.Name,
		Builtin:    builtin,
		DarkMode:   t.DarkMode,
		UsedBy:     usage[t.ID],
		PreviewCSS: string(BrowserCSS(t)),
		Colors:     Colors{Light: t.Palette("light"), Dark: t.Palette("dark")},
	}
}

func (s *Store) Update(t Theme) (Theme, error) {
	if IsBuiltin(t.ID) {
		return Theme{}, ErrBuiltin
	}
	if !isValidID(t.ID) {
		return Theme{}, ErrInvalidID
	}
	if err := t.Validate(); err != nil {
		return Theme{}, err
	}

	s.mu.Lock()
	folder, ok := s.index[t.ID]
	if !ok {
		s.mu.Unlock()
		return Theme{}, ErrNotFound
	}
	oldDir := filepath.Join(s.dir, folder)
	if _, err := os.Stat(oldDir); err != nil {
		s.mu.Unlock()
		if os.IsNotExist(err) {
			return Theme{}, ErrNotFound
		}
		return Theme{}, fmt.Errorf("stat theme %s: %w", t.ID, err)
	}
	if newFolder := toSnakeCase(t.Name); newFolder != folder {
		newPath := filepath.Join(s.dir, newFolder)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			if err := os.Rename(oldDir, newPath); err == nil {
				s.index[t.ID] = newFolder
			}
		}
	}
	s.mu.Unlock()

	if err := s.write(t); err != nil {
		return Theme{}, err
	}
	return t, nil
}

func (s *Store) Delete(id string) error {
	if IsBuiltin(id) {
		return ErrBuiltin
	}
	if !isValidID(id) {
		return ErrInvalidID
	}
	dir, ok := s.themeDir(id)
	if !ok {
		return ErrNotFound
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("delete theme %s: %w", id, err)
	}
	s.mu.Lock()
	delete(s.index, id)
	s.mu.Unlock()
	return nil
}

// Copy duplicates a theme, assets included, into a new theme with a new ID
// and the name "Copy of <name>". The copy owns its own asset files, so it
// survives the original being deleted. Copying a builtin yields a regular,
// editable theme with the builtin's palettes and no assets.
func (s *Store) Copy(id string) (Theme, error) {
	src, err := s.Get(id)
	if err != nil {
		return Theme{}, err
	}
	cp := src
	cp.ID = ""
	cp.Name = "Copy of " + src.Name
	created, err := s.Create(cp)
	if err != nil {
		return Theme{}, err
	}
	if IsBuiltin(id) {
		return created, nil
	}
	if err := s.copyAssets(id, created.ID); err != nil {
		_ = s.Delete(created.ID)
		return Theme{}, err
	}
	return created, nil
}

// copyAssets copies every regular file under the source theme's assets/
// into the destination's. Anything else (symlinks, devices) is skipped so a
// hand-placed link cannot pull files from outside the theme into the copy.
func (s *Store) copyAssets(srcID, dstID string) error {
	srcDir, ok := s.themeDir(srcID)
	if !ok {
		return ErrNotFound
	}
	dstDir, ok := s.themeDir(dstID)
	if !ok {
		return ErrNotFound
	}
	srcRoot := filepath.Join(srcDir, assetsDir)
	dstRoot := filepath.Join(dstDir, assetsDir)
	err := filepath.WalkDir(srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) && path == srcRoot {
				return nil // no assets/ folder yet
			}
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path) //nolint:gosec // G304: path comes from walking this store's own assets folder
		if err != nil {
			return err
		}
		target := filepath.Join(dstRoot, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o600) //nolint:gosec // G703: target is the destination theme's assets dir joined with a path relative to the walked source assets dir
	})
	if err != nil {
		return fmt.Errorf("copy assets: %w", err)
	}
	return nil
}

// validateAssetName accepts a path relative to the theme's assets/ folder.
// Validation lives here rather than in the handler, matching
// dashboard.Store.
func validateAssetName(name string) error {
	if err := ValidateRef("asset:" + name); err != nil {
		return err
	}
	if filepath.Base(filepath.Clean(name)) == themeFile {
		return fmt.Errorf("%w: %s is reserved", ErrInvalidConfig, themeFile)
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
	dir, ok := s.themeDir(id)
	if !ok {
		return "", ErrNotFound
	}
	return filepath.Join(dir, assetsDir, filepath.Clean(name)), nil
}

func (s *Store) SaveAsset(id, name string, data []byte) error {
	if IsBuiltin(id) {
		return ErrBuiltin
	}
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
	if IsBuiltin(id) {
		return ErrBuiltin
	}
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

// ListAssets returns paths relative to the theme's assets/ folder. The
// builtins have none.
func (s *Store) ListAssets(id string) ([]string, error) {
	if IsBuiltin(id) {
		return []string{}, nil
	}
	if !isValidID(id) {
		return nil, ErrInvalidID
	}
	dir, ok := s.themeDir(id)
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
	if scheme != "asset" {
		return nil, fmt.Errorf("%w: unsupported reference scheme %q", ErrInvalidConfig, scheme)
	}
	data, _, err := s.GetAsset(id, path)
	return data, err
}
