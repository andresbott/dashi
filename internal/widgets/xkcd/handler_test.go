package xkcd

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	xkcdclient "github.com/andresbott/dashi/internal/providers/xkcd"
)

func setupXkcdTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"num":        2500,
			"safe_title": "Test Comic",
			"title":      "Test Comic",
			"img":        "https://example.com/comic.png",
			"alt":        "alt text",
			"day":        "1",
			"month":      "1",
			"year":       "2026",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func TestHandler_GetComic_Latest(t *testing.T) {
	srv := setupXkcdTestServer()
	defer srv.Close()

	tmp := t.TempDir()
	client := xkcdclient.NewClient(tmp)
	client.SetBaseURL(srv.URL)
	h := newHandler(client, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/api/v0/widgets/xkcd", nil)
	rec := httptest.NewRecorder()
	h.GetComic(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var comic xkcdclient.Comic
	if err := json.NewDecoder(rec.Body).Decode(&comic); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if comic.Num != 2500 {
		t.Errorf("num = %d, want 2500", comic.Num)
	}
}

func TestHandler_GetComic_RandomModes(t *testing.T) {
	srv := setupXkcdTestServer()
	defer srv.Close()

	for _, mode := range []string{"random", "random-each"} {
		t.Run(mode, func(t *testing.T) {
			tmp := t.TempDir()
			client := xkcdclient.NewClient(tmp)
			client.SetBaseURL(srv.URL)
			h := newHandler(client, slog.Default())

			req := httptest.NewRequest(http.MethodGet, "/api/v0/widgets/xkcd?mode="+mode, nil)
			rec := httptest.NewRecorder()
			h.GetComic(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
		})
	}
}

func TestHandler_GetComic_UpstreamError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	tmp := t.TempDir()
	client := xkcdclient.NewClient(tmp)
	client.SetBaseURL(srv.URL)
	h := newHandler(client, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/api/v0/widgets/xkcd", nil)
	rec := httptest.NewRecorder()
	h.GetComic(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}
