package router

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/gorilla/mux"

	"github.com/andresbott/dashi/app/router/handlers"
	"github.com/andresbott/dashi/internal/data/backgrounds"
	"github.com/andresbott/dashi/internal/data/images"
	"github.com/andresbott/dashi/internal/data/notes"
)

// apiDeps holds shared dependencies for API route handlers.
type apiDeps struct {
	dashStore        *dashboard.Store
	themeStore       *themes.Store
	notesStore       *notes.Store
	imagesStore      *images.Store
	backgroundsStore *backgrounds.Store
	logger           *slog.Logger
	modules          []widgets.Module
}

// attachReadAPIs mounts all read-only (GET) API endpoints on the given router.
func attachReadAPIs(r *mux.Router, deps apiDeps) {
	// Health check
	r.Path("/health").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Dashboard routes (read)
	dh := handlers.NewDashboardHandler(deps.dashStore, deps.themeStore, deps.backgroundsStore, deps.logger)
	r.Path("/dashboards").Methods(http.MethodGet).HandlerFunc(dh.List)
	r.Path("/dashboards/{id}").Methods(http.MethodGet).HandlerFunc(dh.Get)
	r.Path("/dashboards/{id}/download").Methods(http.MethodGet).HandlerFunc(dh.Download)
	r.Path("/dashboards/{id}/assets").Methods(http.MethodGet).HandlerFunc(dh.ListAssets)
	r.Path("/dashboards/{id}/assets/{path:.*}").Methods(http.MethodGet).HandlerFunc(dh.GetAsset)
	r.Path("/backgrounds").Methods(http.MethodGet).HandlerFunc(dh.ListBackgrounds)

	// Widget interactive routes (mounted by each widget's Module.RegisterRoutes)
	for _, m := range deps.modules {
		m.RegisterRoutes(r)
	}

	// Theme routes
	th := handlers.NewThemeHandler(deps.themeStore, deps.logger)
	r.Path("/themes").Methods(http.MethodGet).HandlerFunc(th.List)
	r.Path("/themes/{name}/icons/{icon}").Methods(http.MethodGet).HandlerFunc(th.GetIcon)
	r.Path("/themes/{name}/fonts/{font}").Methods(http.MethodGet).HandlerFunc(th.GetFont)
	r.Path("/themes/{name}/backgrounds/{file}").Methods(http.MethodGet).HandlerFunc(th.GetBackground)

	// Shared user-data (read)
	dataH := handlers.NewDataHandler(deps.notesStore, deps.imagesStore, deps.backgroundsStore, deps.logger)
	dataH.RegisterRead(r)
}

// attachWriteAPIs mounts all write (POST/PUT/DELETE) API endpoints on the given router.
func attachWriteAPIs(r *mux.Router, deps apiDeps) {
	dh := handlers.NewDashboardHandler(deps.dashStore, deps.themeStore, deps.backgroundsStore, deps.logger)

	r.Path("/dashboards").Methods(http.MethodPost).HandlerFunc(dh.Create)
	r.Path("/dashboards/upload").Methods(http.MethodPost).HandlerFunc(dh.Upload)
	r.Path("/dashboards/{id}").Methods(http.MethodPut).HandlerFunc(dh.Update)
	r.Path("/dashboards/{id}").Methods(http.MethodDelete).HandlerFunc(dh.Delete)
	r.Path("/dashboards/{id}/assets/{path:.*}").Methods(http.MethodDelete).HandlerFunc(dh.DeleteAsset)

	// Dashboard auth routes (editor only)
	r.Path("/dashboards/{id}/auth").Methods(http.MethodGet).HandlerFunc(dh.GetAuth)
	r.Path("/dashboards/{id}/auth").Methods(http.MethodPut).HandlerFunc(dh.SetAuth)
	r.Path("/dashboards/{id}/auth").Methods(http.MethodDelete).HandlerFunc(dh.DeleteAuth)

	// Theme admin CRUD (editor only)
	th := handlers.NewThemeHandler(deps.themeStore, deps.logger)
	r.Path("/themes/upload").Methods(http.MethodPost).HandlerFunc(th.Upload)
	r.Path("/themes/{name}").Methods(http.MethodDelete).HandlerFunc(th.Delete)
	r.Path("/themes/{name}/download").Methods(http.MethodGet).HandlerFunc(th.Download)

	// Shared user-data (write)
	dataH := handlers.NewDataHandler(deps.notesStore, deps.imagesStore, deps.backgroundsStore, deps.logger)
	dataH.RegisterWrite(r)
}
