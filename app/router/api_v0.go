package router

import (
	"encoding/json"
	"net/http"

	"github.com/andresbott/dashi/app/router/handlers"
	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/weather"
	"github.com/gorilla/mux"
)

func (h *MainAppHandler) attachApiV0(r *mux.Router, dashStore *dashboard.Store, weatherClient *weather.Client, themeStore *themes.Store) error {

	// Health check
	r.Path("/health").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Dashboard routes
	dh := handlers.NewDashboardHandler(dashStore, h.logger)
	r.Path("/dashboards").Methods(http.MethodGet).HandlerFunc(dh.List)
	r.Path("/dashboards").Methods(http.MethodPost).HandlerFunc(dh.Create)
	r.Path("/dashboards/previews").Methods(http.MethodDelete).HandlerFunc(dh.DeletePreviews)
	r.Path("/dashboards/{id}").Methods(http.MethodGet).HandlerFunc(dh.Get)
	r.Path("/dashboards/{id}").Methods(http.MethodPut).HandlerFunc(dh.Update)
	r.Path("/dashboards/{id}").Methods(http.MethodDelete).HandlerFunc(dh.Delete)

	// Dashboard asset routes
	r.Path("/dashboards/{id}/assets").Methods(http.MethodGet).HandlerFunc(dh.ListAssets)
	r.Path("/dashboards/{id}/assets/{path:.*}").Methods(http.MethodGet).HandlerFunc(dh.GetAsset)
	r.Path("/dashboards/{id}/assets/{path:.*}").Methods(http.MethodPost).HandlerFunc(dh.UploadAsset)
	r.Path("/dashboards/{id}/assets/{path:.*}").Methods(http.MethodDelete).HandlerFunc(dh.DeleteAsset)

	// Weather widget routes
	wh := handlers.NewWeatherHandler(weatherClient)
	r.Path("/widgets/weather").Methods(http.MethodGet).HandlerFunc(wh.GetWeather)
	r.Path("/widgets/weather/geocode").Methods(http.MethodGet).HandlerFunc(wh.Geocode)

	// Theme routes
	th := handlers.NewThemeHandler(themeStore)
	r.Path("/themes").Methods(http.MethodGet).HandlerFunc(th.List)
	r.Path("/themes/{name}/icons/{icon}").Methods(http.MethodGet).HandlerFunc(th.GetIcon)

	return nil
}
