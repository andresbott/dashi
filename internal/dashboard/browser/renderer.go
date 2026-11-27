// Package browser renders dashboards as HTML for real browsers: Tailwind
// classes, flex layout, and per-widget ES modules that refresh values in
// place. Its sibling package internal/dashboard/static renders the same
// dashboards for litehtml (e-ink/PNG), which supports a much smaller CSS
// subset and no JavaScript.
package browser

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"net/url"
	"sort"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/andresbott/dashi/internal/widgetthemes"
)

//go:embed page.html
var pageHTML string

var pageTmpl = template.Must(template.New("page").Parse(pageHTML))

// Renderer assembles a full browser HTML page from a dashboard page.
type Renderer struct {
	registry *widgets.Registry
	// jsTypes holds the widget types that ship a browser ES module, so
	// the shell only emits script tags that will resolve.
	jsTypes map[string]bool
}

// NewRenderer creates a Renderer. jsTypes may be nil, in which case a
// script tag is emitted for every widget type on the page — callers that
// know which modules ship JS should pass the set (see browser.Assets).
func NewRenderer(registry *widgets.Registry, jsTypes map[string]bool) *Renderer {
	return &Renderer{registry: registry, jsTypes: jsTypes}
}

// RenderData holds everything needed to render one dashboard page.
type RenderData struct {
	Name        string
	DashboardID string
	// ThemeID and WidgetTheme are linked as /_dashi/theme/{id}.css and
	// /_dashi/widget-theme/{name}.css. Empty means "default".
	ThemeID     string
	WidgetTheme string
	ColorMode   string
	MaxWidth    string
	HAlign      string
	VAlign      string
	CustomCSS   string
	QueryParams map[string]string
	Rows        []dashboard.Row
	PageIndex   int
	TotalPages  int
	PageNames   []string       // one per dashboard page, for the page nav
	Palette     themes.Palette // carried through to widget renderers
	// EditURL is the target of the Edit button that opens this dashboard
	// in the editor; empty draws no button.
	EditURL string
}

type pageData struct {
	Name        string
	ColorMode   string
	MaxWidth    string
	Placement   dashboard.ContainerPlacement
	ThemeID     string
	WidgetTheme string
	CustomCSS   template.CSS
	Rows        []rowData
	Pages       []pageLink
	WidgetTypes []string
	EditURL     string // html/template escapes it and neutralises unsafe schemes
}

// pageLink is one entry in the page navigation. It is only populated for
// multi-page dashboards; single-page dashboards render no nav at all.
type pageLink struct {
	Index   int
	Name    string
	Current bool
}

type rowData struct {
	Title   string
	Height  string
	Width   string
	Widgets []widgetData
}

type widgetData struct {
	WidthPercent      float64
	MarginLeftPercent float64
	HTML              template.HTML
}

// Render writes the complete HTML page for data to w.
func (r *Renderer) Render(w io.Writer, data RenderData) error {
	colorMode := data.ColorMode
	if colorMode != "dark" {
		colorMode = "light"
	}
	widgetTheme := orDefault(data.WidgetTheme, widgetthemes.DefaultName)

	ctx := widgets.RenderContext{
		DashboardID: data.DashboardID,
		WidgetTheme: widgetTheme,
		ColorMode:   colorMode,
		Palette:     data.Palette,
		QueryParams: data.QueryParams,
		PageIndex:   data.PageIndex,
		TotalPages:  data.TotalPages,
	}

	pd := pageData{
		Name:      data.Name,
		ColorMode: colorMode,
		MaxWidth:  data.MaxWidth,
		Placement: dashboard.PlaceContainer(data.HAlign, data.VAlign),
		// Each reference is escaped to a single path segment, so a stored
		// value can never point the link outside its asset route.
		ThemeID:     url.PathEscape(orDefault(data.ThemeID, themes.DefaultID)),
		WidgetTheme: url.PathEscape(widgetTheme),
		CustomCSS:   template.CSS(dashboard.SanitizeCustomCSS(data.CustomCSS)), //nolint:gosec // G203: owner-uploaded stylesheet, </style> breakout neutralised
		Pages:       buildPageLinks(data.PageIndex, data.TotalPages, data.PageNames),
		EditURL:     data.EditURL,
	}

	present := make(map[string]bool)
	for _, row := range data.Rows {
		hasExplicitHeight := row.Height != "" && row.Height != "auto"
		if len(row.Widgets) == 0 && !hasExplicitHeight {
			continue
		}
		rd := rowData{Title: row.Title, Height: row.Height, Width: row.Width}
		placements := dashboard.PlaceRow(row.Widgets)
		for i, widget := range row.Widgets {
			html, err := r.registry.RenderBrowser(widget.Type, widget.Config, ctx)
			if err != nil {
				return fmt.Errorf("render widget %s (%s): %w", widget.ID, widget.Type, err)
			}
			p := placements[i]
			rd.Widgets = append(rd.Widgets, widgetData{
				WidthPercent:      float64(p.Width) / float64(dashboard.GridColumns) * 100.0,
				MarginLeftPercent: float64(p.Gap) / float64(dashboard.GridColumns) * 100.0,
				HTML:              html,
			})
			present[widget.Type] = true
		}
		pd.Rows = append(pd.Rows, rd)
	}

	for widgetType := range present {
		if r.jsTypes == nil || r.jsTypes[widgetType] {
			pd.WidgetTypes = append(pd.WidgetTypes, widgetType)
		}
	}
	sort.Strings(pd.WidgetTypes) // deterministic output for tests and caching

	return pageTmpl.Execute(w, pd)
}

// buildPageLinks returns the page-navigation entries for a multi-page
// dashboard, or nil when the dashboard has a single page (in which case the
// shell renders no nav). Navigation is plain <a href="?page=N"> links: the
// e-ink stack navigates by swipe headers instead and is untouched by this.
func buildPageLinks(current, total int, names []string) []pageLink {
	if total < 2 {
		return nil
	}
	links := make([]pageLink, 0, total)
	for i := range total {
		name := ""
		if i < len(names) {
			name = names[i]
		}
		if name == "" {
			name = fmt.Sprintf("Page %d", i+1)
		}
		links = append(links, pageLink{Index: i, Name: name, Current: i == current})
	}
	return links
}

// orDefault returns def for an empty reference. Both asset routes resolve
// unknown names to their default too.
func orDefault(ref, def string) string {
	if ref == "" {
		return def
	}
	return ref
}
