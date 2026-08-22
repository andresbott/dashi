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
			name:    "explicit public url wins and loses trailing slash",
			viewer:  PublicViewer{Enabled: true, Port: 8087, BaseURL: "https://dash.example.com/"},
			host:    "localhost:8088",
			wantOn:  true,
			wantURL: "https://dash.example.com",
		},
		{
			name:    "non http public url is dropped",
			viewer:  PublicViewer{Enabled: true, Port: 8087, BaseURL: "javascript:alert(1)"},
			host:    "localhost:8088",
			wantOn:  true,
			wantURL: "",
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

			rec := httptest.NewRecorder()
			NewInfoHandler(tc.viewer).Get(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			var got infoResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if got.Viewer.Enabled != tc.wantOn {
				t.Errorf("viewer.enabled = %v, want %v", got.Viewer.Enabled, tc.wantOn)
			}
			if got.Viewer.URL != tc.wantURL {
				t.Errorf("viewer.url = %q, want %q", got.Viewer.URL, tc.wantURL)
			}
		})
	}
}
