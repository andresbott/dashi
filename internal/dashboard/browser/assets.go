package browser

import (
	"bytes"
	_ "embed"
	"sort"

	"github.com/andresbott/dashi/internal/widgets"
)

//go:embed assets/viewer.css
var viewerCSS []byte

//go:embed assets/dashi.js
var dashiJS []byte

// Assets holds the static browser-stack assets: the generated Tailwind
// stylesheet, the shared JS helper, and every widget's CSS and JS
// collected from the modules that ship them.
type Assets struct {
	widgetsCSS []byte
	widgetJS   map[string][]byte
}

// NewAssets collects browser assets from every module implementing
// widgets.BrowserAssets. Widget CSS is concatenated in type order so the
// output is byte-identical across restarts, which keeps ETags stable.
func NewAssets(modules []widgets.Module) *Assets {
	a := &Assets{widgetJS: make(map[string][]byte)}

	type entry struct {
		typ string
		css []byte
	}
	var entries []entry

	for _, m := range modules {
		provider, ok := m.(widgets.BrowserAssets)
		if !ok {
			continue
		}
		if css := provider.CSS(); len(css) > 0 {
			entries = append(entries, entry{typ: m.Type(), css: css})
		}
		if js := provider.JS(); len(js) > 0 {
			a.widgetJS[m.Type()] = js
		}
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].typ < entries[j].typ })

	var buf bytes.Buffer
	for _, e := range entries {
		buf.WriteString("/* " + e.typ + " */\n")
		buf.Write(e.css)
		buf.WriteString("\n")
	}
	a.widgetsCSS = buf.Bytes()

	return a
}

// ViewerCSS returns the generated Tailwind stylesheet.
func (a *Assets) ViewerCSS() []byte { return viewerCSS }

// DashiJS returns the shared browser runtime.
func (a *Assets) DashiJS() []byte { return dashiJS }

// WidgetsCSS returns every widget's stylesheet, concatenated.
func (a *Assets) WidgetsCSS() []byte { return a.widgetsCSS }

// WidgetJS returns a widget type's ES module.
func (a *Assets) WidgetJS(widgetType string) ([]byte, bool) {
	js, ok := a.widgetJS[widgetType]
	return js, ok
}

// JSTypes returns the set of widget types that ship an ES module, for
// the page shell's script tags.
func (a *Assets) JSTypes() map[string]bool {
	out := make(map[string]bool, len(a.widgetJS))
	for typ := range a.widgetJS {
		out[typ] = true
	}
	return out
}
