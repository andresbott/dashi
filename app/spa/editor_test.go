package spa

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIngressBase(t *testing.T) {
	cases := []struct {
		name, header, want string
	}{
		{"no header", "", "/"},
		{"simple prefix", "/api/hassio_ingress/abc123", "/api/hassio_ingress/abc123/"},
		{"trailing slash normalized", "/api/hassio_ingress/abc123/", "/api/hassio_ingress/abc123/"},
		{"missing leading slash rejected", "api/hassio_ingress/x", "/"},
		{"unsafe chars rejected", `/x"><script>`, "/"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if c.header != "" {
				r.Header.Set("X-Ingress-Path", c.header)
			}
			if got := ingressBase(r); got != c.want {
				t.Fatalf("ingressBase(%q) = %q, want %q", c.header, got, c.want)
			}
		})
	}
}

// TestInjectBase covers the injection logic without depending on the embedded
// SPA (which is a build artifact absent on a fresh CI checkout).
func TestInjectBase(t *testing.T) {
	raw := []byte("<html><head><title>x</title></head><body></body></html>")

	out := string(injectBase(raw, "/"))
	if !strings.Contains(out, `<base href="/">`) {
		t.Errorf("missing default <base href>, got:\n%s", out)
	}
	if !strings.Contains(out, `window.__DASHI_BASE__="/"`) {
		t.Errorf("missing default __DASHI_BASE__, got:\n%s", out)
	}

	out = string(injectBase(raw, "/api/hassio_ingress/abc123/"))
	if !strings.Contains(out, `<base href="/api/hassio_ingress/abc123/">`) {
		t.Errorf("missing injected <base href>, got:\n%s", out)
	}
	// The <base> must precede other head content so relative assets resolve.
	if strings.Index(out, "<base") > strings.Index(out, "<title>") {
		t.Errorf("<base> must come before <title>, got:\n%s", out)
	}
}

// TestEditorHandlerServesInjectedShell is an end-to-end check that runs only
// when the SPA has been embedded (locally / release builds). On the CI test job
// the UI is not built, so it skips — the injection logic itself is covered by
// TestInjectBase.
func TestEditorHandlerServesInjectedShell(t *testing.T) {
	if !FileExists("index.html") {
		t.Skip("SPA not embedded (run `make package-ui`); injection covered by TestInjectBase")
	}
	h, err := EditorHandler()
	if err != nil {
		t.Fatalf("EditorHandler: %v", err)
	}

	// Shell under ingress → injected prefix.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("X-Ingress-Path", "/api/hassio_ingress/abc123")
	h(rec, req)
	if !strings.Contains(rec.Body.String(), `<base href="/api/hassio_ingress/abc123/">`) {
		t.Errorf("shell missing injected base:\n%s", rec.Body.String())
	}

	// Real embedded file (favicon) served verbatim, not the shell.
	rec = httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, "/favicon.svg", nil))
	if strings.Contains(rec.Body.String(), "window.__DASHI_BASE__") {
		t.Errorf("favicon.svg should be served as a file, not the injected shell")
	}
}
