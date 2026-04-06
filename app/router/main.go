package router

import (
	"context"
	_ "embed"
	"encoding/json"
	"log/slog"
	"net/http"
	"path/filepath"

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
	ReadOnly       bool
}

// MainAppHandler is the entrypoint http handler for the whole application
type MainAppHandler struct {
	router         *mux.Router
	logger         *slog.Logger
	productionMode bool
}

func (h *MainAppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func New(cfg Cfg) (*MainAppHandler, error) {
	r := mux.NewRouter()
	app := MainAppHandler{
		router:         r,
		logger:         cfg.Logger,
		productionMode: cfg.ProductionMode,
	}

	prodMid := middleware.New(middleware.Cfg{
		JsonErrors:  true,
		GenericErrs: false,
		Logger:      cfg.Logger,
		PromHisto:   middleware.NewPromHistogram("", nil, nil),
	})
	r.Use(prodMid.Middleware)

	// TODO: attach auth routes
	// app.attachUserAuth(app.router.PathPrefix("/auth").Subrouter())

	// API v0 routes
	dashStore := dashboard.NewStore(filepath.Join(cfg.DataDir, "dashboards"))
	weatherClient := weather.NewClient(nil)
	marketClient := market.NewClient(nil)
	xkcdClient := xkcd.NewClient(filepath.Join(cfg.DataDir, "cache", "xkcd"))
	transportClient := swisstransport.NewClient(nil)
	themeStore := themes.NewStore(filepath.Join(cfg.DataDir, "themes"))
	if err := app.attachApiV0(app.router.PathPrefix("/api/v0").Subrouter(), dashStore, weatherClient, marketClient, xkcdClient, transportClient, themeStore, cfg.ReadOnly); err != nil {
		return nil, err
	}

	// Pre-fetch weather data for all configured locations
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

	// SPA serving — with static dashboard middleware applied before it
	spaRouter := app.router.PathPrefix("/").Subrouter()
	spaRouter.Use(staticMid)
	if err := app.attachSpa(spaRouter, "/"); err != nil {
		return nil, err
	}

	return &app, nil
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

func (h *MainAppHandler) attachSpa(r *mux.Router, path string) error {
	spaHandler, err := spa.App(path)
	if err != nil {
		return err
	}
	r.Methods(http.MethodGet).PathPrefix(path).Handler(spaHandler)
	return nil
}
