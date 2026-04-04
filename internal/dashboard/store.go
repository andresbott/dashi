package dashboard

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/fs"
	"math/big"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

type Store struct {
	dir string
}

func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

func (s *Store) ensureDir() error {
	return os.MkdirAll(s.dir, 0o755)
}

const dashboardFile = "dashboard.json"

func (s *Store) dashDir(id string) string {
	return filepath.Join(s.dir, id)
}

func (s *Store) filePath(id string) string {
	return filepath.Join(s.dir, id, dashboardFile)
}

const idChars = "abcdefghijklmnopqrstuvwxyz0123456789"
const idLen = 6
const previewSuffix = "-prev"

func randomID() string {
	b := make([]byte, idLen)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(idChars))))
		b[i] = idChars[n.Int64()]
	}
	return string(b)
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
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
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

func (s *Store) Create(d Dashboard) (Dashboard, error) {
	if err := s.ensureDir(); err != nil {
		return Dashboard{}, fmt.Errorf("create dir: %w", err)
	}

	if d.ID == "" {
		d.ID = randomID()
	}

	if err := os.MkdirAll(s.dashDir(d.ID), 0o755); err != nil {
		return Dashboard{}, fmt.Errorf("create dashboard dir: %w", err)
	}

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
	if err := os.WriteFile(s.filePath(d.ID), data, 0o644); err != nil {
		return fmt.Errorf("write dashboard file: %w", err)
	}
	return nil
}

// Get reads a single dashboard by ID.
func (s *Store) Get(id string) (Dashboard, error) {
	if !isValidID(id) {
		return Dashboard{}, fmt.Errorf("invalid dashboard ID: %s", id)
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
		id := entry.Name()
		if isPreviewID(id) {
			continue
		}
		d, err := s.Get(id)
		if err != nil {
			continue
		}
		result = append(result, DashboardMeta{
			ID:   d.ID,
			Name: d.Name,
			Icon: d.Icon,
			Type: d.Type,
		})
	}
	return result, nil
}

// Update overwrites an existing dashboard. Returns error if the dashboard does not exist.
func (s *Store) Update(d Dashboard) (Dashboard, error) {
	if !isValidID(d.ID) {
		return Dashboard{}, fmt.Errorf("invalid dashboard ID: %s", d.ID)
	}
	if _, err := os.Stat(s.dashDir(d.ID)); err != nil {
		if os.IsNotExist(err) {
			return Dashboard{}, fmt.Errorf("dashboard %s not found", d.ID)
		}
		return Dashboard{}, fmt.Errorf("stat dashboard %s: %w", d.ID, err)
	}
	if err := s.writeToDisk(d); err != nil {
		return Dashboard{}, err
	}
	return d, nil
}

// Delete removes a dashboard directory. Returns error if the dashboard does not exist.
func (s *Store) Delete(id string) error {
	if !isValidID(id) {
		return fmt.Errorf("invalid dashboard ID: %s", id)
	}
	dir := s.dashDir(id)
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("dashboard %s not found", id)
		}
		return fmt.Errorf("stat dashboard %s: %w", id, err)
	}
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("delete dashboard %s: %w", id, err)
	}
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
		if !isPreviewID(entry.Name()) {
			continue
		}
		if err := os.RemoveAll(filepath.Join(s.dir, entry.Name())); err != nil {
			return count, fmt.Errorf("delete preview %s: %w", entry.Name(), err)
		}
		count++
	}
	return count, nil
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
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return fmt.Errorf("create asset directory: %w", err)
	}
	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
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
		os.Remove(dir)
		dir = filepath.Dir(dir)
	}
	return nil
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
