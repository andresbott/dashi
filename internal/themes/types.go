package themes

const (
	ThemeTypeFont  = "font"
	ThemeTypeImage = "image"
)

// FontInfo describes a display font provided by a theme.
type FontInfo struct {
	Name string `json:"name"`
}

// ThemeInfo is the metadata returned by the list API.
type ThemeInfo struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Fonts       []FontInfo `json:"fonts"`
	HasIcons    bool       `json:"hasIcons"`
	IconType    string     `json:"iconType,omitempty"`
}

// ResolvedIcon is the result of resolving a canonical icon name through a theme.
type ResolvedIcon struct {
	Type      string
	CSSClass  string
	Codepoint string
	FilePath  string
	FontFile  string
}

// themeManifest represents the parsed theme.yaml file.
// Supports both new format (fonts + icons sections) and legacy format (type + font).
type themeManifest struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`

	// New format
	Fonts []manifestFont `yaml:"fonts,omitempty"`
	Icons *manifestIcons `yaml:"icons,omitempty"`

	// Legacy format (backwards compat)
	Type string     `yaml:"type,omitempty"`
	Font *fontTheme `yaml:"font,omitempty"`
}

type manifestFont struct {
	Name string `yaml:"name"`
	File string `yaml:"file"`
}

type manifestIcons struct {
	Type        string              `yaml:"type"`
	ClassPrefix string              `yaml:"classPrefix"`
	FontFile    string              `yaml:"fontFile,omitempty"`
	Icons       map[string]fontIcon `yaml:"icons"`
}

type fontIcon struct {
	Class     string `yaml:"class"`
	Codepoint string `yaml:"codepoint"`
}

// Legacy type — kept for backwards compat parsing
type fontTheme struct {
	CSS         string              `yaml:"css"`
	ClassPrefix string              `yaml:"classPrefix"`
	FontFile    string              `yaml:"fontFile,omitempty"`
	Icons       map[string]fontIcon `yaml:"icons"`
}

// theme is the internal representation of a loaded theme.
type theme struct {
	manifest themeManifest
	dir      string
}

// icons returns the icon config, handling both new and legacy format.
func (t *theme) icons() *manifestIcons {
	if t.manifest.Icons != nil {
		return t.manifest.Icons
	}
	if t.manifest.Font != nil {
		return &manifestIcons{
			Type:        ThemeTypeFont,
			ClassPrefix: t.manifest.Font.ClassPrefix,
			FontFile:    t.manifest.Font.FontFile,
			Icons:       t.manifest.Font.Icons,
		}
	}
	if t.manifest.Type == ThemeTypeImage {
		return &manifestIcons{Type: ThemeTypeImage}
	}
	return nil
}

func (t *theme) hasIcons() bool {
	ic := t.icons()
	return ic != nil && (len(ic.Icons) > 0 || ic.Type == ThemeTypeImage)
}
