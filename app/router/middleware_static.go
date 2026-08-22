package router

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/dashboard/browser"
	"github.com/andresbott/dashi/internal/data/images"
	dashimage "github.com/andresbott/dashi/internal/dashboard/image"
	dashstatic "github.com/andresbott/dashi/internal/dashboard/static"
	"github.com/andresbott/dashi/internal/themes"
)

// NewDashboardMiddleware returns middleware that intercepts single-segment
// GET paths that name a dashboard and renders it. Image dashboards
// requested with display headers render as PNG for e-ink clients;
// everything else renders as browser HTML. Non-dashboard paths fall
// through to the next handler.
func NewDashboardMiddleware(store *dashboard.Store, browserRenderer *browser.Renderer, staticRenderer *dashstatic.Renderer, imageRenderer *dashimage.Renderer, themeStore *themes.Store, backgroundsStore *images.Store) func(http.Handler) http.Handler {
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

			if dash.Type == "image" && hasDisplayHeaders(r) {
				serveImageDashboard(w, r, dash, store, staticRenderer, imageRenderer, themeStore, backgroundsStore)
				return
			}
			serveBrowserDashboard(w, r, dash, store, browserRenderer, themeStore, backgroundsStore)
		})
	}
}

// serveBrowserDashboard renders a dashboard as HTML for a real browser.
func serveBrowserDashboard(w http.ResponseWriter, r *http.Request, dash dashboard.Dashboard, store *dashboard.Store, renderer *browser.Renderer, themeStore *themes.Store, backgroundsStore *images.Store) {
	pageIdx, ok := parsePageIndex(r, len(dash.Pages))
	if !ok {
		http.NotFound(w, r)
		return
	}

	theme := dash.Theme
	if theme == "" {
		theme = "default"
	}
	colorMode := dash.ColorMode
	if colorMode != "dark" {
		colorMode = "light"
	}

	queryParams := make(map[string]string)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			queryParams[k] = v[0]
		}
	}

	// The browser fetches image backgrounds over HTTP, so the shell emits a
	// URL. Inlining the bytes as a data URI (what the litehtml stack needs)
	// would add the whole file to every response, uncacheable, for no gain.
	bgCSS := buildBrowserBackground(dash)

	pageNames := make([]string, len(dash.Pages))
	for i, p := range dash.Pages {
		pageNames[i] = p.Name
	}

	data := browser.RenderData{
		Name:        dash.Name,
		DashboardID: dash.ID,
		Theme:       theme,
		ColorMode:   colorMode,
		MaxWidth:    dash.Container.MaxWidth,
		HAlign:      dash.Container.HorizontalAlign,
		VAlign:      dash.Container.VerticalAlign,
		AccentColor: dash.AccentColor,
		CustomCSS:   store.GetCustomCSS(dash.ID),
		Background:  bgCSS,
		QueryParams: queryParams,
		Rows:        dash.Pages[pageIdx].Rows,
		PageIndex:   pageIdx,
		TotalPages:  len(dash.Pages),
		PageNames:   pageNames,
		Palette:     themeStore.Palette(theme, colorMode),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if interval := dash.Pages[pageIdx].RefreshInterval; interval > 0 {
		w.Header().Set("X-Refresh-Interval", strconv.Itoa(interval))
	}
	if err := renderer.Render(w, data); err != nil {
		http.Error(w, "failed to render dashboard", http.StatusInternalServerError)
		return
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

// pageRedirectTarget builds the swipe-navigation redirect target: the current
// request path with the page query replaced. Reusing the request path verbatim
// is safe here — the middleware only routes to a dashboard after store.Get
// accepted the single path segment, and valid dashboard IDs are [a-z0-9]+, so
// the target can hold neither a slash nor a scheme and cannot leave this host.
func pageRedirectTarget(r *http.Request, page int) string {
	return r.URL.Path + "?page=" + strconv.Itoa(page)
}

// serveImageDashboard handles rendering of image-type dashboards.
func serveImageDashboard(w http.ResponseWriter, r *http.Request, dash dashboard.Dashboard, store *dashboard.Store, staticRenderer *dashstatic.Renderer, imageRenderer *dashimage.Renderer, themeStore *themes.Store, backgroundsStore *images.Store) {
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
		http.Redirect(w, r, pageRedirectTarget(r, nextPage), http.StatusTemporaryRedirect) //nolint:gosec // G710: target is a validated [a-z0-9]+ dashboard path, see pageRedirectTarget
		return
	case "swipe_left":
		prevPage := (pageIdx - 1 + totalPages) % totalPages
		http.Redirect(w, r, pageRedirectTarget(r, prevPage), http.StatusTemporaryRedirect) //nolint:gosec // G710: target is a validated [a-z0-9]+ dashboard path, see pageRedirectTarget
		return
	}

	// Render dashboard to image
	renderData := buildRenderData(dash, pageIdx, r.URL.Query(), store, themeStore)
	bgCSS, bgImageData := buildBackground(dash, store, themeStore, backgroundsStore)
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
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(output) //nolint:gosec // G705: encoded image/binary bytes served with explicit Content-Type and nosniff; not HTML
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
		Palette:     themeStore.Palette(theme, dash.ColorMode),
	}
}

// buildBackground returns the CSS background value and, for image backgrounds,
// the raw image bytes (since litehtml doesn't support CSS background-image).
func buildBackground(dash dashboard.Dashboard, dashStore *dashboard.Store, themeStore *themes.Store, backgroundsStore *images.Store) (css string, imageData []byte) {
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
		return buildImageBackground(bg.Value, dash.ID, dashStore, themeStore, backgroundsStore)
	default:
		return "", nil
	}
}

// buildBrowserBackground returns the CSS background value for the browser
// stack. Unlike buildBackground (which the litehtml stack uses, and which
// must inline image bytes because litehtml fetches nothing over the
// network), this emits a URL pointing at the route that already serves those
// bytes: the browser caches it, and a large background no longer bloats every
// page response.
//
// The referenced file is not read here, so a missing background simply 404s
// and the page falls back to the theme background painted on <html>.
func buildBrowserBackground(dash dashboard.Dashboard) string {
	bg := dash.Background
	if bg == nil || bg.Type == "none" || bg.Value == "" {
		return ""
	}
	switch bg.Type {
	case "color", "gradient":
		return bg.Value
	case "image":
		url, ok := backgroundImageURL(bg.Value, dash.ID)
		if !ok {
			return ""
		}
		return "url('" + url + "') center/cover no-repeat"
	default:
		return ""
	}
}

// backgroundImageURL maps a background reference to the existing read route
// that serves it. The dashboard-asset route sits behind the per-dashboard
// auth middleware, so a protected dashboard's background stays protected.
func backgroundImageURL(bgValue, dashID string) (string, bool) {
	switch {
	case strings.HasPrefix(bgValue, "theme:"):
		rest := bgValue[len("theme:"):]
		slashIdx := strings.Index(rest, "/")
		if slashIdx < 0 {
			return "", false
		}
		themeName, fileName := rest[:slashIdx], rest[slashIdx+1:]
		if themeName == "" || fileName == "" {
			return "", false
		}
		return "/api/v0/themes/" + urlSegment(themeName) + "/backgrounds/" + urlSegment(fileName), true

	case strings.HasPrefix(bgValue, "dashboard:"):
		assetPath := bgValue[len("dashboard:"):]
		if assetPath == "" {
			return "", false
		}
		// The asset route matches {path:.*}, so slashes stay slashes.
		segments := strings.Split(assetPath, "/")
		for i, s := range segments {
			segments[i] = urlSegment(s)
		}
		return "/api/v0/dashboards/" + urlSegment(dashID) + "/assets/" + strings.Join(segments, "/"), true

	case strings.HasPrefix(bgValue, "shared:"):
		fileName := bgValue[len("shared:"):]
		if fileName == "" {
			return "", false
		}
		return "/api/v0/data/backgrounds/" + urlSegment(fileName), true
	}
	return "", false
}

// urlSegment percent-encodes one path segment. It additionally escapes the
// single quote, which url.PathEscape leaves alone but which would terminate
// the url('…') CSS value the caller builds.
func urlSegment(s string) string {
	return strings.ReplaceAll(url.PathEscape(s), "'", "%27")
}

// buildImageBackground loads and encodes an image background.
func buildImageBackground(bgValue, dashID string, dashStore *dashboard.Store, themeStore *themes.Store, backgroundsStore *images.Store) (css string, imageData []byte) {
	data, fileName, err := loadBackgroundImage(bgValue, dashID, dashStore, themeStore, backgroundsStore)
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
func loadBackgroundImage(bgValue, dashID string, dashStore *dashboard.Store, themeStore *themes.Store, backgroundsStore *images.Store) (data []byte, fileName string, err error) {
	if strings.HasPrefix(bgValue, "theme:") {
		return loadThemeBackground(bgValue, themeStore)
	}
	if strings.HasPrefix(bgValue, "dashboard:") {
		fileName = bgValue[len("dashboard:"):]
		data, _, err = dashStore.GetAsset(dashID, fileName)
		return data, fileName, err
	}
	if strings.HasPrefix(bgValue, "shared:") {
		fileName = bgValue[len("shared:"):]
		if backgroundsStore == nil {
			return nil, fileName, fmt.Errorf("backgrounds store not configured")
		}
		data, _, err = backgroundsStore.Get(fileName)
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

