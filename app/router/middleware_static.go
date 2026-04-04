package router

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/andresbott/dashi/internal/dashboard"
	dashimage "github.com/andresbott/dashi/internal/dashboard/image"
	dashstatic "github.com/andresbott/dashi/internal/dashboard/static"
)

// NewStaticDashboardMiddleware returns middleware that intercepts requests
// for static and image dashboards. Static dashboards are rendered as HTML,
// image dashboards are rendered as PNG. Non-matching requests fall through
// to the next handler (SPA).
func NewStaticDashboardMiddleware(store *dashboard.Store, staticRenderer *dashstatic.Renderer, imageRenderer *dashimage.Renderer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			path := strings.TrimPrefix(r.URL.Path, "/")

			if strings.Contains(path, "/") || path == "" {
				next.ServeHTTP(w, r)
				return
			}

			dash, err := store.Get(path)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			switch dash.Type {
			case "static":
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				if err := staticRenderer.Render(w, dash); err != nil {
					http.Error(w, "failed to render dashboard", http.StatusInternalServerError)
				}

			case "image":
				var buf bytes.Buffer
				if err := staticRenderer.Render(&buf, dash); err != nil {
					http.Error(w, "failed to render dashboard HTML", http.StatusInternalServerError)
					return
				}

				width := 0
				height := 0
				if dash.ImageConfig != nil {
					width = dash.ImageConfig.Width
					height = dash.ImageConfig.Height
				}

				pngData, err := imageRenderer.Render(buf.String(), width, height)
				if err != nil {
					http.Error(w, "failed to render dashboard image", http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "image/png")
				w.Write(pngData)

			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}
