package router

import (
	"bytes"
	"encoding/base64"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/andresbott/dashi/internal/dashboard"
	dashimage "github.com/andresbott/dashi/internal/dashboard/image"
	dashstatic "github.com/andresbott/dashi/internal/dashboard/static"
	"github.com/andresbott/dashi/internal/themes"
)

// NewStaticDashboardMiddleware returns middleware that intercepts requests
// for static and image dashboards. Static dashboards are rendered as HTML,
// image dashboards are rendered as PNG. Non-matching requests fall through
// to the next handler (SPA).
func NewStaticDashboardMiddleware(store *dashboard.Store, staticRenderer *dashstatic.Renderer, imageRenderer *dashimage.Renderer, themeStore *themes.Store) func(http.Handler) http.Handler {
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
			case "image":
				// Parse page parameter
				pageIdx := 0
				if pageParam := r.URL.Query().Get("page"); pageParam != "" {
					var err error
					pageIdx, err = strconv.Atoi(pageParam)
					if err != nil || pageIdx < 0 {
						http.NotFound(w, r)
						return
					}
				}

				// Validate page index
				if len(dash.Pages) == 0 {
					http.NotFound(w, r)
					return
				}
				if pageIdx >= len(dash.Pages) {
					http.NotFound(w, r)
					return
				}

				// Build render data from selected page
				theme := dash.Theme
				if theme == "" {
					theme = "default"
				}
				fontFamily := ""
				if themeInfo, ok := themeStore.Get(theme); ok && len(themeInfo.Fonts) > 0 {
					fontFamily = themeInfo.Fonts[0].Name
				}

				queryParams := make(map[string]string)
				for k, v := range r.URL.Query() {
					if len(v) > 0 {
						queryParams[k] = v[0]
					}
				}

				renderData := dashstatic.RenderData{
					Name:        dash.Name,
					MaxWidth:    dash.Container.MaxWidth,
					HAlign:      dash.Container.HorizontalAlign,
					VAlign:      dash.Container.VerticalAlign,
					Theme:       theme,
					FontFamily:  fontFamily,
					CustomCSS:   store.GetCustomCSS(dash.ID),
					QueryParams: queryParams,
					Rows:        dash.Pages[pageIdx].Rows,
				}

				var buf bytes.Buffer
				if err := staticRenderer.Render(&buf, renderData); err != nil {
					http.Error(w, "failed to render dashboard HTML", http.StatusInternalServerError)
					return
				}

				width := 0
				height := 0
				if dash.ImageConfig != nil {
					width = dash.ImageConfig.Width
					height = dash.ImageConfig.Height
				}

				if r.URL.Query().Get("html") != "" {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					html := inlineLocalImages(buf.String())
					if width > 0 || height > 0 {
						style := "margin:2rem auto;border:1px solid #ccc;"
						if width > 0 {
							style += "width:" + strconv.Itoa(width) + "px;"
						}
						if height > 0 {
							style += "height:" + strconv.Itoa(height) + "px;"
						}
						style += "overflow:hidden;"
						wrapper := `<div style="` + style + `">`
						w.Write([]byte(wrapper))
						w.Write([]byte(html))
						w.Write([]byte("</div>"))
					} else {
						w.Write([]byte(html))
					}
					return
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

var imgSrcRe = regexp.MustCompile(`(<img\b[^>]*\bsrc=")([^"]+)(")`)

// inlineLocalImages replaces <img src="/absolute/path"> with base64 data URIs.
func inlineLocalImages(html string) string {
	return imgSrcRe.ReplaceAllStringFunc(html, func(match string) string {
		parts := imgSrcRe.FindStringSubmatch(match)
		if len(parts) != 4 {
			return match
		}
		src := parts[2]
		if !filepath.IsAbs(src) {
			return match
		}
		data, err := os.ReadFile(src)
		if err != nil {
			return match
		}
		mimeType := mime.TypeByExtension(filepath.Ext(src))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		dataURI := "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data)
		return parts[1] + dataURI + parts[3]
	})
}
