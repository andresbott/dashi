package bookmark

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/andresbott/dashi/internal/widgets"
	"github.com/andresbott/dashi/internal/widgetthemes"
)

//go:embed browser.html
var browserHTML string

// browserTmpl holds the "link" template (one bookmark) and the "group"
// template (the bookmarkgroup widget, a list of links).
var browserTmpl = template.Must(template.New("bookmark-browser").Parse(browserHTML))

// browserIcon is the resolved config.icon. At most one of Glyph and Src is
// set; both empty renders no icon element at all (never tofu).
type browserIcon struct {
	Glyph      string // font icon: the code point as a one-rune string
	FontFamily string // font icon: the widget theme's icon-font family
	Src        string // image icon: URL of the widget-theme icon route
}

type browserData struct {
	URL       string
	Title     string
	Subtitle  string
	TextBelow bool
	NewTab    bool
	Icon      browserIcon
	// InGroup drops the link's own negative margin: the group container
	// carries it once, so the links' gaps stay even.
	InGroup bool
}

// RenderBrowser renders the bookmark for the browser stack: one link
// holding the icon and the title/subtitle block. It needs no JavaScript
// and no stylesheet of its own.
func (m *Module) RenderBrowser(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
	var cfg bookmarkConfig
	if len(config) > 0 {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return "", fmt.Errorf("bookmark config: %w", err)
		}
	}

	data := browserData{
		URL:       sanitizeURL(cfg.URL),
		Title:     cfg.Title,
		Subtitle:  cfg.Subtitle,
		TextBelow: cfg.TextBelow,
		NewTab:    cfg.NewTab,
		Icon:      resolveBrowserIcon(m.themes, ctx.WidgetTheme, ctx.DashboardID, cfg.Icon),
	}

	var buf bytes.Buffer
	if err := browserTmpl.ExecuteTemplate(&buf, "link", data); err != nil {
		return "", fmt.Errorf("bookmark browser render: %w", err)
	}
	return template.HTML(buf.String()), nil
}

// resolveBrowserIcon resolves a configured icon name through the widget
// theme, the way weather's resolveIconHTML does, but returns values for the
// template instead of hand-built HTML. The name is tried as stored, then
// without the editor's "ti-" prefix. Anything that does not resolve yields
// the zero browserIcon, so the widget renders without an icon.
func resolveBrowserIcon(store *widgetthemes.Store, themeName, dashboardID, name string) browserIcon {
	// The two image kinds the editor's picker offers next to Tabler names
	// (webui/src/lib/iconUtils.ts): same URLs as the editor uses.
	if file, ok := strings.CutPrefix(name, "selfhst:"); ok {
		return selfhstIcon(file)
	}
	if file, ok := strings.CutPrefix(name, "dashboard:"); ok {
		if file == "" || dashboardID == "" {
			return browserIcon{}
		}
		return browserIcon{Src: "/api/v0/dashboards/" + url.PathEscape(dashboardID) + "/assets/" + url.PathEscape(file)}
	}
	if store == nil || name == "" {
		return browserIcon{}
	}
	if themeName == "" {
		themeName = widgetthemes.DefaultName
	}
	candidates := []string{name}
	if trimmed := strings.TrimPrefix(name, "ti-"); trimmed != name && trimmed != "" {
		candidates = append(candidates, trimmed)
	}
	for _, c := range candidates {
		resolved, err := store.ResolveIcon(themeName, c)
		if err != nil {
			continue
		}
		switch resolved.Type {
		case widgetthemes.ThemeTypeFont:
			cp, err := strconv.ParseUint(resolved.Codepoint, 16, 32)
			if err != nil || cp > unicode.MaxRune {
				continue
			}
			return browserIcon{Glyph: string(rune(cp)), FontFamily: resolved.FontFamily}
		case widgetthemes.ThemeTypeImage:
			// FilePath is a server filesystem path; the browser fetches the
			// same icon through the widget-theme icon route, which resolves
			// the name with the same theme fallback.
			return browserIcon{Src: "/api/v0/widget-themes/" + url.PathEscape(themeName) + "/icons/" + url.PathEscape(c)}
		}
	}
	return browserIcon{}
}

// selfhstIcon points at the selfh.st icon set on jsDelivr, keyed by file
// name ("galaxus.webp"); the extension picks the folder, webp by default.
func selfhstIcon(file string) browserIcon {
	if file == "" || strings.ContainsAny(file, `/\`) || strings.Contains(file, "..") {
		return browserIcon{}
	}
	ext := "webp"
	if i := strings.LastIndexByte(file, '.'); i >= 0 && i < len(file)-1 {
		ext = file[i+1:]
	}
	return browserIcon{Src: "https://cdn.jsdelivr.net/gh/selfhst/icons@main/" + url.PathEscape(ext) + "/" + url.PathEscape(file)}
}
