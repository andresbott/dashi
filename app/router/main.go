package router

import (
	"context"
	_ "embed"
	"encoding/json"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/andresbott/dashi/app/spa"
	"github.com/andresbott/dashi/internal/dashboard"
	dashimage "github.com/andresbott/dashi/internal/dashboard/image"
	dashstatic "github.com/andresbott/dashi/internal/dashboard/static"
	"github.com/andresbott/dashi/internal/market"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/swisstransport"
	"github.com/andresbott/dashi/internal/weather"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/andresbott/dashi/internal/xkcd"
	batterywidget "github.com/andresbott/dashi/internal/widgets/battery"
	bookmarkwidget "github.com/andresbott/dashi/internal/widgets/bookmark"
	clockwidget "github.com/andresbott/dashi/internal/widgets/clock"
	imagewidget "github.com/andresbott/dashi/internal/widgets/image"
	markdownwidget "github.com/andresbott/dashi/internal/widgets/markdown"
	marketwidget "github.com/andresbott/dashi/internal/widgets/market"
	pageindicatorwidget "github.com/andresbott/dashi/internal/widgets/pageindicator"
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
	dashStore       *dashboard.Store
	weatherClient   *weather.Client
	marketClient    *market.Client
	xkcdClient      *xkcd.Client
	transportClient *swisstransport.Client
	themeStore      *themes.Store
	staticRenderer  *dashstatic.Renderer
	imageRenderer   *dashimage.Renderer
	staticMid       func(http.Handler) http.Handler
	promHisto       middleware.Histogram
}

func newSharedDeps(cfg Cfg) (*sharedDeps, error) {
	dashStore := dashboard.NewStore(filepath.Join(cfg.DataDir, "dashboards"))
	weatherClient := weather.NewClient(nil)
	marketClient := market.NewClient(nil)
	xkcdClient := xkcd.NewClient(filepath.Join(cfg.DataDir, "cache", "xkcd"))
	transportClient := swisstransport.NewClient(nil)
	themeStore := themes.NewStore(filepath.Join(cfg.DataDir, "themes"))

	// Pre-fetch weather and market data
	warmupCtx := cfg.Ctx
	if warmupCtx == nil {
		warmupCtx = context.Background()
	}
	go warmupWeather(warmupCtx, dashStore, weatherClient, cfg.Logger)
	go warmupMarket(warmupCtx, dashStore, marketClient, cfg.Logger)

	// Static dashboard rendering
	registry := widgets.NewRegistry()
	registry.Register("weather", weatherwidget.NewStaticRenderer(weatherClient, themeStore))
	registry.Register("weather-compact", weatherwidget.NewStaticCompactRenderer(weatherClient, themeStore))
	registry.Register("bookmark", bookmarkwidget.NewStaticRenderer())
	registry.Register("clock", clockwidget.NewStaticRenderer(nil))
	registry.Register("battery", batterywidget.NewStaticRenderer())
	registry.Register("page-indicator", pageindicatorwidget.NewStaticRenderer())
	registry.Register("market", marketwidget.NewStaticRenderer(marketClient))
	registry.Register("xkcd", xkcdwidget.NewStaticRenderer(xkcdClient))
	registry.Register("transport", swisstransportwidget.NewStaticRenderer(transportClient))
	registry.Register("sysinfo", sysinfowidget.NewStaticRenderer())
	registry.Register("stack", stackwidget.NewStaticRenderer(registry))
	registry.Register("markdown", markdownwidget.NewStaticRenderer(dashStore))
	registry.Register("image", imagewidget.NewStaticRenderer(dashStore))
	staticRenderer := dashstatic.NewRenderer(registry)
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

	staticMid := NewStaticDashboardMiddleware(dashStore, staticRenderer, imageRenderer, themeStore)
	promHisto := middleware.NewPromHistogram("", nil, nil)

	return &sharedDeps{
		dashStore:       dashStore,
		weatherClient:   weatherClient,
		marketClient:    marketClient,
		xkcdClient:      xkcdClient,
		transportClient: transportClient,
		themeStore:      themeStore,
		staticRenderer:  staticRenderer,
		imageRenderer:   imageRenderer,
		staticMid:       staticMid,
		promHisto:       promHisto,
	}, nil
}

func newAPIDeps(deps *sharedDeps, logger *slog.Logger) apiDeps {
	return apiDeps{
		dashStore:       deps.dashStore,
		weatherClient:   deps.weatherClient,
		marketClient:    deps.marketClient,
		xkcdClient:      deps.xkcdClient,
		transportClient: deps.transportClient,
		themeStore:      deps.themeStore,
		logger:          logger,
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

	// Build the SPA handler once
	spaHandler, err := spa.App("/")
	if err != nil {
		return nil, err
	}

	// SPA static assets (JS, CSS, fonts, images)
	r.PathPrefix("/assets/").Methods(http.MethodGet).Handler(spaHandler)

	// Root "/" — serve SPA directly (Vue resolves default dashboard client-side)
	r.Path("/").Methods(http.MethodGet).Handler(spaHandler)

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

	// Root "/" redirects to /dashboards
	r.Path("/").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboards", http.StatusFound)
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

// warmupWeather scans all dashboards for weather widget configs and
// pre-fetches the weather data so the cache is warm on first request.
func warmupWeather(ctx context.Context, store *dashboard.Store, client *weather.Client, logger *slog.Logger) {
	list, err := store.List()
	if err != nil {
		logger.Warn("weather warmup: failed to list dashboards", slog.String("error", err.Error()))
		return
	}

	var locations [][2]float64

	for _, meta := range list {
		dash, err := store.Get(meta.ID)
		if err != nil {
			continue
		}
		for _, page := range dash.Pages {
			for _, row := range page.Rows {
				for _, w := range row.Widgets {
					if w.Type != "weather" && w.Type != "weather-compact" {
						continue
					}
					var cfg struct {
						Latitude  float64 `json:"latitude"`
						Longitude float64 `json:"longitude"`
					}
					if err := json.Unmarshal(w.Config, &cfg); err != nil {
						continue
					}
					if cfg.Latitude != 0 || cfg.Longitude != 0 {
						locations = append(locations, [2]float64{cfg.Latitude, cfg.Longitude})
					}
				}
			}
		}
	}

	if len(locations) > 0 {
		logger.Info("weather warmup: pre-fetching data", slog.Int("locations", len(locations)))
		client.WarmupLocations(ctx, locations)
		logger.Info("weather warmup: done")
	}
}

func warmupMarket(ctx context.Context, store *dashboard.Store, client *market.Client, logger *slog.Logger) {
	list, err := store.List()
	if err != nil {
		logger.Warn("market warmup: failed to list dashboards", slog.String("error", err.Error()))
		return
	}

	var targets []struct{ Symbol, Range string }

	for _, meta := range list {
		dash, err := store.Get(meta.ID)
		if err != nil {
			continue
		}
		for _, page := range dash.Pages {
			for _, row := range page.Rows {
				for _, w := range row.Widgets {
					if w.Type != "market" {
						continue
					}
					var cfg struct {
						Symbol string `json:"symbol"`
						Range  string `json:"range"`
					}
					if err := json.Unmarshal(w.Config, &cfg); err != nil || cfg.Symbol == "" {
						continue
					}
					if cfg.Range == "" {
						cfg.Range = "1mo"
					}
					targets = append(targets, struct{ Symbol, Range string }{cfg.Symbol, cfg.Range})
				}
			}
		}
	}

	if len(targets) > 0 {
		logger.Info("market warmup: pre-fetching data", slog.Int("symbols", len(targets)))
		client.WarmupSymbols(ctx, targets)
		logger.Info("market warmup: done")
	}
}
