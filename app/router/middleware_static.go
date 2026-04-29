package router

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"log"
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
				serveImageDashboard(w, r, dash, store, staticRenderer, imageRenderer, themeStore)
			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}

// displayRequest holds validated display protocol headers.
type displayRequest struct {
	Format   string
	Width    int
	Height   int
	Rotation int
	Action   string
}

var validFormats = map[string]bool{
	"png": true, "png-bw": true, "png-spectra6": true,
	"bw": true, "spectra6": true,
}

var validRotations = map[int]bool{0: true, 90: true, 180: true, 270: true}

// hasDisplayHeaders returns true if any display protocol header or query parameter is present.
func hasDisplayHeaders(r *http.Request) bool {
	return r.Header.Get("X-Display-Format") != "" ||
		r.Header.Get("X-Display-Width") != "" ||
		r.Header.Get("X-Display-Height") != "" ||
		r.URL.Query().Get("format") != "" ||
		r.URL.Query().Get("width") != "" ||
		r.URL.Query().Get("height") != ""
}

// getDisplayParam returns the value from the header first, falling back to the query parameter.
func getDisplayParam(r *http.Request, header, queryParam string) string {
	if v := r.Header.Get(header); v != "" {
		return v
	}
	return r.URL.Query().Get(queryParam)
}

// parseDisplayHeaders validates and extracts required display headers/query params.
// Headers take precedence over query parameters.
func parseDisplayHeaders(r *http.Request) (displayRequest, error) {
	format := getDisplayParam(r, "X-Display-Format", "format")
	if format == "" {
		return displayRequest{}, fmt.Errorf("missing X-Display-Format header or format query parameter")
	}
	if !validFormats[format] {
		return displayRequest{}, fmt.Errorf("invalid X-Display-Format: %s", format)
	}

	widthStr := getDisplayParam(r, "X-Display-Width", "width")
	if widthStr == "" {
		return displayRequest{}, fmt.Errorf("missing X-Display-Width header or width query parameter")
	}
	width, err := strconv.Atoi(widthStr)
	if err != nil || width <= 0 {
		return displayRequest{}, fmt.Errorf("invalid X-Display-Width: %s", widthStr)
	}

	heightStr := getDisplayParam(r, "X-Display-Height", "height")
	if heightStr == "" {
		return displayRequest{}, fmt.Errorf("missing X-Display-Height header or height query parameter")
	}
	height, err := strconv.Atoi(heightStr)
	if err != nil || height <= 0 {
		return displayRequest{}, fmt.Errorf("invalid X-Display-Height: %s", heightStr)
	}

	rotation := 0
	if rotStr := getDisplayParam(r, "X-Display-Rotation", "rotation"); rotStr != "" {
		rotation, err = strconv.Atoi(rotStr)
		if err != nil || !validRotations[rotation] {
			return displayRequest{}, fmt.Errorf("invalid X-Display-Rotation: %s (must be 0, 90, 180, or 270)", rotStr)
		}
	}

	action := getDisplayParam(r, "X-Action", "action")
	if action == "" {
		action = "refresh"
	}

	return displayRequest{Format: format, Width: width, Height: height, Rotation: rotation, Action: action}, nil
}

// serveImageDashboard handles rendering of image-type dashboards.
func serveImageDashboard(w http.ResponseWriter, r *http.Request, dash dashboard.Dashboard, store *dashboard.Store, staticRenderer *dashstatic.Renderer, imageRenderer *dashimage.Renderer, themeStore *themes.Store) {
	// TEMPORARY: log request headers and query params for debugging
	log.Printf("[DEBUG] %s %s", r.Method, r.URL.String())
	for name, values := range r.Header {
		log.Printf("[DEBUG]   Header: %s = %s", name, strings.Join(values, ", "))
	}

	// No display headers at all → serve HTML preview (browser access)
	if !hasDisplayHeaders(r) {
		serveImageHTMLPreview(w, r, dash, store, staticRenderer, themeStore)
		return
	}

	dreq, err := parseDisplayHeaders(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pageIdx, ok := parsePageIndex(r, len(dash.Pages))
	if !ok {
		http.NotFound(w, r)
		return
	}

	// Handle swipe navigation
	totalPages := len(dash.Pages)
	switch dreq.Action {
	case "swipe_right":
		nextPage := (pageIdx + 1) % totalPages
		http.Redirect(w, r, r.URL.Path+"?page="+strconv.Itoa(nextPage), http.StatusTemporaryRedirect)
		return
	case "swipe_left":
		prevPage := (pageIdx - 1 + totalPages) % totalPages
		http.Redirect(w, r, r.URL.Path+"?page="+strconv.Itoa(prevPage), http.StatusTemporaryRedirect)
		return
	}

	// Render dashboard to image
	renderData := buildRenderData(dash, pageIdx, r.URL.Query(), store, themeStore)
	bgCSS, bgImageData := buildBackground(dash, store, themeStore)
	renderData.BackgroundCSS = bgCSS

	var buf bytes.Buffer
	if err := staticRenderer.Render(&buf, renderData); err != nil {
		http.Error(w, "failed to render dashboard HTML", http.StatusInternalServerError)
		return
	}

	renderWidth, renderHeight := dreq.Width, dreq.Height
	if dreq.Rotation == 90 || dreq.Rotation == 270 {
		renderWidth, renderHeight = dreq.Height, dreq.Width
	}

	img, err := imageRenderer.RenderToImage(buf.String(), renderWidth, renderHeight, bgImageData)
	if err != nil {
		http.Error(w, "failed to render dashboard image", http.StatusInternalServerError)
		return
	}

	if dreq.Rotation != 0 {
		img = dashimage.RotateImage(img, dreq.Rotation)
	}

	// Set refresh interval
	if dash.Pages[pageIdx].RefreshInterval > 0 {
		w.Header().Set("X-Refresh-Interval", strconv.Itoa(dash.Pages[pageIdx].RefreshInterval))
	}

	// Encode and serve
	output, contentType, err := encodeForFormat(img, dreq.Format, dreq.Width, dreq.Height)
	if err != nil {
		http.Error(w, "failed to encode image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	_, _ = w.Write(output)
}

// serveImageHTMLPreview renders the dashboard as HTML when display headers are missing.
func serveImageHTMLPreview(w http.ResponseWriter, r *http.Request, dash dashboard.Dashboard, store *dashboard.Store, staticRenderer *dashstatic.Renderer, themeStore *themes.Store) {
	pageIdx, ok := parsePageIndex(r, len(dash.Pages))
	if !ok {
		http.NotFound(w, r)
		return
	}

	renderData := buildRenderData(dash, pageIdx, r.URL.Query(), store, themeStore)
	bgCSS, _ := buildBackground(dash, store, themeStore)
	renderData.BackgroundCSS = bgCSS

	var buf bytes.Buffer
	if err := staticRenderer.Render(&buf, renderData); err != nil {
		http.Error(w, "failed to render dashboard HTML", http.StatusInternalServerError)
		return
	}

	serveHTMLPreview(w, buf.String(), 0, 0)
}

// encodeForFormat encodes an RGBA image in the requested display format.
func encodeForFormat(img *image.RGBA, format string, width, height int) ([]byte, string, error) {
	switch format {
	case "png":
		data, err := dashimage.EncodePNG(img)
		return data, "image/png", err

	case "png-bw":
		rgba := dashimage.DitherBWRGBA(img)
		data, err := dashimage.EncodePNG(rgba)
		return data, "image/png", err

	case "png-spectra6":
		rgba := dashimage.DitherSpectra6RGBA(img)
		data, err := dashimage.EncodePNG(rgba)
		return data, "image/png", err

	case "bw":
		return dashimage.DitherBWPacked(img), "application/octet-stream", nil

	case "spectra6":
		return dashimage.DitherSpectra6Packed(img), "application/octet-stream", nil

	default:
		return nil, "", fmt.Errorf("unsupported format: %s", format)
	}
}

// parsePageIndex extracts and validates the page index from the request.
func parsePageIndex(r *http.Request, totalPages int) (int, bool) {
	if totalPages == 0 {
		return 0, false
	}

	pageIdx := 0
	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		var err error
		pageIdx, err = strconv.Atoi(pageParam)
		if err != nil || pageIdx < 0 {
			return 0, false
		}
	}

	if pageIdx >= totalPages {
		return 0, false
	}

	return pageIdx, true
}

// buildRenderData constructs the render data structure for a dashboard page.
func buildRenderData(dash dashboard.Dashboard, pageIdx int, query map[string][]string, store *dashboard.Store, themeStore *themes.Store) dashstatic.RenderData {
	theme := dash.Theme
	if theme == "" {
		theme = "default"
	}

	fontFamily := ""
	if themeInfo, ok := themeStore.Get(theme); ok && len(themeInfo.Fonts) > 0 {
		fontFamily = themeInfo.Fonts[0].Name
	}

	queryParams := make(map[string]string)
	for k, v := range query {
		if len(v) > 0 {
			queryParams[k] = v[0]
		}
	}

	return dashstatic.RenderData{
		Name:        dash.Name,
		DashboardID: dash.ID,
		MaxWidth:    dash.Container.MaxWidth,
		HAlign:      dash.Container.HorizontalAlign,
		VAlign:      dash.Container.VerticalAlign,
		Theme:       theme,
		ColorMode:   dash.ColorMode,
		FontFamily:  fontFamily,
		CustomCSS:   store.GetCustomCSS(dash.ID),
		QueryParams: queryParams,
		Rows:        dash.Pages[pageIdx].Rows,
		PageIndex:   pageIdx,
		TotalPages:  len(dash.Pages),
	}
}

// serveHTMLPreview writes the HTML preview with optional dimension wrapper.
func serveHTMLPreview(w http.ResponseWriter, html string, width, height int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	inlinedHTML := inlineLocalImages(html)

	if width > 0 || height > 0 {
		style := buildWrapperStyle(width, height)
		wrapper := `<div style="` + style + `">`
		_, _ = w.Write([]byte(wrapper))
		_, _ = w.Write([]byte(inlinedHTML))
		_, _ = w.Write([]byte("</div>"))
	} else {
		_, _ = w.Write([]byte(inlinedHTML))
	}
}

// buildWrapperStyle creates CSS style for preview wrapper.
func buildWrapperStyle(width, height int) string {
	style := "margin:2rem auto;border:1px solid #ccc;"
	if width > 0 {
		style += "width:" + strconv.Itoa(width) + "px;"
	}
	if height > 0 {
		style += "height:" + strconv.Itoa(height) + "px;"
	}
	style += "overflow:hidden;"
	return style
}

// buildBackground returns the CSS background value and, for image backgrounds,
// the raw image bytes (since litehtml doesn't support CSS background-image).
func buildBackground(dash dashboard.Dashboard, dashStore *dashboard.Store, themeStore *themes.Store) (css string, imageData []byte) {
	bg := dash.Background
	if bg == nil || bg.Type == "none" || bg.Value == "" {
		return "", nil
	}
	switch bg.Type {
	case "color":
		return bg.Value, nil
	case "gradient":
		return bg.Value, nil
	case "image":
		return buildImageBackground(bg.Value, dash.ID, dashStore, themeStore)
	default:
		return "", nil
	}
}

// buildImageBackground loads and encodes an image background.
func buildImageBackground(bgValue, dashID string, dashStore *dashboard.Store, themeStore *themes.Store) (css string, imageData []byte) {
	data, fileName, err := loadBackgroundImage(bgValue, dashID, dashStore, themeStore)
	if err != nil {
		return "", nil
	}

	mimeType := mime.TypeByExtension(filepath.Ext(fileName))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	dataURI := "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data)
	return "url('" + dataURI + "') center/cover no-repeat", data
}

// loadBackgroundImage loads background image data from theme or dashboard assets.
func loadBackgroundImage(bgValue, dashID string, dashStore *dashboard.Store, themeStore *themes.Store) (data []byte, fileName string, err error) {
	if strings.HasPrefix(bgValue, "theme:") {
		return loadThemeBackground(bgValue, themeStore)
	}
	if strings.HasPrefix(bgValue, "dashboard:") {
		fileName = bgValue[len("dashboard:"):]
		data, _, err = dashStore.GetAsset(dashID, fileName)
		return data, fileName, err
	}
	return nil, "", fmt.Errorf("unsupported background type")
}

// loadThemeBackground loads a theme background image.
func loadThemeBackground(bgValue string, themeStore *themes.Store) ([]byte, string, error) {
	rest := bgValue[6:]
	slashIdx := strings.Index(rest, "/")
	if slashIdx < 0 {
		return nil, "", fmt.Errorf("invalid theme background format")
	}
	themeName := rest[:slashIdx]
	fileName := rest[slashIdx+1:]
	data, err := themeStore.GetBackgroundData(themeName, fileName)
	return data, fileName, err
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

