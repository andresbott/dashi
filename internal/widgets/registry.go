package widgets

import (
	"encoding/json"
	"html/template"
)

// RenderContext provides dashboard-level settings to widget renderers.
type RenderContext struct {
	DashboardID string            // dashboard identifier for asset resolution
	Theme       string            // dashboard theme name
	QueryParams map[string]string // URL query parameters from the HTTP request
	PageIndex   int               // zero-based index of the current page
	TotalPages  int               // total number of pages in the dashboard
}

// StaticRenderer renders a widget's HTML fragment from its JSON config.
type StaticRenderer func(config json.RawMessage, ctx RenderContext) (template.HTML, error)

// Registry maps widget type strings to their static renderers.
type Registry struct {
	renderers map[string]StaticRenderer
}

// NewRegistry creates an empty widget registry.
func NewRegistry() *Registry {
	return &Registry{renderers: make(map[string]StaticRenderer)}
}

// Register adds a static renderer for the given widget type.
func (r *Registry) Register(widgetType string, renderer StaticRenderer) {
	r.renderers[widgetType] = renderer
}

// Render calls the registered renderer for widgetType.
// If the type is not registered, it returns an empty placeholder div.
func (r *Registry) Render(widgetType string, config json.RawMessage, ctx RenderContext) (template.HTML, error) {
	renderer, ok := r.renderers[widgetType]
	if !ok {
		return template.HTML(`<div class="widget-placeholder">&nbsp;</div>`), nil
	}
	return renderer(config, ctx)
}
