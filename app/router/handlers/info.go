package handlers

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// PublicViewer describes where the public, server-rendered dashboard viewer is
// reachable from a browser.
type PublicViewer struct {
	Enabled bool
	Port    int
	// BaseURL is the explicit public URL, already normalized at config load
	// (http/https only, no trailing slash); it is used verbatim and wins over
	// deriving one from the request.
	BaseURL string
}

// PublicEditor describes where the editor is reachable from a browser. URL is
// the configured, already normalized public URL; it is never derived from the
// request, because behind an ingress (HA add-on) the request cannot reveal it.
type PublicEditor struct {
	Enabled bool
	URL     string
}

// InfoHandler serves the runtime facts the admin SPA needs at init time. The
// SPA is the editor UI only — it does not render dashboards — so its "view"
// links must point at the viewer server, which lives on a different port and,
// behind a reverse proxy, possibly on a different host.
type InfoHandler struct {
	viewer PublicViewer
	editor PublicEditor
}

func NewInfoHandler(viewer PublicViewer, editor PublicEditor) *InfoHandler {
	return &InfoHandler{viewer: viewer, editor: editor}
}

type serverInfo struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
}

type infoResponse struct {
	Viewer serverInfo `json:"viewer"`
	Editor serverInfo `json:"editor"`
}

func (h *InfoHandler) Get(w http.ResponseWriter, r *http.Request) {
	resp := infoResponse{
		Viewer: serverInfo{
			Enabled: h.viewer.Enabled,
			URL:     h.viewerURL(r),
		},
		Editor: serverInfo{
			Enabled: h.editor.Enabled,
			URL:     h.editorURL(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		ErrorJSON(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// viewerURL returns the browser-reachable base URL of the viewer server,
// without a trailing slash. An explicit BaseURL wins verbatim; otherwise the URL is
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
		return h.viewer.BaseURL
	}
	host := requestHost(r)
	if host == "" || h.viewer.Port <= 0 {
		return ""
	}
	return requestScheme(r) + "://" + net.JoinHostPort(host, strconv.Itoa(h.viewer.Port))
}

// editorURL returns the editor's public base URL, or "" when the editor is
// disabled or no public URL is configured.
func (h *InfoHandler) editorURL() string {
	if !h.editor.Enabled {
		return ""
	}
	return h.editor.URL
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
