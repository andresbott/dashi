package router

import (
	"encoding/json"
	"net/http"

	"github.com/andresbott/dashi/app/router/handlers"
	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/market"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/swisstransport"
	"github.com/andresbott/dashi/internal/weather"
	"github.com/andresbott/dashi/internal/xkcd"
	"github.com/gorilla/mux"
)

func (h *MainAppHandler) attachApiV0(r *mux.Router, dashStore *dashboard.Store, weatherClient *weather.Client, marketClient *market.Client, xkcdClient *xkcd.Client, transportClient *swisstransport.Client, themeStore *themes.Store, readOnly bool) error {

	// Health check
	r.Path("/health").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Settings endpoint
	r.Path("/settings").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"readOnly": readOnly})
	})

	// Read-only guard for write operations
	readOnlyGuard := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if readOnly {
				handlers.ErrorJSON(w, "read-only mode", http.StatusForbidden)
				return
			}
			next(w, r)
		}
	}

	// Dashboard routes
	dh := handlers.NewDashboardHandler(dashStore, themeStore, h.logger)
	r.Path("/dashboards").Methods(http.MethodGet).HandlerFunc(dh.List)
	r.Path("/dashboards").Methods(http.MethodPost).HandlerFunc(readOnlyGuard(dh.Create))
	r.Path("/dashboards/previews").Methods(http.MethodDelete).HandlerFunc(readOnlyGuard(dh.DeletePreviews))
	r.Path("/dashboards/{id}").Methods(http.MethodGet).HandlerFunc(dh.Get)
	r.Path("/dashboards/{id}").Methods(http.MethodPut).HandlerFunc(readOnlyGuard(dh.Update))
	r.Path("/dashboards/{id}").Methods(http.MethodDelete).HandlerFunc(readOnlyGuard(dh.Delete))

	// Dashboard asset routes
	r.Path("/dashboards/{id}/assets").Methods(http.MethodGet).HandlerFunc(dh.ListAssets)
	r.Path("/dashboards/{id}/assets/{path:.*}").Methods(http.MethodGet).HandlerFunc(dh.GetAsset)
	r.Path("/dashboards/{id}/assets/{path:.*}").Methods(http.MethodPost).HandlerFunc(readOnlyGuard(dh.UploadAsset))
	r.Path("/dashboards/{id}/assets/{path:.*}").Methods(http.MethodDelete).HandlerFunc(readOnlyGuard(dh.DeleteAsset))
	r.Path("/backgrounds").Methods(http.MethodGet).HandlerFunc(dh.ListBackgrounds)

	// Weather widget routes
	wh := handlers.NewWeatherHandler(weatherClient, h.logger)
	r.Path("/widgets/weather").Methods(http.MethodGet).HandlerFunc(wh.GetWeather)
	r.Path("/widgets/weather/geocode").Methods(http.MethodGet).HandlerFunc(wh.Geocode)

	// Theme routes
	th := handlers.NewThemeHandler(themeStore, h.logger)
	r.Path("/themes").Methods(http.MethodGet).HandlerFunc(th.List)
	r.Path("/themes/{name}/icons/{icon}").Methods(http.MethodGet).HandlerFunc(th.GetIcon)
	r.Path("/themes/{name}/fonts/{font}").Methods(http.MethodGet).HandlerFunc(th.GetFont)
	r.Path("/themes/{name}/backgrounds/{file}").Methods(http.MethodGet).HandlerFunc(th.GetBackground)

	// Market widget routes
	mh := handlers.NewMarketHandler(marketClient, h.logger)
	r.Path("/widgets/market").Methods(http.MethodGet).HandlerFunc(mh.GetMarketData)

	// XKCD widget routes
	xh := handlers.NewXkcdHandler(xkcdClient, h.logger)
	r.Path("/widgets/xkcd").Methods(http.MethodGet).HandlerFunc(xh.GetComic)

	// Transport widget routes
	trh := handlers.NewTransportHandler(transportClient, h.logger)
	r.Path("/widgets/transport/stationboard").Methods(http.MethodGet).HandlerFunc(trh.GetDepartures)
	r.Path("/widgets/transport/stations").Methods(http.MethodGet).HandlerFunc(trh.SearchStations)

	// Sysinfo widget routes
	sh := handlers.NewSysinfoHandler(h.logger)
	r.Path("/widgets/sysinfo").Methods(http.MethodGet).HandlerFunc(sh.GetSysinfo)

	return nil
}
