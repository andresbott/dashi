package swisstransport

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	swisstransportpkg "github.com/andresbott/dashi/internal/providers/swisstransport"
)

func setupTransportTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch path {
		case "/v1/locations":
			resp := map[string]any{
				"stations": []map[string]any{
					{
						"id":   "8501120",
						"name": "Lausanne",
						"coordinate": map[string]any{
							"type": "WGS84",
							"x":    46.516667,
							"y":    6.629167,
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		case "/v1/stationboard":
			resp := map[string]any{
				"stationboard": []map[string]any{
					{
						"name":     "IC5",
						"category": "IC",
						"number":   "5",
						"to":       "Geneva",
						"stop": map[string]any{
							"departure":          "2026-01-01T12:00:00+0000",
							"departureTimestamp": int64(1735732800),
							"delay":              0,
							"platform":           "3",
							"prognosis": map[string]any{
								"departure": nil,
							},
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestHandler_SearchStations_MissingQuery(t *testing.T) {
	h := newHandler(swisstransportpkg.NewClient(nil), slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/api/v0/widgets/transport/stations", nil)
	rec := httptest.NewRecorder()
	h.SearchStations(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHandler_GetDepartures_MissingID(t *testing.T) {
	h := newHandler(swisstransportpkg.NewClient(nil), slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/api/v0/widgets/transport/stationboard", nil)
	rec := httptest.NewRecorder()
	h.GetDepartures(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHandler_SearchStations_OK(t *testing.T) {
	srv := setupTransportTestServer()
	defer srv.Close()

	client := swisstransportpkg.NewClient(nil)
	client.SetBaseURL(srv.URL)
	h := newHandler(client, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/api/v0/widgets/transport/stations?query=Lausanne", nil)
	rec := httptest.NewRecorder()
	h.SearchStations(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
}

func TestHandler_GetDepartures_OK(t *testing.T) {
	srv := setupTransportTestServer()
	defer srv.Close()

	client := swisstransportpkg.NewClient(nil)
	client.SetBaseURL(srv.URL)
	h := newHandler(client, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/api/v0/widgets/transport/stationboard?id=8501120&limit=3", nil)
	rec := httptest.NewRecorder()
	h.GetDepartures(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", rec.Code, rec.Body.String())
	}
}
