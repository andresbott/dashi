package router

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/andresbott/dashi/app/router/handlers"
	"github.com/andresbott/dashi/app/spa"
	"github.com/andresbott/dashi/internal/backgrounds"
	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/dashboard/browser"
	dashimage "github.com/andresbott/dashi/internal/dashboard/image"
	dashstatic "github.com/andresbott/dashi/internal/dashboard/static"
	"github.com/andresbott/dashi/internal/data/images"
	"github.com/andresbott/dashi/internal/data/notes"
	"github.com/andresbott/dashi/internal/providers/market"
	"github.com/andresbott/dashi/internal/providers/swisstransport"
	"github.com/andresbott/dashi/internal/providers/weather"
	"github.com/andresbott/dashi/internal/providers/xkcd"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/widgets"
	batterywidget "github.com/andresbott/dashi/internal/widgets/battery"
	bookmarkwidget "github.com/andresbott/dashi/internal/widgets/bookmark"
	clockwidget "github.com/andresbott/dashi/internal/widgets/clock"
	imagewidget "github.com/andresbott/dashi/internal/widgets/image"
	markdownwidget "github.com/andresbott/dashi/internal/widgets/markdown"
	marketwidget "github.com/andresbott/dashi/internal/widgets/market"
	pageindicatorwidget "github.com/andresbott/dashi/internal/widgets/pageindicator"
	searchwidget "github.com/andresbott/dashi/internal/widgets/search"
	stackwidget "github.com/andresbott/dashi/internal/widgets/stack"
	swisstransportwidget "github.com/andresbott/dashi/internal/widgets/swisstransport"
	sysinfowidget "github.com/andresbott/dashi/internal/widgets/sysinfo"
	weatherwidget "github.com/andresbott/dashi/internal/widgets/weather"
	xkcdwidget "github.com/andresbott/dashi/internal/widgets/xkcd"
	"github.com/go-bumbu/http/middleware"
	"github.com/gorilla/mux"
)

type Cfg struct {
	Ctx            context.Context
	Logger         *slog.Logger
	ProductionMode bool
	DataDir        string

	// Where the public viewer is reachable from a browser. Advertised to the
	// admin SPA through GET /api/v0/info: the SPA renders no dashboards, so it
	// links to the viewer server instead. ViewerPublicURL overrides the URL
	// derived from the request host (needed behind a reverse proxy).
	ViewerEnabled   bool
	ViewerPort      int
	ViewerPublicURL string
}

// ViewerHandler serves the read-only dashboard viewer.
type ViewerHandler struct {
	router         *mux.Router
	logger         *slog.Logger
	productionMode bool
}

func (h *ViewerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

// EditorHandler serves the full-CRUD dashboard editor.
type EditorHandler struct {
	router         *mux.Router
	logger         *slog.Logger
	productionMode bool
}

func (h *EditorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

// sharedDeps holds all shared clients, stores, renderers and middleware
// that are built once and reused by both viewer and editor handlers.
type sharedDeps struct {
	dashStore        *dashboard.Store
	weatherClient    *weather.Client
	marketClient     *market.Client
	xkcdClient       *xkcd.Client
	transportClient  *swisstransport.Client
	themeStore      *themes.Store
	notesStore      *notes.Store
	imagesStore     *images.Store
	sharedBgImages  *images.Store
	bgStore         *backgrounds.Store
	staticRenderer  *dashstatic.Renderer
	imageRenderer    *dashimage.Renderer
	browserRenderer  *browser.Renderer
	browserAssets    *browser.Assets
	staticMid        func(http.Handler) http.Handler
	promHisto        middleware.Histogram
	modules          []widgets.Module
	publicViewer     handlers.PublicViewer
}

func newSharedDeps(cfg Cfg) (*sharedDeps, error) {
	dashStore := dashboard.NewStore(filepath.Join(cfg.DataDir, "dashboards"))
	weatherClient := weather.NewClient(nil)
	marketClient := market.NewClient(nil)
	xkcdClient := xkcd.NewClient(filepath.Join(cfg.DataDir, "cache", "xkcd"))
	transportClient := swisstransport.NewClient(nil)
	themeStore := themes.NewStore(filepath.Join(cfg.DataDir, "themes"))
	notesStore, err := notes.NewStore(filepath.Join(cfg.DataDir, "data", "notes"))
	if err != nil {
		return nil, fmt.Errorf("create notes store: %w", err)
	}
	imagesStore, err := images.NewStore(filepath.Join(cfg.DataDir, "data", "images"))
	if err != nil {
		return nil, fmt.Errorf("create images store: %w", err)
	}
	sharedBgImages, err := images.NewStore(filepath.Join(cfg.DataDir, "data", "shared-background-images"))
	if err != nil {
		return nil, fmt.Errorf("create shared background images store: %w", err)
	}
	bgStore := backgrounds.NewStore(filepath.Join(cfg.DataDir, "data", "backgrounds"), sharedBgImages)

	// Static dashboard rendering
	registry := widgets.NewRegistry()

	modules := []widgets.Module{
		weatherwidget.NewModule(weatherClient, themeStore, cfg.Logger),
		weatherwidget.NewCompactModule(weatherClient, themeStore, cfg.Logger),
		bookmarkwidget.NewModule(),
		clockwidget.NewModule(),
		batterywidget.NewModule(),
		pageindicatorwidget.NewModule(),
		marketwidget.NewModule(marketClient, cfg.Logger),
		xkcdwidget.NewModule(xkcdClient, cfg.Logger),
		swisstransportwidget.NewModule(transportClient, cfg.Logger),
		sysinfowidget.NewModule(cfg.Logger),
		stackwidget.NewModule(registry),
		markdownwidget.NewModule(notesStore),
		imagewidget.NewModule(imagesStore),
		searchwidget.NewModule(),
	}

	warmupCtx := cfg.Ctx
	if warmupCtx == nil {
		warmupCtx = context.Background()
	}

	for _, m := range modules {
		registry.Register(m.Type(), m.Renderer())
		if br, ok := m.(widgets.BrowserRenderable); ok {
			registry.RegisterBrowser(m.Type(), br.RenderBrowser)
		}
		configs := widgets.CollectConfigs(dashStore, m.Type())
		go m.Warmup(warmupCtx, configs)
	}

	staticRenderer := dashstatic.NewRenderer(registry)
	browserAssets := browser.NewAssets(modules)
	browserRenderer := browser.NewRenderer(registry, browserAssets.JSTypes())
	imageRenderer := dashimage.NewRenderer()

	for _, themeInfo := range themeStore.List() {
		if themeInfo.HasIcons && themeInfo.IconType == themes.ThemeTypeFont {
			fontData, err := themeStore.GetFontData(themeInfo.Name)
			if err != nil {
				continue
			}
			imageRenderer.RegisterFont("icon-font-"+themeInfo.Name, fontData)
		}
	}
	// Register display fonts for image rendering
	for _, themeInfo := range themeStore.List() {
		for _, font := range themeInfo.Fonts {
			fontData, err := themeStore.GetDisplayFontData(themeInfo.Name, font.Name)
			if err != nil {
				continue
			}
			imageRenderer.RegisterFont(font.Name, fontData)
		}
	}

	staticMid := NewDashboardMiddleware(dashStore, browserRenderer, staticRenderer, imageRenderer, themeStore, sharedBgImages, bgStore)
	promHisto := middleware.NewPromHistogram("", nil, nil)

	return &sharedDeps{
		dashStore:       dashStore,
		weatherClient:   weatherClient,
		marketClient:    marketClient,
		xkcdClient:      xkcdClient,
		transportClient: transportClient,
		themeStore:      themeStore,
		notesStore:      notesStore,
		imagesStore:     imagesStore,
		sharedBgImages:  sharedBgImages,
		bgStore:         bgStore,
		staticRenderer:  staticRenderer,
		imageRenderer:   imageRenderer,
		browserRenderer: browserRenderer,
		browserAssets:   browserAssets,
		staticMid:       staticMid,
		promHisto:       promHisto,
		modules:         modules,
		publicViewer: handlers.PublicViewer{
			Enabled: cfg.ViewerEnabled,
			Port:    cfg.ViewerPort,
			BaseURL: cfg.ViewerPublicURL,
		},
	}, nil
}

func newAPIDeps(deps *sharedDeps, logger *slog.Logger) apiDeps {
	return apiDeps{
		dashStore:      deps.dashStore,
		themeStore:     deps.themeStore,
		notesStore:     deps.notesStore,
		imagesStore:    deps.imagesStore,
		sharedBgImages: deps.sharedBgImages,
		bgStore:        deps.bgStore,
		logger:         logger,
		modules:        deps.modules,
		publicViewer:   deps.publicViewer,
	}
}

// NewViewerFromDeps creates a viewer handler using pre-built shared deps.
func NewViewerFromDeps(cfg Cfg, deps *sharedDeps) (*ViewerHandler, error) {
	r := mux.NewRouter()
	h := &ViewerHandler{
		router:         r,
		logger:         cfg.Logger,
		productionMode: cfg.ProductionMode,
	}

	prodMid := middleware.New(middleware.Cfg{
		JsonErrors:  true,
		GenericErrs: false,
		Logger:      cfg.Logger,
		PromHisto:   deps.promHisto,
	})
	r.Use(prodMid.Middleware)

	// Per-dashboard basic auth (checks /{id} and /api/v0/dashboards/{id} paths)
	authMid := NewDashboardAuthMiddleware(deps.dashStore)
	r.Use(authMid)

	// API v0 routes (read-only)
	ad := newAPIDeps(deps, cfg.Logger)
	attachReadAPIs(r.PathPrefix("/api/v0").Subrouter(), ad)

	// Browser-stack assets (CSS/JS/theme). Registered before the SPA and
	// dashboard-ID routes so the /_dashi prefix always wins.
	attachDashiAssets(r, deps.browserAssets, deps.themeStore)

	// Build the SPA handler once
	spaHandler, err := spa.App("/")
	if err != nil {
		return nil, err
	}

	// SPA static assets (JS, CSS, fonts, images)
	r.PathPrefix("/assets/").Methods(http.MethodGet).Handler(spaHandler)

	// Root "/" — resolve the default dashboard server-side and redirect.
	r.Path("/").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		list, err := deps.dashStore.List()
		if err != nil || len(list) == 0 {
			http.Error(w, "no dashboards configured", http.StatusNotFound)
			return
		}
		target := list[0]
		for _, meta := range list {
			if meta.Default {
				target = meta
				break
			}
		}
		http.Redirect(w, req, "/"+target.ID, http.StatusFound)
	})

	// "/:id" — single-segment paths only (no slashes), with static middleware + SPA
	spaSubrouter := r.PathPrefix("/").Subrouter()
	spaSubrouter.Use(deps.staticMid)
	spaSubrouter.Methods(http.MethodGet).MatcherFunc(viewerPathMatcher).Handler(spaHandler)

	return h, nil
}

// viewerPathMatcher matches only dashboard ID paths (single segment, not a known editor route).
func viewerPathMatcher(r *http.Request, rm *mux.RouteMatch) bool {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" || strings.Contains(path, "/") {
		return false
	}
	// Reject known editor/SPA-only routes
	switch path {
	case "dashboards", "docs":
		return false
	}
	return true
}

// NewEditorFromDeps creates an editor handler using pre-built shared deps.
func NewEditorFromDeps(cfg Cfg, deps *sharedDeps) (*EditorHandler, error) {
	r := mux.NewRouter()
	h := &EditorHandler{
		router:         r,
		logger:         cfg.Logger,
		productionMode: cfg.ProductionMode,
	}

	prodMid := middleware.New(middleware.Cfg{
		JsonErrors:  true,
		GenericErrs: false,
		Logger:      cfg.Logger,
		PromHisto:   deps.promHisto,
	})
	r.Use(prodMid.Middleware)

	// API v0 routes (read + write)
	ad := newAPIDeps(deps, cfg.Logger)
	apiRouter := r.PathPrefix("/api/v0").Subrouter()
	attachReadAPIs(apiRouter, ad)
	attachWriteAPIs(apiRouter, ad)

	// Browser-stack assets (CSS/JS/theme). Registered before the SPA and
	// dashboard-ID routes so the /_dashi prefix always wins.
	attachDashiAssets(r, deps.browserAssets, deps.themeStore)

	// Root "/" redirects to /admin
	r.Path("/").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin", http.StatusFound)
	})

	// Static dashboard middleware (image rendering) + full SPA on all paths
	spaHandler, err := spa.App("/")
	if err != nil {
		return nil, err
	}
	spaRouter := r.PathPrefix("/").Subrouter()
	spaRouter.Use(deps.staticMid)
	spaRouter.PathPrefix("/").Handler(spaHandler)

	return h, nil
}

// NewViewer creates a viewer handler (convenience constructor).
func NewViewer(cfg Cfg) (*ViewerHandler, error) {
	deps, err := newSharedDeps(cfg)
	if err != nil {
		return nil, err
	}
	return NewViewerFromDeps(cfg, deps)
}

// NewEditor creates an editor handler (convenience constructor).
func NewEditor(cfg Cfg) (*EditorHandler, error) {
	deps, err := newSharedDeps(cfg)
	if err != nil {
		return nil, err
	}
	return NewEditorFromDeps(cfg, deps)
}

// NewBoth creates both viewer and editor handlers sharing the same deps.
func NewBoth(cfg Cfg) (*ViewerHandler, *EditorHandler, error) {
	deps, err := newSharedDeps(cfg)
	if err != nil {
		return nil, nil, err
	}
	viewer, err := NewViewerFromDeps(cfg, deps)
	if err != nil {
		return nil, nil, err
	}
	editor, err := NewEditorFromDeps(cfg, deps)
	if err != nil {
		return nil, nil, err
	}
	return viewer, editor, nil
}
