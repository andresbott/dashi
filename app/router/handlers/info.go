package handlers

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// PublicViewer describes where the public, server-rendered dashboard viewer is
// reachable from a browser.
type PublicViewer struct {
	Enabled bool
	Port    int
	BaseURL string // explicit public URL; wins over deriving one from the request
}

// InfoHandler serves the runtime facts the admin SPA needs at init time. The
// SPA is the editor UI only — it does not render dashboards — so its "view"
// links must point at the viewer server, which lives on a different port and,
// behind a reverse proxy, possibly on a different host.
type InfoHandler struct {
	viewer PublicViewer
}

func NewInfoHandler(viewer PublicViewer) *InfoHandler {
	return &InfoHandler{viewer: viewer}
}

type viewerInfo struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
}

type infoResponse struct {
	Viewer viewerInfo `json:"viewer"`
}

func (h *InfoHandler) Get(w http.ResponseWriter, r *http.Request) {
	resp := infoResponse{
		Viewer: viewerInfo{
			Enabled: h.viewer.Enabled,
			URL:     h.viewerURL(r),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		ErrorJSON(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// viewerURL returns the browser-reachable base URL of the viewer server,
// without a trailing slash. An explicit BaseURL wins; otherwise the URL is
// derived from the request host with the viewer port swapped in — which is
// what makes both `vite dev` (same host, /api proxied to the editor) and a
// plain two-port deployment work with no configuration at all.
//
// The result ends up in an SPA href, so only http/https are ever emitted.
func (h *InfoHandler) viewerURL(r *http.Request) string {
	if !h.viewer.Enabled {
		return ""
	}
	if h.viewer.BaseURL != "" {
		return normalizeBaseURL(h.viewer.BaseURL)
	}
	host := requestHost(r)
	if host == "" || h.viewer.Port <= 0 {
		return ""
	}
	return requestScheme(r) + "://" + net.JoinHostPort(host, strconv.Itoa(h.viewer.Port))
}

// normalizeBaseURL validates a configured public URL and strips its trailing
// slash. Anything that is not an absolute http/https URL is dropped rather
// than handed to the SPA.
func normalizeBaseURL(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return ""
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}
	return strings.TrimSuffix(u.Scheme+"://"+u.Host+u.Path, "/")
}

// requestHost returns the hostname the client used, without a port.
func requestHost(r *http.Request) string {
	host := r.Host
	if fwd := r.Header.Get("X-Forwarded-Host"); fwd != "" {
		host = strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	if !isPlainHost(host) {
		return ""
	}
	return host
}

// isPlainHost accepts only the characters a hostname or IP literal can hold,
// so a forged Host header cannot smuggle a path or another scheme into the
// URL the SPA links to.
func isPlainHost(host string) bool {
	if host == "" {
		return false
	}
	for _, c := range host {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		case c == '.', c == '-', c == ':', c == '_':
		default:
			return false
		}
	}
	return true
}

// requestScheme reports the scheme the client used, trusting
// X-Forwarded-Proto only when it names http or https.
func requestScheme(r *http.Request) string {
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		switch strings.TrimSpace(strings.Split(proto, ",")[0]) {
		case "http":
			return "http"
		case "https":
			return "https"
		}
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}
