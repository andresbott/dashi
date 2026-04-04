package themes

const (
	ThemeTypeFont  = "font"
	ThemeTypeImage = "image"
)

// ThemeInfo is the metadata returned by the list API.
type ThemeInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// ResolvedIcon is the result of resolving a canonical icon name through a theme.
type ResolvedIcon struct {
	Type     string // "font" or "image"
	CSSClass string // for font themes: full CSS class (e.g. "ti ti-sun")
	FilePath string // for image themes: absolute path to the icon file on disk
}

// themeManifest represents the parsed theme.yaml file.
type themeManifest struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	Type        string     `yaml:"type"`
	Font        *fontTheme `yaml:"font,omitempty"`
}

type fontTheme struct {
	CSS         string            `yaml:"css"`
	ClassPrefix string            `yaml:"classPrefix"`
	Icons       map[string]string `yaml:"icons"`
}

// theme is the internal representation of a loaded theme.
type theme struct {
	manifest themeManifest
	dir      string // absolute path to theme directory (empty for embedded)
}
