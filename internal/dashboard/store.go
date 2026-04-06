package dashboard

import (
	"crypto/rand"
	"encoding/json"
	"errors"
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

var (
	ErrNotFound  = errors.New("dashboard not found")
	ErrInvalidID = errors.New("invalid dashboard ID")
)

type Store struct {
	dir   string
	mu    sync.RWMutex
	index map[string]string // id -> folder name
}

func NewStore(dir string) *Store {
	s := &Store{dir: dir, index: make(map[string]string)}
	s.buildIndex()
	return s
}

func (s *Store) ensureDir() error {
	return os.MkdirAll(s.dir, 0o750)
}

const dashboardFile = "dashboard.json"

func (s *Store) dashDir(id string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if folder, ok := s.index[id]; ok {
		return filepath.Join(s.dir, folder)
	}
	// Fallback for legacy folders named by ID
	return filepath.Join(s.dir, id)
}

func (s *Store) filePath(id string) string {
	return filepath.Join(s.dashDir(id), dashboardFile)
}

const idChars = "abcdefghijklmnopqrstuvwxyz0123456789"
const idLen = 6
const previewSuffix = "-prev"

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

func isPreviewID(id string) bool {
	return strings.HasSuffix(id, previewSuffix)
}

// isValidID checks that the ID contains only lowercase alphanumeric chars,
// optionally followed by the preview suffix "-prev".
func isValidID(id string) bool {
	base := strings.TrimSuffix(id, previewSuffix)
	if len(base) == 0 {
		return false
	}
	for _, c := range base {
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}

var allowedAssetExts = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".svg":  true,
	".webp": true,
	".css":  true,
}

func isAllowedAssetExt(name string) bool {
	return allowedAssetExts[strings.ToLower(filepath.Ext(name))]
}

func validateAssetPath(assetPath string) error {
	if assetPath == "" {
		return fmt.Errorf("empty asset path")
	}
	if filepath.IsAbs(assetPath) {
		return fmt.Errorf("absolute paths not allowed")
	}
	cleaned := filepath.Clean(assetPath)
	if strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, string(filepath.Separator)+"..") {
		return fmt.Errorf("path traversal not allowed")
	}
	if filepath.Base(cleaned) == dashboardFile {
		return fmt.Errorf("%s is reserved", dashboardFile)
	}
	if !isAllowedAssetExt(cleaned) {
		return fmt.Errorf("file extension not allowed: %s", filepath.Ext(cleaned))
	}
	return nil
}

// buildIndex scans the store directory and populates the id -> folder mapping.
func (s *Store) buildIndex() {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		fp := filepath.Join(s.dir, entry.Name(), dashboardFile)
		data, err := os.ReadFile(fp)
		if err != nil {
			continue
		}
		var d Dashboard
		if err := json.Unmarshal(data, &d); err != nil {
			continue
		}
		if d.ID != "" {
			s.index[d.ID] = entry.Name()
		}
	}
}

// toSnakeCase converts a dashboard name to a filesystem-safe snake_case string.
var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

func toSnakeCase(name string) string {
	// Normalize unicode and lowercase
	s := strings.ToLower(norm.NFKD.String(name))
	// Strip non-ASCII (accents etc.)
	var buf strings.Builder
	for _, r := range s {
		if r < unicode.MaxASCII {
			buf.WriteRune(r)
		}
	}
	s = buf.String()
	// Replace non-alphanumeric runs with underscore
	s = nonAlphaNum.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	if s == "" {
		s = "dashboard"
	}
	return s
}

// uniqueFolder returns a folder name based on the snake_case dashboard name,
// ensuring it doesn't collide with existing folders in the store directory.
func (s *Store) uniqueFolder(base string) string {
	candidate := base
	i := 2
	for {
		path := filepath.Join(s.dir, candidate)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return candidate
		}
		candidate = fmt.Sprintf("%s_%d", base, i)
		i++
	}
}

func (s *Store) Create(d Dashboard) (Dashboard, error) {
	if err := s.ensureDir(); err != nil {
		return Dashboard{}, fmt.Errorf("create dir: %w", err)
	}

	if d.ID == "" {
		id, err := randomID()
		if err != nil {
			return Dashboard{}, err
		}
		d.ID = id
	}

	s.mu.Lock()
	folder := s.uniqueFolder(toSnakeCase(d.Name))
	dirPath := filepath.Join(s.dir, folder)
	if err := os.MkdirAll(dirPath, 0o750); err != nil {
		s.mu.Unlock()
		return Dashboard{}, fmt.Errorf("create dashboard dir: %w", err)
	}
	s.index[d.ID] = folder
	s.mu.Unlock()

	if err := s.writeToDisk(d); err != nil {
		return Dashboard{}, err
	}
	return d, nil
}

func (s *Store) writeToDisk(d Dashboard) error {
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal dashboard: %w", err)
	}
	if err := os.WriteFile(s.filePath(d.ID), data, 0o600); err != nil {
		return fmt.Errorf("write dashboard file: %w", err)
	}
	return nil
}

// Get reads a single dashboard by ID.
func (s *Store) Get(id string) (Dashboard, error) {
	if !isValidID(id) {
		return Dashboard{}, ErrInvalidID
	}
	data, err := os.ReadFile(s.filePath(id))
	if err != nil {
		return Dashboard{}, fmt.Errorf("read dashboard %s: %w", id, err)
	}
	var d Dashboard
	if err := json.Unmarshal(data, &d); err != nil {
		return Dashboard{}, fmt.Errorf("unmarshal dashboard %s: %w", id, err)
	}
	return d, nil
}

func (s *Store) List() ([]DashboardMeta, error) {
	if err := s.ensureDir(); err != nil {
		return nil, fmt.Errorf("ensure dir: %w", err)
	}

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	result := []DashboardMeta{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		fp := filepath.Join(s.dir, entry.Name(), dashboardFile)
		data, err := os.ReadFile(fp)
		if err != nil {
			continue
		}
		var d Dashboard
		if err := json.Unmarshal(data, &d); err != nil {
			continue
		}
		if isPreviewID(d.ID) {
			continue
		}
		result = append(result, DashboardMeta{
			ID:      d.ID,
			Name:    d.Name,
			Icon:    d.Icon,
			Type:    d.Type,
			Default: d.Default,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})
	return result, nil
}

// Update overwrites an existing dashboard. Returns error if the dashboard does not exist.
func (s *Store) Update(d Dashboard) (Dashboard, error) {
	if !isValidID(d.ID) {
		return Dashboard{}, ErrInvalidID
	}

	s.mu.Lock()
	currentFolder := s.index[d.ID]
	var oldDir string
	if currentFolder != "" {
		oldDir = filepath.Join(s.dir, currentFolder)
	} else {
		oldDir = filepath.Join(s.dir, d.ID)
	}

	if _, err := os.Stat(oldDir); err != nil {
		s.mu.Unlock()
		if os.IsNotExist(err) {
			return Dashboard{}, ErrNotFound
		}
		return Dashboard{}, fmt.Errorf("stat dashboard %s: %w", d.ID, err)
	}

	// Rename folder if the dashboard name changed
	newFolder := toSnakeCase(d.Name)
	if currentFolder != "" && newFolder != currentFolder {
		newPath := filepath.Join(s.dir, newFolder)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			if err := os.Rename(oldDir, newPath); err == nil {
				s.index[d.ID] = newFolder
			}
		}
	}
	s.mu.Unlock()

	if err := s.writeToDisk(d); err != nil {
		return Dashboard{}, err
	}
	return d, nil
}

// Delete removes a dashboard directory. Returns error if the dashboard does not exist.
func (s *Store) Delete(id string) error {
	if !isValidID(id) {
		return ErrInvalidID
	}
	dir := s.dashDir(id)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return fmt.Errorf("stat dashboard %s: %w", id, err)
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("delete dashboard %s: %w", id, err)
	}

	s.mu.Lock()
	delete(s.index, id)
	s.mu.Unlock()

	return nil
}

// DeletePreviews removes all preview dashboard directories and returns the count deleted.
func (s *Store) DeletePreviews() (int, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read dir: %w", err)
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Read the dashboard.json to check if it's a preview
		fp := filepath.Join(s.dir, entry.Name(), dashboardFile)
		data, err := os.ReadFile(fp)
		if err != nil {
			continue
		}
		var d Dashboard
		if err := json.Unmarshal(data, &d); err != nil {
			continue
		}
		if !isPreviewID(d.ID) {
			continue
		}
		if err := os.RemoveAll(filepath.Join(s.dir, entry.Name())); err != nil {
			return count, fmt.Errorf("delete preview %s: %w", entry.Name(), err)
		}
		s.mu.Lock()
		delete(s.index, d.ID)
		s.mu.Unlock()
		count++
	}
	return count, nil
}

const customCSSFile = "custom.css"

// GetCustomCSS reads the custom.css sidecar file for a dashboard.
// Returns an empty string if the file does not exist.
func (s *Store) GetCustomCSS(id string) string {
	data, err := os.ReadFile(filepath.Join(s.dashDir(id), customCSSFile))
	if err != nil {
		return ""
	}
	return string(data)
}

func (s *Store) SaveAsset(id string, assetPath string, data []byte) error {
	if !isValidID(id) {
		return fmt.Errorf("invalid dashboard ID: %s", id)
	}
	if err := validateAssetPath(assetPath); err != nil {
		return fmt.Errorf("invalid asset path: %w", err)
	}

	dashDir := s.dashDir(id)
	if _, err := os.Stat(dashDir); os.IsNotExist(err) {
		return fmt.Errorf("dashboard %s not found", id)
	}

	fullPath := filepath.Join(dashDir, filepath.Clean(assetPath))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o750); err != nil {
		return fmt.Errorf("create asset directory: %w", err)
	}
	if err := os.WriteFile(fullPath, data, 0o600); err != nil {
		return fmt.Errorf("write asset: %w", err)
	}
	return nil
}

func (s *Store) GetAsset(id string, assetPath string) ([]byte, string, error) {
	if !isValidID(id) {
		return nil, "", fmt.Errorf("invalid dashboard ID: %s", id)
	}
	if err := validateAssetPath(assetPath); err != nil {
		return nil, "", fmt.Errorf("invalid asset path: %w", err)
	}

	fullPath := filepath.Join(s.dashDir(id), filepath.Clean(assetPath))
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, "", fmt.Errorf("read asset: %w", err)
	}

	mimeType := mime.TypeByExtension(filepath.Ext(assetPath))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	return data, mimeType, nil
}

func (s *Store) DeleteAsset(id string, assetPath string) error {
	if !isValidID(id) {
		return fmt.Errorf("invalid dashboard ID: %s", id)
	}
	if err := validateAssetPath(assetPath); err != nil {
		return fmt.Errorf("invalid asset path: %w", err)
	}

	dashDir := s.dashDir(id)
	fullPath := filepath.Join(dashDir, filepath.Clean(assetPath))
	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("delete asset: %w", err)
	}

	// Clean up empty parent directories up to the dashboard root
	dir := filepath.Dir(fullPath)
	for dir != dashDir {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			break
		}
		_ = os.Remove(dir)
		dir = filepath.Dir(dir)
	}
	return nil
}

// DashDir returns the filesystem path of the dashboard directory for the given ID.
func (s *Store) DashDir(id string) (string, error) {
	if !isValidID(id) {
		return "", ErrInvalidID
	}
	dir := s.dashDir(id)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", ErrNotFound
	}
	return dir, nil
}

func (s *Store) ListAssets(id string) ([]string, error) {
	if !isValidID(id) {
		return nil, fmt.Errorf("invalid dashboard ID: %s", id)
	}

	dashDir := s.dashDir(id)
	if _, err := os.Stat(dashDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("dashboard %s not found", id)
	}

	var assets []string
	err := filepath.WalkDir(dashDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dashDir, path)
		if err != nil {
			return err
		}
		if rel == dashboardFile {
			return nil
		}
		assets = append(assets, rel)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk dashboard dir: %w", err)
	}
	if assets == nil {
		assets = []string{}
	}
	return assets, nil
}
