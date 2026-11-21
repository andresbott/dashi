package bookmark

import (
	"encoding/json"
	"fmt"

	"github.com/andresbott/dashi/internal/widgets"
)

// Module implements widgets.Module for the bookmark widget.
type Module struct {
	widgets.NoopModule
}

// NewModule constructs a bookmark Module.
func NewModule() *Module { return &Module{} }

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
