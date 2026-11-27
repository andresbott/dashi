package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInfoHandlerViewerURL(t *testing.T) {
	tests := []struct {
		name    string
		viewer  PublicViewer
		host    string
		headers map[string]string
		tls     bool
		wantOn  bool
		wantURL string
	}{
		{
			name:    "derived from request host",
			viewer:  PublicViewer{Enabled: true, Port: 8087},
			host:    "localhost:8088",
			wantOn:  true,
			wantURL: "http://localhost:8087",
		},
		{
			name:    "host without port",
			viewer:  PublicViewer{Enabled: true, Port: 8087},
			host:    "dashi.example.com",
			wantOn:  true,
			wantURL: "http://dashi.example.com:8087",
		},
		{
			name:    "ipv6 host",
			viewer:  PublicViewer{Enabled: true, Port: 8087},
			host:    "[::1]:8088",
			wantOn:  true,
			wantURL: "http://[::1]:8087",
		},
		{
			name:    "explicit public url wins",
			viewer:  PublicViewer{Enabled: true, Port: 8087, BaseURL: "https://dash.example.com"},
			host:    "localhost:8088",
			wantOn:  true,
			wantURL: "https://dash.example.com",
		},
		{
			// Normalization happens once at config load (app/cmd/config.go).
			name:    "explicit public url used verbatim",
			viewer:  PublicViewer{Enabled: true, Port: 8087, BaseURL: "https://ha.example.com/dash/"},
			host:    "localhost:8088",
			wantOn:  true,
			wantURL: "https://ha.example.com/dash/",
		},
		{
			name:    "forwarded proto and host",
			viewer:  PublicViewer{Enabled: true, Port: 8087},
			host:    "internal:8088",
			headers: map[string]string{"X-Forwarded-Proto": "https", "X-Forwarded-Host": "dashi.example.com"},
			wantOn:  true,
			wantURL: "https://dashi.example.com:8087",
		},
		{
			name:    "forged host header is rejected",
			viewer:  PublicViewer{Enabled: true, Port: 8087},
			host:    "localhost:8088",
			headers: map[string]string{"X-Forwarded-Host": "evil.com/javascript:alert(1)"},
			wantOn:  true,
			wantURL: "",
		},
		{
			name:    "tls request derives https",
			viewer:  PublicViewer{Enabled: true, Port: 8087},
			host:    "dashi.example.com:8088",
			tls:     true,
			wantOn:  true,
			wantURL: "https://dashi.example.com:8087",
		},
		{
			name:    "disabled viewer reports no url",
			viewer:  PublicViewer{Enabled: false, Port: 8087},
			host:    "localhost:8088",
			wantOn:  false,
			wantURL: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			scheme := "http://"
			if tc.tls {
				scheme = "https://"
			}
			req := httptest.NewRequest(http.MethodGet, scheme+"example.invalid/api/v0/info", nil)
			req.Host = tc.host
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			got := getInfo(t, NewInfoHandler(tc.viewer, PublicEditor{}), req)
			if got.Viewer.Enabled != tc.wantOn {
				t.Errorf("viewer.enabled = %v, want %v", got.Viewer.Enabled, tc.wantOn)
			}
			if got.Viewer.URL != tc.wantURL {
				t.Errorf("viewer.url = %q, want %q", got.Viewer.URL, tc.wantURL)
			}
		})
	}
}

func TestInfoHandlerEditor(t *testing.T) {
	tests := []struct {
		name    string
		editor  PublicEditor
		wantOn  bool
		wantURL string
	}{
		{
			name:    "configured url used verbatim",
			editor:  PublicEditor{Enabled: true, URL: "https://ha.example.com/hassio/ingress/dashi"},
			wantOn:  true,
			wantURL: "https://ha.example.com/hassio/ingress/dashi",
		},
		{
			// Never derived from the request: the editor may sit behind an
			// ingress whose URL the request cannot reveal.
			name:    "enabled editor without url reports none",
			editor:  PublicEditor{Enabled: true},
			wantOn:  true,
			wantURL: "",
		},
		{
			name:    "disabled editor reports no url",
			editor:  PublicEditor{Enabled: false, URL: "https://edit.example.com"},
			wantOn:  false,
			wantURL: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://localhost:8088/api/v0/info", nil)
			got := getInfo(t, NewInfoHandler(PublicViewer{Enabled: true, Port: 8087}, tc.editor), req)
			if got.Editor.Enabled != tc.wantOn {
				t.Errorf("editor.enabled = %v, want %v", got.Editor.Enabled, tc.wantOn)
			}
			if got.Editor.URL != tc.wantURL {
				t.Errorf("editor.url = %q, want %q", got.Editor.URL, tc.wantURL)
			}
		})
	}
}

func getInfo(t *testing.T, h *InfoHandler, req *http.Request) infoResponse {
	t.Helper()
	rec := httptest.NewRecorder()
	h.Get(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got infoResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return got
}
