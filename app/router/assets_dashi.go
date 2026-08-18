package router

import (
	"net/http"
	"strconv"

	"github.com/andresbott/dashi/internal/dashboard/browser"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/gorilla/mux"
)

// attachDashiAssets mounts the browser-stack static assets under
// /_dashi/. The underscore prefix matters: dashboard IDs are six-character
// lowercase alphanumerics, so an all-alphanumeric prefix could collide
// with a real dashboard path.
func attachDashiAssets(r *mux.Router, assets *browser.Assets, themeStore *themes.Store) {
	const cssType = "text/css; charset=utf-8"
	const jsType = "text/javascript; charset=utf-8"

	serve := func(contentType string, body func(*http.Request) ([]byte, bool)) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			data, ok := body(req)
			if !ok {
				http.NotFound(w, req)
				return
			}
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Content-Length", strconv.Itoa(len(data)))
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Cache-Control", "no-cache")
			_, _ = w.Write(data)
		}
	}

	sub := r.PathPrefix("/_dashi").Subrouter()

	sub.Path("/assets/viewer.css").Methods(http.MethodGet).
		HandlerFunc(serve(cssType, func(*http.Request) ([]byte, bool) {
			return assets.ViewerCSS(), true
		}))

	sub.Path("/assets/widgets.css").Methods(http.MethodGet).
		HandlerFunc(serve(cssType, func(*http.Request) ([]byte, bool) {
			return assets.WidgetsCSS(), true
		}))

	sub.Path("/assets/dashi.js").Methods(http.MethodGet).
		HandlerFunc(serve(jsType, func(*http.Request) ([]byte, bool) {
			return assets.DashiJS(), true
		}))

	// Theme name is a single path segment resolved through the theme
	// store; unknown themes yield the default palette rather than a 404,
	// so a deleted theme degrades to default colours instead of an
	// unstyled page.
	sub.Path("/theme/{name}.css").Methods(http.MethodGet).
		HandlerFunc(serve(cssType, func(req *http.Request) ([]byte, bool) {
			return themeStore.ThemeCSS(mux.Vars(req)["name"]), true
		}))

	sub.Path("/widgets/{type}.js").Methods(http.MethodGet).
		HandlerFunc(serve(jsType, func(req *http.Request) ([]byte, bool) {
			return assets.WidgetJS(mux.Vars(req)["type"])
		}))
}
