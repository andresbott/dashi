package bookmark

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"

	"github.com/andresbott/dashi/internal/data"
	bookmarkfiles "github.com/andresbott/dashi/internal/data/bookmarks"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/andresbott/dashi/internal/widgetthemes"
)

// Direction values of bookmarkgroupConfig.Direction. Anything else,
// including empty, renders horizontal.
const (
	directionHorizontal = "horizontal"
	directionVertical   = "vertical"
)

// Source values of bookmarkgroupConfig.Source. Anything else, including
// empty, uses the bookmarks stored in the dashboard.
const (
	sourceDashboard = "dashboard"
	sourceFile      = "file"
)

// bookmarkgroupConfig is the bookmarkgroup widget's config: a list of
// bookmarks sharing one layout. textBelow and newTab mean the same as on a
// single bookmark and apply to every item.
type bookmarkgroupConfig struct {
	Bookmarks []groupItem `json:"bookmarks"`
	// Source picks where the bookmarks come from: the Bookmarks list above
	// ("dashboard", the default) or the content file named by File
	// ("file"). Bookmarks is kept, unused, while the source is a file.
	Source    string `json:"source,omitempty"`
	File      string `json:"file,omitempty"`
	TextBelow bool   `json:"textBelow,omitempty"`
	NewTab    bool   `json:"newTab,omitempty"`
	// Direction lays the items out in a wrapping row ("horizontal", the
	// default) or a column ("vertical").
	Direction string `json:"direction,omitempty"`
	// Columns is how many bookmarks a horizontal group puts in one row
	// before starting the next, in equal-width columns. 0 (the default)
	// fits as many as the width allows. Vertical groups ignore it.
	Columns int `json:"columns,omitempty"`
}

// maxGroupColumns caps Columns; the dashboard grid itself has 12 columns.
const maxGroupColumns = 12

// columns returns the row length a horizontal group renders with: 0 for
// "fit to width", otherwise 1..maxGroupColumns. Vertical groups get 0.
func (c bookmarkgroupConfig) columns() int {
	if c.vertical() || c.Columns <= 0 {
		return 0
	}
	return min(c.Columns, maxGroupColumns)
}

// groupItem is one bookmark of a group. Its name must not end in "Config":
// the contract test treats every such struct as a widget config.
type groupItem struct {
	URL      string `json:"url"`
	Icon     string `json:"icon"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

func (c bookmarkgroupConfig) vertical() bool { return c.Direction == directionVertical }

func (c bookmarkgroupConfig) fromFile() bool { return c.Source == sourceFile }

func parseGroupConfig(config json.RawMessage) (bookmarkgroupConfig, error) {
	var cfg bookmarkgroupConfig
	if len(config) > 0 {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return cfg, fmt.Errorf("bookmarkgroup config: %w", err)
		}
	}
	return cfg, nil
}

//go:embed group.html
var groupHTML string

var groupTmpl = template.Must(template.New("bookmarkgroup").Parse(groupHTML))

// FileSource loads the bookmarks of a content file by name; implemented by
// *bookmarkfiles.Store, which caches each file for its CacheTTL.
type FileSource interface {
	Bookmarks(name string) ([]bookmarkfiles.Item, error)
}

// GroupModule implements widgets.Module for the bookmarkgroup widget. It
// shares the bookmark's URL rule, icon resolution and link template.
type GroupModule struct {
	widgets.NoopModule
	themes *widgetthemes.Store // resolves each item's icon for the browser stack; nil renders no icons
	files  FileSource          // bookmarks of source "file"; nil renders those groups with an error note
}

// NewGroupModule constructs a bookmarkgroup Module.
func NewGroupModule(themes *widgetthemes.Store, files FileSource) *GroupModule {
	return &GroupModule{themes: themes, files: files}
}

// items returns the bookmarks to render: the dashboard's own, or those of
// the configured file. A file that cannot be loaded yields a short message
// for the widget to show instead of its bookmarks.
func (m *GroupModule) items(cfg bookmarkgroupConfig) ([]groupItem, string) {
	if !cfg.fromFile() {
		return cfg.Bookmarks, ""
	}
	if cfg.File == "" {
		return nil, "No bookmarks file selected"
	}
	if m.files == nil {
		return nil, "Bookmarks files are not available"
	}
	loaded, err := m.files.Bookmarks(cfg.File)
	switch {
	case errors.Is(err, data.ErrNotFound):
		return nil, fmt.Sprintf("Bookmarks file %q not found", cfg.File)
	case err != nil:
		return nil, fmt.Sprintf("Bookmarks file %q could not be read", cfg.File)
	}
	out := make([]groupItem, len(loaded))
	for i, b := range loaded {
		out[i] = groupItem{URL: b.URL, Icon: b.Icon, Title: b.Title, Subtitle: b.Subtitle}
	}
	return out, ""
}

// Type returns the widget type string.
func (m *GroupModule) Type() string { return "bookmarkgroup" }

// ValidateConfig implements widgets.ConfigValidator: it rejects a config
// that does not decode or holds a bookmark whose url fails validateURL.
// Items are numbered from 1, as the editor lists them.
func (m *GroupModule) ValidateConfig(config json.RawMessage) error {
	cfg, err := parseGroupConfig(config)
	if err != nil {
		return err
	}
	if cfg.Columns < 0 || cfg.Columns > maxGroupColumns {
		return fmt.Errorf("columns must be between 0 and %d", maxGroupColumns)
	}
	if cfg.Source != "" && cfg.Source != sourceDashboard && cfg.Source != sourceFile {
		return fmt.Errorf("source must be %q or %q", sourceDashboard, sourceFile)
	}
	for i, item := range cfg.Bookmarks {
		if err := validateURL(item.URL); err != nil {
			return fmt.Errorf("bookmark %d %w", i+1, err)
		}
	}
	return nil
}

type groupImageLink struct {
	URL      string
	Title    string
	Subtitle string
}

type groupImageData struct {
	Vertical bool
	// ItemWidth is each link's width ("33.3333%") when the row length is
	// fixed; empty lets links size to their text and wrap.
	ItemWidth string
	Links     []groupImageLink
	// Error replaces the links when a bookmarks file cannot be loaded.
	Error   string
	Palette themes.Palette
}

// Renderer returns the image-stack renderer: titles and subtitles only, in
// a row that wraps or a column. litehtml has no CSS grid, so a fixed row
// length is a percentage width per link.
func (m *GroupModule) Renderer() widgets.StaticRenderer {
	return func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
		cfg, err := parseGroupConfig(config)
		if err != nil {
			return "", err
		}
		items, msg := m.items(cfg)
		data := groupImageData{Vertical: cfg.vertical(), Error: msg, Palette: ctx.EffectivePalette()}
		if n := cfg.columns(); n > 0 {
			data.ItemWidth = fmt.Sprintf("%.4f%%", 100/float64(n))
		}
		for _, item := range items {
			data.Links = append(data.Links, groupImageLink{URL: sanitizeURL(item.URL), Title: item.Title, Subtitle: item.Subtitle})
		}
		var buf bytes.Buffer
		if err := groupTmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("bookmarkgroup render: %w", err)
		}
		return template.HTML(buf.String()), nil
	}
}

type groupBrowserData struct {
	Vertical bool
	// Columns fixes a horizontal row's length; 0 fits as many as the width
	// allows.
	Columns int
	Links   []browserData
	// Error replaces the links when a bookmarks file cannot be loaded.
	Error string
}

// RenderBrowser renders the group for the browser stack: one container
// holding a bookmark link per item, in stored order. Each link is the same
// link a single bookmark widget renders, placed in an equal-width grid
// column the way separate widgets sit in grid cells.
func (m *GroupModule) RenderBrowser(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
	cfg, err := parseGroupConfig(config)
	if err != nil {
		return "", err
	}
	items, msg := m.items(cfg)
	data := groupBrowserData{Vertical: cfg.vertical(), Columns: cfg.columns(), Error: msg}
	for _, item := range items {
		data.Links = append(data.Links, browserData{
			URL:       sanitizeURL(item.URL),
			Title:     item.Title,
			Subtitle:  item.Subtitle,
			TextBelow: cfg.TextBelow,
			NewTab:    cfg.NewTab,
			Icon:      resolveBrowserIcon(m.themes, ctx.WidgetTheme, ctx.DashboardID, item.Icon),
			InGroup:   true,
		})
	}
	var buf bytes.Buffer
	if err := browserTmpl.ExecuteTemplate(&buf, "group", data); err != nil {
		return "", fmt.Errorf("bookmarkgroup browser render: %w", err)
	}
	return template.HTML(buf.String()), nil
}
