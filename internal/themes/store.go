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

//go:embed defaults/tabler-icons.ttf
var defaultFontTTF []byte

//go:embed defaults/Inter-Regular.ttf
var interRegularTTF []byte

//go:embed defaults/Inter-Bold.ttf
var interBoldTTF []byte

//go:embed defaults/Inter-Italic.ttf
var interItalicTTF []byte

//go:embed defaults/Inter-BoldItalic.ttf
var interBoldItalicTTF []byte

var embeddedFonts = map[string][]byte{
	"tabler-icons.ttf":  defaultFontTTF,
	"Inter-Regular.ttf": interRegularTTF,
}

type Store struct {
	themes map[string]*theme
}

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
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		themeDir := filepath.Join(dir, entry.Name())
		manifestPath := filepath.Join(themeDir, "theme.yaml")
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

func (s *Store) List() []ThemeInfo {
	result := make([]ThemeInfo, 0, len(s.themes))
	for name, th := range s.themes {
		fonts := make([]FontInfo, len(th.manifest.Fonts))
		for i, f := range th.manifest.Fonts {
			fonts[i] = FontInfo{Name: f.Name}
		}
		iconType := ""
		if ic := th.icons(); ic != nil {
			iconType = ic.Type
		}
		result = append(result, ThemeInfo{
			Name:        name,
			Description: th.manifest.Description,
			Fonts:       fonts,
			HasIcons:    th.hasIcons(),
			IconType:    iconType,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func (s *Store) Get(name string) (ThemeInfo, bool) {
	th, ok := s.themes[name]
	if !ok {
		return ThemeInfo{}, false
	}
	fonts := make([]FontInfo, len(th.manifest.Fonts))
	for i, f := range th.manifest.Fonts {
		fonts[i] = FontInfo{Name: f.Name}
	}
	iconType := ""
	if ic := th.icons(); ic != nil {
		iconType = ic.Type
	}
	return ThemeInfo{
		Name:        name,
		Description: th.manifest.Description,
		Fonts:       fonts,
		HasIcons:    th.hasIcons(),
		IconType:    iconType,
	}, true
}

func (s *Store) ResolveIcon(themeName, canonicalName string) (ResolvedIcon, error) {
	th, ok := s.themes[themeName]
	if !ok {
		return ResolvedIcon{}, fmt.Errorf("theme %q not found", themeName)
	}
	ic := th.icons()
	if ic == nil {
		return ResolvedIcon{}, fmt.Errorf("theme %q has no icon config", themeName)
	}
	switch ic.Type {
	case ThemeTypeFont:
		return s.resolveFontIcon(th, ic, canonicalName)
	case ThemeTypeImage:
		return s.resolveImageIcon(th, canonicalName)
	default:
		return ResolvedIcon{}, fmt.Errorf("theme %q has unknown icon type %q", themeName, ic.Type)
	}
}

func (s *Store) resolveFontIcon(th *theme, ic *manifestIcons, canonicalName string) (ResolvedIcon, error) {
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
		Type:      ThemeTypeFont,
		CSSClass:  ic.ClassPrefix + icon.Class,
		Codepoint: icon.Codepoint,
		FontFile:  fontFile,
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
	th, ok := s.themes[themeName]
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
	th, ok := s.themes[themeName]
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

// ListBackgrounds returns the filenames of background images in a theme's backgrounds/ directory.
func (s *Store) ListBackgrounds(themeName string) []string {
	th, ok := s.themes[themeName]
	if !ok || th.dir == "" {
		return nil
	}
	bgDir := filepath.Join(th.dir, "backgrounds")
	entries, err := os.ReadDir(bgDir)
	if err != nil {
		return nil
	}
	imageExts := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".svg": true}
	var result []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if imageExts[ext] {
			result = append(result, e.Name())
		}
	}
	return result
}

// GetBackgroundData returns the raw bytes of a background image file.
func (s *Store) GetBackgroundData(themeName, fileName string) ([]byte, error) {
	th, ok := s.themes[themeName]
	if !ok || th.dir == "" {
		return nil, fmt.Errorf("theme %q not found or has no directory", themeName)
	}
	if strings.ContainsAny(fileName, "/\\") || strings.Contains(fileName, "..") {
		return nil, fmt.Errorf("invalid filename %q", fileName)
	}
	bgPath := filepath.Join(th.dir, "backgrounds", fileName)
	if !strings.HasPrefix(bgPath, filepath.Join(th.dir, "backgrounds")+string(filepath.Separator)) {
		return nil, fmt.Errorf("invalid filename %q", fileName)
	}
	return os.ReadFile(bgPath)
}
