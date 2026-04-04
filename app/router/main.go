package router

import (
	_ "embed"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/andresbott/dashi/app/spa"
	"github.com/andresbott/dashi/internal/dashboard"
	dashimage "github.com/andresbott/dashi/internal/dashboard/image"
	dashstatic "github.com/andresbott/dashi/internal/dashboard/static"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/weather"
	"github.com/andresbott/dashi/internal/widgets"
	batterywidget "github.com/andresbott/dashi/internal/widgets/battery"
	bookmarkwidget "github.com/andresbott/dashi/internal/widgets/bookmark"
	clockwidget "github.com/andresbott/dashi/internal/widgets/clock"
	weatherwidget "github.com/andresbott/dashi/internal/widgets/weather"
	"github.com/go-bumbu/http/middleware"
	"github.com/gorilla/mux"
)

type Cfg struct {
	Logger         *slog.Logger
	ProductionMode bool
	DataDir        string
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
	themeStore := themes.NewStore(filepath.Join(cfg.DataDir, "themes"))
	if err := app.attachApiV0(app.router.PathPrefix("/api/v0").Subrouter(), dashStore, weatherClient, themeStore); err != nil {
		return nil, err
	}

	// Static dashboard rendering
	registry := widgets.NewRegistry()
	registry.Register("weather", weatherwidget.NewStaticRenderer(weatherClient, themeStore))
	registry.Register("weather-compact", weatherwidget.NewStaticCompactRenderer(weatherClient, themeStore))
	registry.Register("bookmark", bookmarkwidget.NewStaticRenderer())
	registry.Register("clock", clockwidget.NewStaticRenderer(nil))
	registry.Register("battery", batterywidget.NewStaticRenderer())
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

func (h *MainAppHandler) attachSpa(r *mux.Router, path string) error {
	spaHandler, err := spa.App(path)
	if err != nil {
		return err
	}
	r.Methods(http.MethodGet).PathPrefix(path).Handler(spaHandler)
	return nil
}
