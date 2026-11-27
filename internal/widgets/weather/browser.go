package weather

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"strconv"
	"unicode"

	weatherpkg "github.com/andresbott/dashi/internal/providers/weather"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/andresbott/dashi/internal/widgetthemes"
)

//go:embed browser.html
var browserHTML string

var browserTmpl = template.Must(template.New("weather-browser").Parse(browserHTML))

// browserIconFallback mirrors the image stack's fallback for icons the
// theme cannot resolve.
const browserIconFallback = template.HTML(`<i class="ti ti-cloud-question" aria-hidden="true"></i>`)

// RenderBrowser renders the weather widget for the browser stack. Content
// matches the image renderer (same buildWeatherData); only the markup
// differs. The widget ships no JS: values are refreshed by a page reload.
func (m *Module) RenderBrowser(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
	return renderBrowser(m.client, m.themeStore, config, ctx)
}

// RenderBrowser renders weather-compact for the browser stack, forcing
// compact mode the same way the image renderer does.
func (m *CompactModule) RenderBrowser(config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
	modified, err := forceCompact(config)
	if err != nil {
		return "", err
	}
	return renderBrowser(m.client, m.themeStore, modified, ctx)
}

func renderBrowser(client *weatherpkg.Client, themeStore *widgetthemes.Store, config json.RawMessage, ctx widgets.RenderContext) (template.HTML, error) {
	cfg, wd, err := loadWeather(client, config)
	if err != nil {
		return "", err
	}

	themeName := effectiveThemeName(ctx)
	data := buildWeatherData(cfg, wd, ctx, func(canonical string) template.HTML {
		return browserIconHTML(themeStore, themeName, canonical)
	})

	var buf bytes.Buffer
	if err := browserTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("weather browser render: %w", err)
	}
	return template.HTML(buf.String()), nil
}

// browserIconHTML resolves an icon for the browser. Unlike resolveIconHTML
// it sets no font size, so each slot in browser.html sizes its icon with
// the surrounding text size. Styles are inline because Tailwind only scans
// browser.html, never Go source. Font icons use the family served by
// /_dashi/widget-theme/{name}.css; image icons point at the widget-theme
// icon route, since a filesystem path does not resolve in a browser.
func browserIconHTML(store *widgetthemes.Store, themeName, canonicalName string) template.HTML {
	resolved, err := store.ResolveIcon(themeName, canonicalName)
	if err != nil {
		return browserIconFallback
	}
	switch resolved.Type {
	case widgetthemes.ThemeTypeFont:
		if resolved.Codepoint != "" {
			cp, err := strconv.ParseUint(resolved.Codepoint, 16, 32)
			if err == nil && cp <= unicode.MaxRune {
				return template.HTML(fmt.Sprintf(
					`<span style="font-family: %s; line-height: 1;" aria-hidden="true">%s</span>`,
					template.HTMLEscapeString(resolved.FontFamily),
					template.HTMLEscapeString(string(rune(cp)))))
			}
		}
		return template.HTML(fmt.Sprintf(`<i class="%s" aria-hidden="true"></i>`,
			template.HTMLEscapeString(resolved.CSSClass)))
	case widgetthemes.ThemeTypeImage:
		src := "/api/v0/widget-themes/" + url.PathEscape(themeName) + "/icons/" + url.PathEscape(canonicalName)
		return template.HTML(fmt.Sprintf(
			`<img src="%s" alt="%s" style="display: inline-block; width: 1em; height: 1em; vertical-align: middle;">`,
			template.HTMLEscapeString(src),
			template.HTMLEscapeString(canonicalName)))
	default:
		return browserIconFallback
	}
}
