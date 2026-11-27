package router

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/widgets"
	"github.com/andresbott/dashi/internal/widgetthemes"
	"github.com/gorilla/mux"

	"github.com/andresbott/dashi/app/router/handlers"
	"github.com/andresbott/dashi/internal/data/images"
	"github.com/andresbott/dashi/internal/data/notes"
)

// apiDeps holds shared dependencies for API route handlers.
type apiDeps struct {
	dashStore        *dashboard.Store
	themeStore       *themes.Store
	widgetThemeStore *widgetthemes.Store
	notesStore       *notes.Store
	imagesStore      *images.Store
	logger           *slog.Logger
	modules          []widgets.Module
	publicViewer     handlers.PublicViewer
	publicEditor     handlers.PublicEditor
}

// attachReadAPIs mounts all read-only (GET) API endpoints on the given router.
func attachReadAPIs(r *mux.Router, deps apiDeps) {
	// Health check
	r.Path("/health").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Runtime info the SPA reads at init (where the public viewer and editor live)
	ih := handlers.NewInfoHandler(deps.publicViewer, deps.publicEditor)
	r.Path("/info").Methods(http.MethodGet).HandlerFunc(ih.Get)

	// Dashboard routes (read)
	dh := handlers.NewDashboardHandler(deps.dashStore, deps.logger)
	r.Path("/dashboards").Methods(http.MethodGet).HandlerFunc(dh.List)
	r.Path("/dashboards/{id}").Methods(http.MethodGet).HandlerFunc(dh.Get)
	r.Path("/dashboards/{id}/download").Methods(http.MethodGet).HandlerFunc(dh.Download)
	r.Path("/dashboards/{id}/assets").Methods(http.MethodGet).HandlerFunc(dh.ListAssets)
	r.Path("/dashboards/{id}/assets/{path:.*}").Methods(http.MethodGet).HandlerFunc(dh.GetAsset)

	// Theme entity (read)
	tmh := handlers.NewThemeHandler(deps.themeStore, deps.dashStore, deps.logger)
	r.Path("/themes").Methods(http.MethodGet).HandlerFunc(tmh.List)
	r.Path("/themes/{id}").Methods(http.MethodGet).HandlerFunc(tmh.Get)
	r.Path("/themes/{id}/assets").Methods(http.MethodGet).HandlerFunc(tmh.ListAssets)
	r.Path("/themes/{id}/assets/{path:.*}").Methods(http.MethodGet).HandlerFunc(tmh.GetAsset)

	// Widget interactive routes (mounted by each widget's Module.RegisterRoutes)
	for _, m := range deps.modules {
		m.RegisterRoutes(r)
	}

	// Widget theme routes
	th := handlers.NewWidgetThemeHandler(deps.widgetThemeStore, deps.logger)
	r.Path("/widget-themes").Methods(http.MethodGet).HandlerFunc(th.List)
	r.Path("/widget-themes/{name}/icons/{icon}").Methods(http.MethodGet).HandlerFunc(th.GetIcon)
	r.Path("/widget-themes/{name}/fonts/{font}").Methods(http.MethodGet).HandlerFunc(th.GetFont)

	// Shared user-data (read)
	dataH := handlers.NewDataHandler(deps.notesStore, deps.imagesStore, deps.logger)
	dataH.RegisterRead(r)
}

// attachWriteAPIs mounts all write (POST/PUT/DELETE) API endpoints on the given router.
func attachWriteAPIs(r *mux.Router, deps apiDeps) {
	dh := handlers.NewDashboardHandler(deps.dashStore, deps.logger)

	r.Path("/dashboards").Methods(http.MethodPost).HandlerFunc(dh.Create)
	r.Path("/dashboards/upload").Methods(http.MethodPost).HandlerFunc(dh.Upload)
	r.Path("/dashboards/{id}/duplicate").Methods(http.MethodPost).HandlerFunc(dh.Duplicate)
	r.Path("/dashboards/{id}").Methods(http.MethodPut).HandlerFunc(dh.Update)
	r.Path("/dashboards/{id}").Methods(http.MethodDelete).HandlerFunc(dh.Delete)
	r.Path("/dashboards/{id}/assets/{path:.*}").Methods(http.MethodDelete).HandlerFunc(dh.DeleteAsset)

	// Dashboard auth routes (editor only)
	r.Path("/dashboards/{id}/auth").Methods(http.MethodGet).HandlerFunc(dh.GetAuth)
	r.Path("/dashboards/{id}/auth").Methods(http.MethodPut).HandlerFunc(dh.SetAuth)
	r.Path("/dashboards/{id}/auth").Methods(http.MethodDelete).HandlerFunc(dh.DeleteAuth)

	// Widget theme admin CRUD (editor only)
	th := handlers.NewWidgetThemeHandler(deps.widgetThemeStore, deps.logger)
	r.Path("/widget-themes/upload").Methods(http.MethodPost).HandlerFunc(th.Upload)
	r.Path("/widget-themes/{name}").Methods(http.MethodDelete).HandlerFunc(th.Delete)
	r.Path("/widget-themes/{name}/download").Methods(http.MethodGet).HandlerFunc(th.Download)

	// Shared user-data (write)
	dataH := handlers.NewDataHandler(deps.notesStore, deps.imagesStore, deps.logger)
	dataH.RegisterWrite(r)

	// Theme entity CRUD (editor only)
	tmh := handlers.NewThemeHandler(deps.themeStore, deps.dashStore, deps.logger)
	r.Path("/themes").Methods(http.MethodPost).HandlerFunc(tmh.Create)
	r.Path("/themes/{id}").Methods(http.MethodPut).HandlerFunc(tmh.Update)
	r.Path("/themes/{id}").Methods(http.MethodDelete).HandlerFunc(tmh.Delete)
	r.Path("/themes/{id}/copy").Methods(http.MethodPost).HandlerFunc(tmh.Copy)
	r.Path("/themes/{id}/assets/{path:.*}").Methods(http.MethodPost).HandlerFunc(tmh.SaveAsset)
	r.Path("/themes/{id}/assets/{path:.*}").Methods(http.MethodDelete).HandlerFunc(tmh.DeleteAsset)
}
