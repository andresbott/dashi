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
	DashboardID string // dashboard identifier for asset resolution
	WidgetTheme string // widget theme name: fonts and icons
	ColorMode   string // "light" or "dark"
	// Palette holds the resolved theme colours. Image-stack templates use
	// these directly because litehtml cannot read CSS custom properties;
	// the image stack passes a flattened (alpha-free) palette.
	Palette     themes.Palette
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
	if c.Palette.TextMuted != "" {
		return c.Palette
	}
	return themes.DefaultPalette(c.ColorMode)
}

// StaticRenderer renders a widget's HTML fragment from its JSON config.
type StaticRenderer func(config json.RawMessage, ctx RenderContext) (template.HTML, error)

// TypedRenderer renders a widget of any type, e.g. Registry.Render or
// Registry.RenderBrowser. Containers receive one to render their children.
type TypedRenderer func(widgetType string, config json.RawMessage, ctx RenderContext) (template.HTML, error)

// Child is one widget nested in a container widget.
type Child struct {
	Type   string
	Config json.RawMessage
}

// ChildrenFunc lists a container widget's direct children from its config.
// A config that does not decode yields none.
type ChildrenFunc func(config json.RawMessage) []Child

// ConfigValidateFunc checks a widget's config JSON before it is persisted.
type ConfigValidateFunc func(config json.RawMessage) error

// Registry maps widget type strings to their renderers. It holds one map
// per output stack: image (litehtml/e-ink) and browser, plus the optional
// per-type config validators.
type Registry struct {
	renderers  map[string]StaticRenderer
	browser    map[string]StaticRenderer
	validators map[string]ConfigValidateFunc
	children   map[string]ChildrenFunc
}

// NewRegistry creates an empty widget registry.
func NewRegistry() *Registry {
	return &Registry{
		renderers:  make(map[string]StaticRenderer),
		browser:    make(map[string]StaticRenderer),
		validators: make(map[string]ConfigValidateFunc),
		children:   make(map[string]ChildrenFunc),
	}
}

// RegisterChildren marks widgetType as a container whose direct children
// are listed by fn.
func (r *Registry) RegisterChildren(widgetType string, fn ChildrenFunc) {
	r.children[widgetType] = fn
}

// TypesIn returns widgetType plus the types of every widget nested in it,
// at any depth. The browser page uses it to load the scripts of widgets
// that only appear inside a container. Duplicates are allowed.
func (r *Registry) TypesIn(widgetType string, config json.RawMessage) []string {
	types := []string{widgetType}
	fn, ok := r.children[widgetType]
	if !ok {
		return types
	}
	for _, child := range fn(config) {
		types = append(types, r.TypesIn(child.Type, child.Config)...)
	}
	return types
}

// RegisterValidator adds a config validator for the given widget type.
func (r *Registry) RegisterValidator(widgetType string, validate ConfigValidateFunc) {
	r.validators[widgetType] = validate
}

// ValidateConfig runs the validator registered for widgetType. Types
// without a validator (including unknown types) accept any config: widget
// config stays opaque unless the widget itself declares a rule.
func (r *Registry) ValidateConfig(widgetType string, config json.RawMessage) error {
	validate, ok := r.validators[widgetType]
	if !ok {
		return nil
	}
	return validate(config)
}

// Register adds a static renderer for the given widget type.
func (r *Registry) Register(widgetType string, renderer StaticRenderer) {
	r.renderers[widgetType] = renderer
}

// Render calls the registered renderer for widgetType.
// If the type is not registered, it renders nothing (empty output).
func (r *Registry) Render(widgetType string, config json.RawMessage, ctx RenderContext) (template.HTML, error) {
	renderer, ok := r.renderers[widgetType]
	if !ok {
		return template.HTML(""), nil
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

// ConfigValidator is implemented by widget modules that reject unsafe or
// malformed config before it is saved. It is optional like
// BrowserRenderable: the router registers it via Registry.RegisterValidator
// and the dashboard store calls Registry.ValidateConfig for every widget on
// create, update and zip import, answering 400 when it fails. Validation
// is a gate on writes only — renderers must still treat stored config as
// untrusted, because dashboards saved before a rule existed (or edited by
// hand on disk) never pass through it.
type ConfigValidator interface {
	ValidateConfig(config json.RawMessage) error
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

// Container is implemented by widget modules that nest other widgets
// (stack). It is optional like BrowserRenderable; the router registers it
// via Registry.RegisterChildren so Registry.TypesIn can see nested types.
type Container interface {
	Children(config json.RawMessage) []Child
}
