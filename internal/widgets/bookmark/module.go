package bookmark

import (
	"encoding/json"
	"fmt"

	"github.com/andresbott/dashi/internal/widgets"
	"github.com/andresbott/dashi/internal/widgetthemes"
)

// Module implements widgets.Module for the bookmark widget.
type Module struct {
	widgets.NoopModule
	themes *widgetthemes.Store // resolves config.icon for the browser stack; nil renders no icon
}

// NewModule constructs a bookmark Module. The widget-theme store resolves
// the configured icon in the browser renderer; the image renderer does not
// use it.
func NewModule(themes *widgetthemes.Store) *Module { return &Module{themes: themes} }

// Type returns the widget type string.
func (m *Module) Type() string { return "bookmark" }

// Renderer returns the bookmark's static renderer.
func (m *Module) Renderer() widgets.StaticRenderer {
	return NewStaticRenderer()
}

// ValidateConfig implements widgets.ConfigValidator: it rejects a config
// that does not decode or whose url fails validateURL.
func (m *Module) ValidateConfig(config json.RawMessage) error {
	var cfg bookmarkConfig
	if len(config) > 0 {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return fmt.Errorf("bookmark config: %w", err)
		}
	}
	if err := validateURL(cfg.URL); err != nil {
		return fmt.Errorf("bookmark %w", err)
	}
	return nil
}
