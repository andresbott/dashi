package stack

import (
	"encoding/json"
	"fmt"

	"github.com/andresbott/dashi/internal/widgets"
)

// Module implements widgets.Module for the stack widget. Stack renders
// child widgets recursively, so it needs a reference to the registry.
type Module struct {
	widgets.NoopModule
	registry *widgets.Registry
}

// NewModule constructs a stack Module. The registry is captured by
// reference; it may be populated after construction because the
// renderer is only invoked at request time.
func NewModule(registry *widgets.Registry) *Module {
	return &Module{registry: registry}
}

// Type returns the widget type string.
func (m *Module) Type() string { return "stack" }

// Renderer returns the stack's static renderer.
func (m *Module) Renderer() widgets.StaticRenderer {
	return NewStaticRenderer(m.registry)
}

// ValidateConfig implements widgets.ConfigValidator. A stack has no rule of
// its own; it runs each child's validator through the registry, so a child
// (a bookmark, say) cannot slip past save-time validation by being nested.
func (m *Module) ValidateConfig(config json.RawMessage) error {
	var cfg stackConfig
	if len(config) > 0 {
		if err := json.Unmarshal(config, &cfg); err != nil {
			return fmt.Errorf("stack config: %w", err)
		}
	}
	for _, child := range cfg.Widgets {
		if err := m.registry.ValidateConfig(child.Type, child.Config); err != nil {
			return fmt.Errorf("stack child %q (%s): %w", child.ID, child.Type, err)
		}
	}
	return nil
}
