package widgetthemes

const (
	ThemeTypeFont  = "font"
	ThemeTypeImage = "image"
)

// ManifestFile is the manifest every widget theme pack carries at its root.
const ManifestFile = "widget_theme.yaml"

// legacyManifestFile is the pre-rework manifest name. Packs carrying only it
// are rejected on upload with a message that points at ManifestFile.
const legacyManifestFile = "theme.yaml"

// DefaultName is the builtin widget theme that unknown names and missing
// parts (fonts, icons) fall back to.
const DefaultName = "default"

// FontInfo describes a display font provided by a theme.
type FontInfo struct {
	Name string `json:"name"`
}

// Info is the metadata returned by the list API.
type Info struct {
	Name        string     `json:"name"`
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
	// FontFamily is the CSS family of the icon font, IconFontFamily of the
	// theme the icon actually resolved through. Set for font icons only.
	FontFamily string
}

// themeManifest represents the parsed widget_theme.yaml file.
type themeManifest struct {
	Name        string         `yaml:"name"`
	Description string         `yaml:"description"`
	Fonts       []manifestFont `yaml:"fonts,omitempty"`
	Icons       *manifestIcons `yaml:"icons,omitempty"`
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
