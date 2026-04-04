package themes

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed defaults/theme.yaml
var defaultThemeYAML []byte

// Store manages theme loading and icon resolution.
type Store struct {
	themes map[string]*theme
}

// NewStore creates a Store, loading the embedded default theme
// and any user themes found in themesDir.
// If themesDir is empty, only the embedded default is loaded.
func NewStore(themesDir string) *Store {
	s := &Store{themes: make(map[string]*theme)}
	s.loadEmbeddedDefault()
	if themesDir != "" {
		s.loadUserThemes(themesDir)
	}
	return s
}

func (s *Store) loadEmbeddedDefault() {
	var m themeManifest
	if err := yaml.Unmarshal(defaultThemeYAML, &m); err != nil {
		panic(fmt.Sprintf("embedded default theme.yaml is invalid: %v", err))
	}
	s.themes[m.Name] = &theme{manifest: m}
}

func (s *Store) loadUserThemes(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return // directory doesn't exist or unreadable — no user themes
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		themeDir := filepath.Join(dir, entry.Name())
		manifestPath := filepath.Join(themeDir, "theme.yaml")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue // no manifest — skip
		}
		var m themeManifest
		if err := yaml.Unmarshal(data, &m); err != nil {
			continue // invalid manifest — skip
		}
		// Use directory name as theme key (overrides embedded if same name)
		s.themes[entry.Name()] = &theme{manifest: m, dir: themeDir}
	}
}

// List returns metadata for all loaded themes, sorted by name.
func (s *Store) List() []ThemeInfo {
	result := make([]ThemeInfo, 0, len(s.themes))
	for name, th := range s.themes {
		result = append(result, ThemeInfo{
			Name:        name,
			Description: th.manifest.Description,
			Type:        th.manifest.Type,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// Get returns the theme info for the given name.
func (s *Store) Get(name string) (ThemeInfo, bool) {
	th, ok := s.themes[name]
	if !ok {
		return ThemeInfo{}, false
	}
	return ThemeInfo{
		Name:        name,
		Description: th.manifest.Description,
		Type:        th.manifest.Type,
	}, true
}

// ResolveIcon resolves a canonical icon name to a concrete icon reference.
func (s *Store) ResolveIcon(themeName, canonicalName string) (ResolvedIcon, error) {
	th, ok := s.themes[themeName]
	if !ok {
		return ResolvedIcon{}, fmt.Errorf("theme %q not found", themeName)
	}

	switch th.manifest.Type {
	case ThemeTypeFont:
		return s.resolveFontIcon(th, canonicalName)
	case ThemeTypeImage:
		return s.resolveImageIcon(th, canonicalName)
	default:
		return ResolvedIcon{}, fmt.Errorf("theme %q has unknown type %q", themeName, th.manifest.Type)
	}
}

func (s *Store) resolveFontIcon(th *theme, canonicalName string) (ResolvedIcon, error) {
	if th.manifest.Font == nil {
		return ResolvedIcon{}, fmt.Errorf("font theme missing font config")
	}
	suffix, ok := th.manifest.Font.Icons[canonicalName]
	if !ok {
		return ResolvedIcon{}, fmt.Errorf("icon %q not found in font theme", canonicalName)
	}
	return ResolvedIcon{
		Type:     ThemeTypeFont,
		CSSClass: th.manifest.Font.ClassPrefix + suffix,
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
