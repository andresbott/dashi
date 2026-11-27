package bookmark

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"

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

// bookmarkgroupConfig is the bookmarkgroup widget's config: a list of
// bookmarks sharing one layout. textBelow and newTab mean the same as on a
// single bookmark and apply to every item.
type bookmarkgroupConfig struct {
	Bookmarks []groupItem `json:"bookmarks"`
	TextBelow bool        `json:"textBelow,omitempty"`
	NewTab    bool        `json:"newTab,omitempty"`
	// Direction lays the items out in a wrapping row ("horizontal", the
	// default) or a column ("vertical").
	Direction string `json:"direction,omitempty"`
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

// GroupModule implements widgets.Module for the bookmarkgroup widget. It
// shares the bookmark's URL rule, icon resolution and link template.
type GroupModule struct {
	widgets.NoopModule
	themes *widgetthemes.Store // resolves each item's icon for the browser stack; nil renders no icons
}

// NewGroupModule constructs a bookmarkgroup Module.
func NewGroupModule(themes *widgetthemes.Store) *GroupModule {
	return &GroupModule{themes: themes}
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
	Links    []groupImageLink
	Palette  themes.Palette
}

// Renderer returns the image-stack renderer: titles and subtitles only, in
// a row that wraps or a column.
func (m *GroupModule) Renderer() widgets.StaticRenderer {
	return func(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
		cfg, err := parseGroupConfig(config)
		if err != nil {
			return "", err
		}
		data := groupImageData{Vertical: cfg.vertical(), Palette: ctx.EffectivePalette()}
		for _, item := range cfg.Bookmarks {
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
	Links    []browserData
}

// RenderBrowser renders the group for the browser stack: one container
// holding a bookmark link per item, in stored order.
func (m *GroupModule) RenderBrowser(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
	cfg, err := parseGroupConfig(config)
	if err != nil {
		return "", err
	}
	data := groupBrowserData{Vertical: cfg.vertical()}
	for _, item := range cfg.Bookmarks {
		data.Links = append(data.Links, browserData{
			URL:       sanitizeURL(item.URL),
			Title:     item.Title,
			Subtitle:  item.Subtitle,
			TextBelow: cfg.TextBelow,
			NewTab:    cfg.NewTab,
			Icon:      resolveBrowserIcon(m.themes, ctx.WidgetTheme, ctx.DashboardID, item.Icon),
			InGroup:   true,
			Tile:      cfg.TextBelow && !cfg.vertical(),
		})
	}
	var buf bytes.Buffer
	if err := browserTmpl.ExecuteTemplate(&buf, "group", data); err != nil {
		return "", fmt.Errorf("bookmarkgroup browser render: %w", err)
	}
	return template.HTML(buf.String()), nil
}
