package spa

import (
	"bytes"
	"encoding/json"
	"html"
	"net/http"
	"strings"
	"sync"
)

// ingressBase turns HA's X-Ingress-Path header into a base path that always
// ends in "/". Empty or malformed values yield "/", so non-ingress serving is
// unchanged and a forged header cannot break out of the href / JS string.
func ingressBase(r *http.Request) string {
	p := r.Header.Get("X-Ingress-Path")
	if p == "" || !strings.HasPrefix(p, "/") || !isSafeIngressPath(p) {
		return "/"
	}
	return strings.TrimSuffix(p, "/") + "/"
}

func isSafeIngressPath(p string) bool {
	for _, c := range p {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		case c == '/', c == '-', c == '_', c == '.':
		default:
			return false
		}
	}
	return true
}

// injectBase inserts a <base href> and window.__DASHI_BASE__ (both derived from
// base, which must end in "/") immediately after the shell's <head>, so the SPA
// resolves assets, routes and API calls under Home Assistant's ingress prefix.
// It is placed first in <head> so it precedes any asset link.
func injectBase(raw []byte, base string) []byte {
	js, _ := json.Marshal(base)
	head := "<head>\n" +
		`<base href="` + html.EscapeString(base) + `">` + "\n" +
		`<script>window.__DASHI_BASE__=` + string(js) + `;</script>`
	return bytes.Replace(raw, []byte("<head>"), []byte(head), 1)
}

// EditorHandler serves real embedded files verbatim and every other path as the
// SPA shell with the ingress base injected (from X-Ingress-Path). This lets the
// editor run under Home Assistant's dynamic ingress prefix; direct/dev access
// gets base "/" and behaves exactly as before.
//
// The shell is loaded lazily, so construction never fails when the UI has not
// been embedded yet (e.g. `make run` without `make package-ui`, or the CI test
// job) — in that case requests fall back to the plain asset handler.
func EditorHandler() (http.HandlerFunc, error) {
	assets, err := App("/")
	if err != nil {
		return nil, err
	}
	var (
		once sync.Once
		raw  []byte
	)
	shell := func() []byte {
		once.Do(func() {
			if b, e := IndexHTML(); e == nil {
				raw = b
			}
		})
		return raw
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if name := strings.TrimPrefix(r.URL.Path, "/"); name != "" && name != "index.html" && FileExists(name) {
			assets.ServeHTTP(w, r)
			return
		}
		idx := shell()
		if idx == nil {
			// UI not embedded — behave as before this handler existed.
			assets.ServeHTTP(w, r)
			return
		}
		out := injectBase(idx, ingressBase(r))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(out)
	}, nil
}
