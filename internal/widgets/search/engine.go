package search

import (
	"encoding/json"
	"fmt"

	"github.com/andresbott/dashi/internal/widgetthemes"
)

type searchConfig struct {
	Engine      string `json:"engine"`
	Placeholder string `json:"placeholder"`
}

// engine is one allowlisted search engine. Both renders read the label and
// the URL from the same entry, so they cannot drift apart.
type engine struct {
	Label string
	// Action is the GET form target; the query travels in parameter q.
	Action string
}

const (
	defaultEngine      = "google"
	defaultPlaceholder = "Search..."
)

// engines is the allowlist; it mirrors searchEngineUrls in
// webui/src/widgets/search/types.ts (both checked against
// testdata/engines.json). Config can only pick a key, never supply a URL.
var engines = map[string]engine{
	"google":     {Label: "Google", Action: "https://www.google.com/search"},
	"duckduckgo": {Label: "DuckDuckGo", Action: "https://duckduckgo.com/"},
	"bing":       {Label: "Bing", Action: "https://www.bing.com/search"},
}

// Tabler code points, painted with the default widget theme's icon font
// (a pack's own icon font would map them to unrelated glyphs).
const (
	searchGlyph = "" // ti-search
	arrowGlyph  = "" // ti-arrow-right
)

var iconFontFamily = widgetthemes.IconFontFamily(widgetthemes.DefaultName)

// view is the resolved config both templates render.
type view struct {
	Action      string
	Label       string
	Placeholder string
}

// resolve parses the widget config and applies the defaults: an unknown or
// empty engine falls back to google, so a form never targets an arbitrary
// URL, and an empty placeholder becomes "Search...".
func resolve(config json.RawMessage) (view, error) {
	var cfg searchConfig
	if len(config) > 0 {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return view{}, fmt.Errorf("search config: %w", err)
		}
	}
	e, ok := engines[cfg.Engine]
	if !ok {
		e = engines[defaultEngine]
	}
	placeholder := cfg.Placeholder
	if placeholder == "" {
		placeholder = defaultPlaceholder
	}
	return view{Action: e.Action, Label: e.Label, Placeholder: placeholder}, nil
}
