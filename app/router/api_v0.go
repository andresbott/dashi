package router

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/market"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/swisstransport"
	"github.com/andresbott/dashi/internal/weather"
	"github.com/andresbott/dashi/internal/xkcd"
	"github.com/gorilla/mux"

	"github.com/andresbott/dashi/app/router/handlers"
)

// apiDeps holds shared dependencies for API route handlers.
type apiDeps struct {
	dashStore       *dashboard.Store
	weatherClient   *weather.Client
	marketClient    *market.Client
	xkcdClient      *xkcd.Client
	transportClient *swisstransport.Client
	themeStore      *themes.Store
	logger          *slog.Logger
}

// attachReadAPIs mounts all read-only (GET) API endpoints on the given router.
func attachReadAPIs(r *mux.Router, deps apiDeps) {
	// Health check
	r.Path("/health").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Dashboard routes (read)
	dh := handlers.NewDashboardHandler(deps.dashStore, deps.themeStore, deps.logger)
	r.Path("/dashboards").Methods(http.MethodGet).HandlerFunc(dh.List)
	r.Path("/dashboards/{id}").Methods(http.MethodGet).HandlerFunc(dh.Get)
	r.Path("/dashboards/{id}/download").Methods(http.MethodGet).HandlerFunc(dh.Download)
	r.Path("/dashboards/{id}/assets").Methods(http.MethodGet).HandlerFunc(dh.ListAssets)
	r.Path("/dashboards/{id}/assets/{path:.*}").Methods(http.MethodGet).HandlerFunc(dh.GetAsset)
	r.Path("/backgrounds").Methods(http.MethodGet).HandlerFunc(dh.ListBackgrounds)

	// Weather widget routes
	wh := handlers.NewWeatherHandler(deps.weatherClient, deps.logger)
	r.Path("/widgets/weather").Methods(http.MethodGet).HandlerFunc(wh.GetWeather)
	r.Path("/widgets/weather/geocode").Methods(http.MethodGet).HandlerFunc(wh.Geocode)

	// Theme routes
	th := handlers.NewThemeHandler(deps.themeStore, deps.logger)
	r.Path("/themes").Methods(http.MethodGet).HandlerFunc(th.List)
	r.Path("/themes/{name}/icons/{icon}").Methods(http.MethodGet).HandlerFunc(th.GetIcon)
	r.Path("/themes/{name}/fonts/{font}").Methods(http.MethodGet).HandlerFunc(th.GetFont)
	r.Path("/themes/{name}/backgrounds/{file}").Methods(http.MethodGet).HandlerFunc(th.GetBackground)

	// Market widget routes
	mh := handlers.NewMarketHandler(deps.marketClient, deps.logger)
	r.Path("/widgets/market").Methods(http.MethodGet).HandlerFunc(mh.GetMarketData)

	// XKCD widget routes
	xh := handlers.NewXkcdHandler(deps.xkcdClient, deps.logger)
	r.Path("/widgets/xkcd").Methods(http.MethodGet).HandlerFunc(xh.GetComic)

	// Transport widget routes
	trh := handlers.NewTransportHandler(deps.transportClient, deps.logger)
	r.Path("/widgets/transport/stationboard").Methods(http.MethodGet).HandlerFunc(trh.GetDepartures)
	r.Path("/widgets/transport/stations").Methods(http.MethodGet).HandlerFunc(trh.SearchStations)

	// Sysinfo widget routes
	sh := handlers.NewSysinfoHandler(deps.logger)
	r.Path("/widgets/sysinfo").Methods(http.MethodGet).HandlerFunc(sh.GetSysinfo)
}

// attachWriteAPIs mounts all write (POST/PUT/DELETE) API endpoints on the given router.
func attachWriteAPIs(r *mux.Router, deps apiDeps) {
	dh := handlers.NewDashboardHandler(deps.dashStore, deps.themeStore, deps.logger)

	r.Path("/dashboards").Methods(http.MethodPost).HandlerFunc(dh.Create)
	r.Path("/dashboards/upload").Methods(http.MethodPost).HandlerFunc(dh.Upload)
	r.Path("/dashboards/previews").Methods(http.MethodDelete).HandlerFunc(dh.DeletePreviews)
	r.Path("/dashboards/{id}").Methods(http.MethodPut).HandlerFunc(dh.Update)
	r.Path("/dashboards/{id}").Methods(http.MethodDelete).HandlerFunc(dh.Delete)
	r.Path("/dashboards/{id}/assets/{path:.*}").Methods(http.MethodPost).HandlerFunc(dh.UploadAsset)
	r.Path("/dashboards/{id}/assets/{path:.*}").Methods(http.MethodDelete).HandlerFunc(dh.DeleteAsset)

	// Dashboard auth routes (editor only)
	r.Path("/dashboards/{id}/auth").Methods(http.MethodGet).HandlerFunc(dh.GetAuth)
	r.Path("/dashboards/{id}/auth").Methods(http.MethodPut).HandlerFunc(dh.SetAuth)
	r.Path("/dashboards/{id}/auth").Methods(http.MethodDelete).HandlerFunc(dh.DeleteAuth)
}
