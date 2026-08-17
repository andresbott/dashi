package themes

const (
	ThemeTypeFont  = "font"
	ThemeTypeImage = "image"
)

// Theme kinds — the top-level classification of a theme as shown in the
// admin UI. Required field in theme.yaml.
const (
	ThemeKindTheme = "theme" // full theme (fonts + icons + backgrounds)
	ThemeKindIcon  = "icon"  // icon pack
	ThemeKindStyle = "style" // style/font-only pack
)

// AllowedThemeKinds is the closed set of accepted values for the
// top-level `type` field in theme.yaml. No inference — the field is
// required and must be one of these.
var AllowedThemeKinds = map[string]bool{
	ThemeKindTheme: true,
	ThemeKindIcon:  true,
	ThemeKindStyle: true,
}

// FontInfo describes a display font provided by a theme.
type FontInfo struct {
	Name string `json:"name"`
}

// ThemeInfo is the metadata returned by the list API.
type ThemeInfo struct {
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Description string     `json:"description"`
	Fonts       []FontInfo `json:"fonts"`
	HasIcons    bool       `json:"hasIcons"`
	IconType    string     `json:"iconType,omitempty"`
	Builtin     bool       `json:"builtin,omitempty"`
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
type themeManifest struct {
	Name        string            `yaml:"name"`
	Type        string            `yaml:"type"` // ThemeKindIcon or ThemeKindStyle; required.
	Description string            `yaml:"description"`
	Fonts       []manifestFont    `yaml:"fonts,omitempty"`
	Icons       *manifestIcons    `yaml:"icons,omitempty"`
	Colors      map[string]string `yaml:"colors,omitempty"`
	ColorsDark  map[string]string `yaml:"colorsDark,omitempty"`
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

// theme is the internal representation of a loaded theme.
type theme struct {
	manifest themeManifest
	dir      string // empty for the embedded default
}

func (t *theme) icons() *manifestIcons { return t.manifest.Icons }

func (t *theme) hasIcons() bool {
	ic := t.icons()
	return ic != nil && (len(ic.Icons) > 0 || ic.Type == ThemeTypeImage)
}
