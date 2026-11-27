package router

import (
	"bytes"
	"fmt"
	"image"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/andresbott/dashi/internal/dashboard"
	"github.com/andresbott/dashi/internal/dashboard/browser"
	dashimage "github.com/andresbott/dashi/internal/dashboard/image"
	dashstatic "github.com/andresbott/dashi/internal/dashboard/static"
	"github.com/andresbott/dashi/internal/themes"
	"github.com/andresbott/dashi/internal/widgetthemes"
)

// NewDashboardMiddleware returns middleware that intercepts single-segment
// GET paths that name a dashboard and renders it. Image dashboards
// requested with display headers render as PNG for e-ink clients;
// everything else renders as browser HTML. Non-dashboard paths fall
// through to the next handler. editorBase is the editor's public URL, the
// target of the optional Edit button; empty means no button.
func NewDashboardMiddleware(store *dashboard.Store, browserRenderer *browser.Renderer, staticRenderer *dashstatic.Renderer, imageRenderer *dashimage.Renderer, themeStore *themes.Store, widgetThemeStore *widgetthemes.Store, editorBase string) func(http.Handler) http.Handler {
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

			// A missing or dangling theme reference is not an error:
			// Resolve falls back to the builtin default.
			theme := themeStore.Resolve(dash.ThemeID)
			if dash.Type == "image" && hasDisplayHeaders(r) {
				serveImageDashboard(w, r, dash, store, staticRenderer, imageRenderer, theme, themeStore, widgetThemeStore)
				return
			}
			serveBrowserDashboard(w, r, dash, store, browserRenderer, theme, editLink(dash, editorBase))
		})
	}
}

// editLink returns the editor URL for dash's Edit button, or "" when the
// button is not drawn: image dashboards never get one (their browser view is
// a preview of the PNG), nor does a dashboard with the setting off, nor any
// dashboard while the editor has no public URL. base is already normalised
// (no trailing slash) at config load.
func editLink(dash dashboard.Dashboard, base string) string {
	if dash.Type == "image" || !dash.ShowEditButton || base == "" {
		return ""
	}
	return base + "/dashboards/" + url.PathEscape(dash.ID) + "/edit"
}

// serveBrowserDashboard renders a dashboard as HTML for a real browser.
// Theme colours, background and fonts reach the page through the two
// linked stylesheets; the palette is only carried for widget renderers.
// editURL is the Edit button's target; empty draws no button.
func serveBrowserDashboard(w http.ResponseWriter, r *http.Request, dash dashboard.Dashboard, store *dashboard.Store, renderer *browser.Renderer, theme themes.Theme, editURL string) {
	pageIdx, ok := parsePageIndex(r, len(dash.Pages))
	if !ok {
		http.NotFound(w, r)
		return
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

	pageNames := make([]string, len(dash.Pages))
	for i, p := range dash.Pages {
		pageNames[i] = p.Name
	}

	data := browser.RenderData{
		Name:        dash.Name,
		DashboardID: dash.ID,
		ThemeID:     dash.ThemeID,
		WidgetTheme: dash.WidgetTheme,
		ColorMode:   colorMode,
		MaxWidth:    dash.Container.MaxWidth,
		HAlign:      dash.Container.HorizontalAlign,
		VAlign:      dash.Container.VerticalAlign,
		CustomCSS:   store.GetCustomCSS(dash.ID),
		QueryParams: queryParams,
		Rows:        dash.Pages[pageIdx].Rows,
		PageIndex:   pageIdx,
		TotalPages:  len(dash.Pages),
		PageNames:   pageNames,
		Palette:     theme.Palette(colorMode),
		EditURL:     editURL,
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

// maxDisplaySide caps the requested width and height in pixels. A render
// allocates width×height RGBA up front, so an unbounded value lets one request
// exhaust memory; the largest real panels (13.3" Spectra 6, 1600×1200) sit far
// below it. The editor's preview dialog uses the same limit
// (MAX_PREVIEW_SIDE in webui/src/lib/einkPreview.ts).
const maxDisplaySide = 4096

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
	if err != nil || width <= 0 || width > maxDisplaySide {
		return displayRequest{}, fmt.Errorf("invalid X-Display-Width: %s (must be 1 to %d)", widthStr, maxDisplaySide)
	}

	heightStr := getDisplayParam(r, "X-Display-Height", "height")
	if heightStr == "" {
		return displayRequest{}, fmt.Errorf("missing X-Display-Height header or height query parameter")
	}
	height, err := strconv.Atoi(heightStr)
	if err != nil || height <= 0 || height > maxDisplaySide {
		return displayRequest{}, fmt.Errorf("invalid X-Display-Height: %s (must be 1 to %d)", heightStr, maxDisplaySide)
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
func serveImageDashboard(w http.ResponseWriter, r *http.Request, dash dashboard.Dashboard, store *dashboard.Store, staticRenderer *dashstatic.Renderer, imageRenderer *dashimage.Renderer, theme themes.Theme, themeStore *themes.Store, widgetThemeStore *widgetthemes.Store) {
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
	renderData := buildRenderData(dash, pageIdx, r.URL.Query(), store, theme, widgetThemeStore)
	bgCSS, bgImageData := resolveImageBackground(theme, dash.ColorMode, themeStore)
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
	output, contentType, err := encodeForFormat(img, dreq.Format, dash.Dither)
	if err != nil {
		http.Error(w, "failed to encode image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(output)
}

// encodeForFormat encodes an RGBA image in the requested display format.
// dither is the dashboard's dither mode; it applies to every dithered
// format and is ignored for plain png.
func encodeForFormat(img *image.RGBA, format, dither string) ([]byte, string, error) {
	switch format {
	case "png":
		data, err := dashimage.EncodePNG(img)
		return data, "image/png", err

	case "png-bw":
		rgba := dashimage.DitherBWRGBA(img, dither)
		data, err := dashimage.EncodePNG(rgba)
		return data, "image/png", err

	case "png-spectra6":
		rgba := dashimage.DitherSpectra6RGBA(img, dither)
		data, err := dashimage.EncodePNG(rgba)
		return data, "image/png", err

	case "bw":
		return dashimage.DitherBWPacked(img, dither), "application/octet-stream", nil

	case "spectra6":
		return dashimage.DitherSpectra6Packed(img, dither), "application/octet-stream", nil

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

// buildRenderData constructs the image-stack render data for a dashboard
// page. The palette is flattened because litehtml cannot render alpha: a
// translucent role must reach it as an opaque colour composited over bg.
func buildRenderData(dash dashboard.Dashboard, pageIdx int, query map[string][]string, store *dashboard.Store, theme themes.Theme, widgetThemeStore *widgetthemes.Store) dashstatic.RenderData {
	widgetTheme := dash.WidgetTheme
	if widgetTheme == "" {
		widgetTheme = widgetthemes.DefaultName
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
		WidgetTheme: widgetTheme,
		ColorMode:   dash.ColorMode,
		FontFamily:  widgetThemeStore.FontFamily(widgetTheme),
		CustomCSS:   store.GetCustomCSS(dash.ID),
		QueryParams: queryParams,
		Rows:        dash.Pages[pageIdx].Rows,
		PageIndex:   pageIdx,
		TotalPages:  len(dash.Pages),
		Palette:     theme.Palette(dash.ColorMode).Flatten(),
	}
}

// resolveImageBackground resolves one variant server-side. Litehtml cannot
// read CSS custom properties, so it needs concrete values. Fit, position and
// repeat are deliberately ignored here — RenderToImage pre-paints the image
// with cover maths and pngContainer.DrawImage ignores those properties.
//
// When an image ref is present and loads successfully, returns "transparent"
// as the css. This prevents the template's {{else}}{{.Palette.BG}}{{end}}
// fallback from painting an opaque theme colour over the pre-composited image.
func resolveImageBackground(theme themes.Theme, colorMode string, themeStore *themes.Store) (css string, imageData []byte) {
	css, ref := themes.ImageCSS(theme, colorMode)
	if ref == "" {
		return css, nil
	}
	data, err := themeStore.LoadImage(theme.ID, ref)
	if err != nil {
		return css, nil
	}
	return "transparent", data
}
