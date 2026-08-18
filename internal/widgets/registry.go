package widgets

import (
	"context"
	"encoding/json"
	"html/template"

	"github.com/andresbott/dashi/internal/themes"
	"github.com/gorilla/mux"
)

// RenderContext provides dashboard-level settings to widget renderers.
type RenderContext struct {
	DashboardID string            // dashboard identifier for asset resolution
	Theme       string            // dashboard theme name
	ColorMode   string            // "light" or "dark"
	Palette     themes.Palette    // resolved theme colours; image-stack templates
	                              // use these directly because litehtml cannot
	                              // read CSS custom properties
	QueryParams map[string]string // URL query parameters from the HTTP request
	PageIndex   int               // zero-based index of the current page
	TotalPages  int               // total number of pages in the dashboard
}

// EffectivePalette returns the context's palette, or the built-in palette
// for the context's colour mode when no palette was supplied.
//
// Both production callers (internal/dashboard/static and
// internal/dashboard/browser) always populate Palette, so the fallback only
// fires for hand-built contexts — unit tests today, and any future caller
// that assembles a RenderContext directly. Deriving the fallback from
// ColorMode rather than hardcoding the light palette is what keeps such a
// caller from rendering dark-mode output with light greys.
func (c RenderContext) EffectivePalette() themes.Palette {
	if c.Palette.Muted != "" {
		return c.Palette
	}
	return themes.DefaultPalette(c.ColorMode)
}

// StaticRenderer renders a widget's HTML fragment from its JSON config.
type StaticRenderer func(config json.RawMessage, ctx RenderContext) (template.HTML, error)

// Registry maps widget type strings to their renderers. It holds one map
// per output stack: image (litehtml/e-ink) and browser.
type Registry struct {
	renderers map[string]StaticRenderer
	browser   map[string]StaticRenderer
}

// NewRegistry creates an empty widget registry.
func NewRegistry() *Registry {
	return &Registry{
		renderers: make(map[string]StaticRenderer),
		browser:   make(map[string]StaticRenderer),
	}
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

// RegisterBrowser adds a browser-stack renderer for the given widget type.
func (r *Registry) RegisterBrowser(widgetType string, renderer StaticRenderer) {
	r.browser[widgetType] = renderer
}

// RenderBrowser renders a widget for the browser stack. Widgets that have
// not been ported to the browser stack yet fall back to their image
// renderer, so the browser view is never blank while the port is in
// progress.
func (r *Registry) RenderBrowser(widgetType string, config json.RawMessage, ctx RenderContext) (template.HTML, error) {
	if renderer, ok := r.browser[widgetType]; ok {
		return renderer(config, ctx)
	}
	return r.Render(widgetType, config, ctx)
}

// Module is the self-contained widget contract. A widget package
// implements this interface and the app/router layer iterates over
// a []Module to register renderers, routes, and warmup.
//
// RegisterRoutes and Warmup are optional: widgets with no interactive
// API or no warmup can embed NoopModule to inherit no-op defaults.
type Module interface {
	// Type returns the widget type string as stored in dashboard JSON.
	Type() string

	// Renderer returns the StaticRenderer used for image-mode rendering.
	// Widgets with no static rendering may return a renderer that emits
	// a placeholder div.
	Renderer() StaticRenderer

	// RegisterRoutes mounts the widget's interactive HTTP endpoints on r.
	// Called once at startup. Must not fail.
	RegisterRoutes(r *mux.Router)

	// Warmup pre-fetches data for all instances of this widget across
	// all dashboards. Runs in a goroutine at startup; errors are
	// non-fatal and should be logged by the widget.
	Warmup(ctx context.Context, configs []json.RawMessage)
}

// NoopModule provides no-op implementations of the optional Module
// methods (RegisterRoutes, Warmup). Embed it in a Module type to
// inherit the defaults.
type NoopModule struct{}

// RegisterRoutes is a no-op.
func (NoopModule) RegisterRoutes(r *mux.Router) {}

// Warmup is a no-op.
func (NoopModule) Warmup(ctx context.Context, configs []json.RawMessage) {}

// BrowserRenderable is implemented by widget modules that have a
// browser-stack renderer. It is deliberately an optional interface, not
// a method on Module: unported modules keep compiling untouched and fall
// back to their image renderer via Registry.RenderBrowser.
type BrowserRenderable interface {
	RenderBrowser(config json.RawMessage, ctx RenderContext) (template.HTML, error)
}

// BrowserAssets is implemented by widget modules that ship browser-stack
// assets. Either method may return nil — a widget with no behaviour
// ships no JS, and a widget styled purely with Tailwind utilities ships
// no CSS.
type BrowserAssets interface {
	// CSS returns the widget's structural stylesheet. All selectors must
	// descend from .dashi-{type}.
	CSS() []byte
	// JS returns the widget's ES module, served at /_dashi/widgets/{type}.js.
	JS() []byte
}
