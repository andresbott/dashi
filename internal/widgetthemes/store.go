package widgetthemes

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode"

	"gopkg.in/yaml.v3"
)

// MaxUploadSize caps uploaded theme zip bodies. 20 MB is generous for
// icon/style themes (a few fonts + an icons manifest); upload handlers
// should enforce the same cap via http.MaxBytesReader.
const MaxUploadSize = 20 << 20 // 20 MB

// Sentinel errors. Handlers map these to HTTP status codes with
// errors.Is — do not wrap or rename without also updating the handler.
var (
	ErrNotFound       = errors.New("themes: not found")
	ErrConflict       = errors.New("themes: already exists")
	ErrBuiltin        = errors.New("themes: builtin theme cannot be modified")
	ErrInvalidName    = errors.New("themes: invalid name")
	ErrInvalidArchive = errors.New("themes: invalid archive")
)

//go:embed defaults/widget_theme.yaml
var defaultThemeYAML []byte

//go:embed defaults/tabler-icons.ttf
var defaultFontTTF []byte

//go:embed defaults/Inter-Regular.ttf
var interRegularTTF []byte

var embeddedFonts = map[string][]byte{
	"tabler-icons.ttf":  defaultFontTTF,
	"Inter-Regular.ttf": interRegularTTF,
}

type Store struct {
	mu       sync.RWMutex
	themes   map[string]*theme
	builtins map[string]bool // names of embedded themes that cannot be deleted
	rootDir  string          // directory where user themes live on disk (empty = disk writes disabled)
	fallback *theme          // the embedded default; what effective falls back to
}

func NewStore(themesDir string) *Store {
	s := &Store{
		themes:   make(map[string]*theme),
		builtins: make(map[string]bool),
		rootDir:  themesDir,
	}
	s.loadEmbeddedDefault()
	if themesDir != "" {
		s.loadUserThemes(themesDir)
	}
	return s
}

func (s *Store) loadEmbeddedDefault() {
	var m themeManifest
	if err := yaml.Unmarshal(defaultThemeYAML, &m); err != nil {
		panic(fmt.Sprintf("embedded default %s is invalid: %v", ManifestFile, err))
	}
	s.fallback = &theme{manifest: m}
	s.themes[m.Name] = s.fallback
	s.builtins[m.Name] = true
}

func (s *Store) loadUserThemes(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		themeDir := filepath.Join(dir, entry.Name())
		manifestPath := filepath.Join(themeDir, ManifestFile)
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}
		var m themeManifest
		if err := yaml.Unmarshal(data, &m); err != nil {
			continue
		}
		s.themes[entry.Name()] = &theme{manifest: m, dir: themeDir}
	}
}

func (s *Store) themeInfoLocked(name string, th *theme) Info {
	fonts := make([]FontInfo, len(th.manifest.Fonts))
	for i, f := range th.manifest.Fonts {
		fonts[i] = FontInfo{Name: f.Name}
	}
	iconType := ""
	if ic := th.icons(); ic != nil {
		iconType = ic.Type
	}
	return Info{
		Name:        name,
		Description: th.manifest.Description,
		Fonts:       fonts,
		HasIcons:    th.hasIcons(),
		IconType:    iconType,
		Builtin:     s.builtins[name],
	}
}

func (s *Store) List() []Info {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Info, 0, len(s.themes))
	for name, th := range s.themes {
		result = append(result, s.themeInfoLocked(name, th))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func (s *Store) Get(name string) (Info, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	th, ok := s.themes[name]
	if !ok {
		return Info{}, false
	}
	return s.themeInfoLocked(name, th), true
}

// lookupTheme returns a reference to the in-memory theme under an
// RLock. The returned *theme is safe to use read-only afterwards — its
// fields are never mutated once loaded; Upload/Delete swap the map
// entry rather than mutating in place.
func (s *Store) lookupTheme(name string) (*theme, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	th, ok := s.themes[name]
	return th, ok
}

// effective returns the name and theme to use for a request about `name`: the
// theme itself when it exists and satisfies need, otherwise the builtin
// default. This is the one place the "unknown name or missing part falls back
// to default" rule lives. The default is always the embedded one, so a
// disk override of "default" that lacks a part cannot leave the fallback
// itself incomplete.
func (s *Store) effective(name string, need func(*theme) bool) (string, *theme) {
	if th, ok := s.lookupTheme(name); ok && need(th) {
		return name, th
	}
	return DefaultName, s.fallback
}

func hasFonts(t *theme) bool { return len(t.manifest.Fonts) > 0 }

func hasIcons(t *theme) bool { return t.hasIcons() }

func hasIconFont(t *theme) bool {
	ic := t.icons()
	return ic != nil && ic.Type == ThemeTypeFont && ic.FontFile != ""
}

// ResolveIcon resolves a canonical icon name through the named widget theme.
// An unknown name, or a pack without icons, resolves through the default.
func (s *Store) ResolveIcon(themeName, canonicalName string) (ResolvedIcon, error) {
	name, th := s.effective(themeName, hasIcons)
	ic := th.icons()
	if ic == nil {
		return ResolvedIcon{}, fmt.Errorf("theme %q has no icon config", name)
	}
	switch ic.Type {
	case ThemeTypeFont:
		return s.resolveFontIcon(name, th, ic, canonicalName)
	case ThemeTypeImage:
		return s.resolveImageIcon(th, canonicalName)
	default:
		return ResolvedIcon{}, fmt.Errorf("theme %q has unknown icon type %q", name, ic.Type)
	}
}

func (s *Store) resolveFontIcon(effName string, th *theme, ic *manifestIcons, canonicalName string) (ResolvedIcon, error) {
	icon, ok := ic.Icons[canonicalName]
	if !ok {
		return ResolvedIcon{}, fmt.Errorf("icon %q not found in font theme", canonicalName)
	}
	fontFile := ""
	if ic.FontFile != "" {
		if th.dir != "" {
			fontFile = filepath.Join(th.dir, ic.FontFile)
		} else {
			fontFile = "embedded:default"
		}
	}
	return ResolvedIcon{
		Type:       ThemeTypeFont,
		CSSClass:   ic.ClassPrefix + icon.Class,
		Codepoint:  icon.Codepoint,
		FontFile:   fontFile,
		FontFamily: IconFontFamily(effName),
	}, nil
}

func (s *Store) resolveImageIcon(th *theme, canonicalName string) (ResolvedIcon, error) {
	if strings.ContainsAny(canonicalName, "/\\") || strings.Contains(canonicalName, "..") {
		return ResolvedIcon{}, fmt.Errorf("invalid icon name %q", canonicalName)
	}
	iconsDir := filepath.Join(th.dir, "widgets", "weather", "icons")
	for _, ext := range []string{".svg", ".png", ".jpg", ".webp"} {
		path := filepath.Join(iconsDir, canonicalName+ext)
		if !strings.HasPrefix(path, iconsDir+string(filepath.Separator)) {
			return ResolvedIcon{}, fmt.Errorf("invalid icon name %q", canonicalName)
		}
		if _, err := os.Stat(path); err == nil {
			return ResolvedIcon{
				Type:     ThemeTypeImage,
				FilePath: path,
			}, nil
		}
	}
	return ResolvedIcon{}, fmt.Errorf("icon %q not found in image theme %q", canonicalName, th.manifest.Name)
}

func (s *Store) GetFontData(themeName string) ([]byte, error) {
	th, ok := s.lookupTheme(themeName)
	if !ok {
		return nil, fmt.Errorf("theme %q not found", themeName)
	}
	ic := th.icons()
	if ic == nil || ic.Type != ThemeTypeFont || ic.FontFile == "" {
		return nil, fmt.Errorf("theme %q has no icon font file", themeName)
	}
	if th.dir == "" {
		if data, ok := embeddedFonts[ic.FontFile]; ok {
			return data, nil
		}
		return nil, fmt.Errorf("embedded font %q not found", ic.FontFile)
	}
	fontPath := filepath.Join(th.dir, ic.FontFile)
	return os.ReadFile(fontPath)
}

func (s *Store) GetDisplayFontData(themeName, fontName string) ([]byte, error) {
	th, ok := s.lookupTheme(themeName)
	if !ok {
		return nil, fmt.Errorf("theme %q not found", themeName)
	}
	for _, f := range th.manifest.Fonts {
		if f.Name == fontName {
			if th.dir == "" {
				if data, ok := embeddedFonts[f.File]; ok {
					return data, nil
				}
				return nil, fmt.Errorf("embedded font file %q not found", f.File)
			}
			fontPath := filepath.Join(th.dir, f.File)
			return os.ReadFile(fontPath)
		}
	}
	return nil, fmt.Errorf("font %q not found in theme %q", fontName, themeName)
}

// ---------- admin CRUD ----------

// validateThemeName enforces the directory-safe name rules for themes.
// Theme names become directory names on disk, so the rules match the
// flat-store rules used for files in internal/data.
func validateThemeName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("%w: empty", ErrInvalidName)
	}
	if len(trimmed) > 100 {
		return fmt.Errorf("%w: name exceeds 100 bytes", ErrInvalidName)
	}
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
		if r == 0 || unicode.IsControl(r) {
			return fmt.Errorf("%w: control character not allowed", ErrInvalidName)
		}
	}
	return nil
}

// Upload validates zipBytes as a theme archive and extracts it into the
// on-disk themes root as {rootDir}/{name}. The archive must contain a
// widget_theme.yaml at the zip root with a `name` field; the
// on-disk directory is named after the manifest's name (not the
// uploaded filename).
//
// Returns ErrConflict if a theme with the same name already exists
// (including the embedded default) — upload is reject-on-conflict.
func (s *Store) Upload(zipBytes []byte) (Info, error) {
	if s.rootDir == "" {
		return Info{}, errors.New("themes: store has no writable directory")
	}
	if int64(len(zipBytes)) > MaxUploadSize {
		return Info{}, fmt.Errorf("%w: exceeds %d bytes", ErrInvalidArchive, MaxUploadSize)
	}

	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return Info{}, fmt.Errorf("%w: %v", ErrInvalidArchive, err)
	}

	// 1. Sanity-check every entry up front (name, traversal) before we
	//    touch the filesystem.
	var manifestEntry *zip.File
	legacy := false
	for _, f := range zr.File {
		if err := validateZipMemberName(f.Name); err != nil {
			return Info{}, err
		}
		switch f.Name {
		case ManifestFile:
			manifestEntry = f
		case legacyManifestFile:
			legacy = true
		}
	}
	if manifestEntry == nil {
		if legacy {
			return Info{}, fmt.Errorf("%w: found legacy %s; rename it to %s and drop its type field",
				ErrInvalidArchive, legacyManifestFile, ManifestFile)
		}
		return Info{}, fmt.Errorf("%w: missing %s at archive root", ErrInvalidArchive, ManifestFile)
	}

	// 2. Parse the manifest and validate required fields.
	manifestBytes, err := readZipFile(manifestEntry)
	if err != nil {
		return Info{}, fmt.Errorf("%w: read %s: %v", ErrInvalidArchive, ManifestFile, err)
	}
	var m themeManifest
	if err := yaml.Unmarshal(manifestBytes, &m); err != nil {
		return Info{}, fmt.Errorf("%w: parse %s: %v", ErrInvalidArchive, ManifestFile, err)
	}
	if err := validateThemeName(m.Name); err != nil {
		return Info{}, err
	}

	// 3. Conflict check + reserve the slot. We use the map as the lock:
	//    hold the write lock for the whole upload so concurrent uploads
	//    of the same name can't race.
	s.mu.Lock()
	if _, exists := s.themes[m.Name]; exists {
		s.mu.Unlock()
		return Info{}, ErrConflict
	}
	destDir := filepath.Join(s.rootDir, m.Name)
	if _, err := os.Stat(destDir); err == nil {
		s.mu.Unlock()
		return Info{}, ErrConflict
	}

	// 4. Extract into a sibling temp dir, then atomic-rename into place.
	if err := os.MkdirAll(s.rootDir, 0o750); err != nil {
		s.mu.Unlock()
		return Info{}, fmt.Errorf("create themes dir: %w", err)
	}
	tmpDir, err := os.MkdirTemp(s.rootDir, ".upload-*")
	if err != nil {
		s.mu.Unlock()
		return Info{}, fmt.Errorf("create tmp dir: %w", err)
	}
	if err := extractZipInto(zr, tmpDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		s.mu.Unlock()
		return Info{}, err
	}
	if err := os.Rename(tmpDir, destDir); err != nil {
		_ = os.RemoveAll(tmpDir)
		s.mu.Unlock()
		return Info{}, fmt.Errorf("rename into place: %w", err)
	}

	// 5. Register in-memory.
	s.themes[m.Name] = &theme{manifest: m, dir: destDir}
	info := s.themeInfoLocked(m.Name, s.themes[m.Name])
	s.mu.Unlock()
	return info, nil
}

// Delete removes an uploaded theme by name. Builtin themes cannot be
// deleted. Returns ErrNotFound if the theme does not exist.
func (s *Store) Delete(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.builtins[name] {
		return ErrBuiltin
	}
	th, ok := s.themes[name]
	if !ok {
		return ErrNotFound
	}
	if th.dir == "" {
		// Defensive: a non-builtin with no dir shouldn't exist, but if
		// it somehow does we refuse to touch it.
		return ErrBuiltin
	}
	// Sanity check: the theme dir must live under rootDir.
	if s.rootDir == "" || !strings.HasPrefix(th.dir, s.rootDir+string(filepath.Separator)) {
		return fmt.Errorf("themes: theme dir %q is outside rootDir", th.dir)
	}
	if err := os.RemoveAll(th.dir); err != nil {
		return fmt.Errorf("remove theme dir: %w", err)
	}
	delete(s.themes, name)
	return nil
}

// Zip re-packages a theme into a zip archive for download. Uploaded
// themes are walked from disk; the embedded default is reconstructed
// from the bytes baked into the binary.
func (s *Store) Zip(name string) ([]byte, error) {
	th, ok := s.lookupTheme(name)
	if !ok {
		return nil, ErrNotFound
	}
	if th.dir == "" {
		return zipEmbeddedDefault()
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	err := filepath.Walk(th.dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(th.dir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		// Use forward slashes in zip entries regardless of OS.
		zipPath := filepath.ToSlash(rel)
		if info.IsDir() {
			_, err := zw.Create(zipPath + "/")
			return err
		}
		w, err := zw.Create(zipPath)
		if err != nil {
			return err
		}
		f, err := os.Open(path) //nolint:gosec // G304: path comes from filepath.Walk of th.dir, which is a controlled themes root.
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		_, err = io.Copy(w, f)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("zip theme: %w", err)
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("close zip: %w", err)
	}
	return buf.Bytes(), nil
}

// zipEmbeddedDefault reconstructs the builtin default theme as a zip
// archive from the bytes baked into the binary. The layout mirrors
// what Upload would produce on disk: widget_theme.yaml + font TTFs at the
// root.
func zipEmbeddedDefault() ([]byte, error) {
	entries := []struct {
		path string
		data []byte
	}{
		{ManifestFile, defaultThemeYAML},
		{"Inter-Regular.ttf", interRegularTTF},
		{"tabler-icons.ttf", defaultFontTTF},
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		w, err := zw.Create(e.path)
		if err != nil {
			return nil, fmt.Errorf("zip builtin: create %q: %w", e.path, err)
		}
		if _, err := w.Write(e.data); err != nil {
			return nil, fmt.Errorf("zip builtin: write %q: %w", e.path, err)
		}
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("zip builtin: close: %w", err)
	}
	return buf.Bytes(), nil
}

// validateZipMemberName rejects absolute paths, parent-traversal
// segments, and backslashes (Windows-style paths are not allowed in
// zips per APPNOTE, but archivers produce them anyway).
func validateZipMemberName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: empty member name", ErrInvalidArchive)
	}
	if strings.ContainsRune(name, '\\') {
		return fmt.Errorf("%w: member %q uses backslash", ErrInvalidArchive, name)
	}
	if strings.HasPrefix(name, "/") {
		return fmt.Errorf("%w: member %q is absolute", ErrInvalidArchive, name)
	}
	clean := filepath.ToSlash(filepath.Clean(name))
	if clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
		return fmt.Errorf("%w: member %q contains traversal", ErrInvalidArchive, name)
	}
	return nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rc.Close() }()
	return io.ReadAll(io.LimitReader(rc, MaxUploadSize))
}

// extractZipInto writes every regular file in zr under destDir,
// preserving directory structure. The caller is responsible for
// cleanup if this function returns an error.
func extractZipInto(zr *zip.Reader, destDir string) error {
	for _, f := range zr.File {
		target := filepath.Join(destDir, filepath.FromSlash(f.Name))
		// Belt-and-suspenders: ensure we didn't escape destDir even
		// after cleaning.
		if !strings.HasPrefix(target, destDir+string(filepath.Separator)) && target != destDir {
			return fmt.Errorf("%w: member %q escapes destination", ErrInvalidArchive, f.Name)
		}
		if strings.HasSuffix(f.Name, "/") {
			if err := os.MkdirAll(target, 0o750); err != nil {
				return fmt.Errorf("mkdir: %w", err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
			return fmt.Errorf("mkdir parent: %w", err)
		}
		if err := writeZipFile(f, target); err != nil {
			return err
		}
	}
	return nil
}

func writeZipFile(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("open zip entry %q: %w", f.Name, err)
	}
	defer func() { _ = rc.Close() }()
	out, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("create %q: %w", target, err)
	}
	defer func() { _ = out.Close() }()
	if _, err := io.Copy(out, io.LimitReader(rc, MaxUploadSize)); err != nil {
		return fmt.Errorf("write %q: %w", target, err)
	}
	return nil
}
