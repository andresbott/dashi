package swisstransport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCache_GetSet(t *testing.T) {
	c := newCache(1 * time.Minute)
	deps := []Departure{{Number: "11", Destination: "St-Louis"}}
	c.set("8500073", 5, deps)

	got, ok := c.get("8500073", 5)
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(got) != 1 || got[0].Number != "11" {
		t.Errorf("unexpected data: %+v", got)
	}
}

func TestCache_Miss(t *testing.T) {
	c := newCache(1 * time.Minute)
	_, ok := c.get("nonexistent", 5)
	if ok {
		t.Fatal("expected cache miss")
	}
}

func TestCache_Expiry(t *testing.T) {
	c := newCache(1 * time.Millisecond)
	deps := []Departure{{Number: "11", Destination: "St-Louis"}}
	c.set("8500073", 5, deps)
	time.Sleep(5 * time.Millisecond)

	_, ok := c.get("8500073", 5)
	if ok {
		t.Fatal("expected cache miss after expiry")
	}
}

func TestCache_DifferentKeys(t *testing.T) {
	c := newCache(1 * time.Minute)
	deps1 := []Departure{{Number: "11"}}
	deps2 := []Departure{{Number: "3"}}
	c.set("8500073", 5, deps1)
	c.set("8500073", 10, deps2)

	got1, ok := c.get("8500073", 5)
	if !ok || got1[0].Number != "11" {
		t.Errorf("wrong data for limit=5: %+v", got1)
	}
	got2, ok := c.get("8500073", 10)
	if !ok || got2[0].Number != "3" {
		t.Errorf("wrong data for limit=10: %+v", got2)
	}
}

func TestClient_SearchStations(t *testing.T) {
	response := map[string]any{
		"stations": []any{
			map[string]any{
				"id": "8500073", "name": "Basel, Aeschenplatz",
				"coordinate": map[string]any{"type": "WGS84", "x": 47.5513, "y": 7.594862},
			},
			map[string]any{
				"id": "8500010", "name": "Basel SBB",
				"coordinate": map[string]any{"type": "WGS84", "x": 47.547408, "y": 7.589566},
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/locations" {
			http.Error(w, "not found", 404)
			return
		}
		if r.URL.Query().Get("query") == "" {
			http.Error(w, "missing query", 400)
			return
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	client := NewClient(nil)
	client.baseURL = srv.URL

	stations, err := client.SearchStations("Basel")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stations) != 2 {
		t.Fatalf("expected 2 stations, got %d", len(stations))
	}
	if stations[0].ID != "8500073" {
		t.Errorf("id = %q, want 8500073", stations[0].ID)
	}
	if stations[0].Name != "Basel, Aeschenplatz" {
		t.Errorf("name = %q", stations[0].Name)
	}
	if stations[0].Latitude != 47.5513 {
		t.Errorf("latitude = %f, want 47.5513", stations[0].Latitude)
	}
}

func TestClient_GetDepartures(t *testing.T) {
	response := map[string]any{
		"station": map[string]any{"id": "8500073", "name": "Basel, Aeschenplatz"},
		"stationboard": []any{
			map[string]any{
				"category": "T", "number": "11", "to": "Basel, St-Louis Grenze",
				"stop": map[string]any{
					"departure": "2026-04-05T21:48:00+0200", "departureTimestamp": 1775418480,
					"delay": 2, "platform": "A",
					"prognosis": map[string]any{"departure": "2026-04-05T21:50:00+0200"},
				},
			},
			map[string]any{
				"category": "B", "number": "30", "to": "Basel, Badischer Bahnhof",
				"stop": map[string]any{
					"departure": "2026-04-05T21:55:00+0200", "departureTimestamp": 1775418900,
					"delay": 0, "platform": "B",
					"prognosis": map[string]any{"departure": nil},
				},
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/stationboard" {
			http.Error(w, "not found", 404)
			return
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	client := NewClient(nil)
	client.baseURL = srv.URL

	deps, err := client.GetDepartures("8500073", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deps) != 2 {
		t.Fatalf("expected 2 departures, got %d", len(deps))
	}
	if deps[0].Category != "T" {
		t.Errorf("category = %q, want T", deps[0].Category)
	}
	if deps[0].Number != "11" {
		t.Errorf("number = %q, want 11", deps[0].Number)
	}
	if deps[0].Destination != "Basel, St-Louis Grenze" {
		t.Errorf("destination = %q", deps[0].Destination)
	}
	if deps[0].Delay != 2 {
		t.Errorf("delay = %d, want 2", deps[0].Delay)
	}
	if deps[0].Platform != "A" {
		t.Errorf("platform = %q, want A", deps[0].Platform)
	}
	if deps[0].Expected.IsZero() {
		t.Error("expected time should not be zero")
	}
	// Second departure has no prognosis, so Expected should equal Scheduled
	if !deps[1].Expected.Equal(deps[1].Scheduled) {
		t.Errorf("expected = scheduled when no prognosis, got expected=%v scheduled=%v", deps[1].Expected, deps[1].Scheduled)
	}
}

func TestClient_GetDepartures_Cache(t *testing.T) {
	callCount := 0
	response := map[string]any{
		"station": map[string]any{"id": "8500073", "name": "Test"},
		"stationboard": []any{},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	client := NewClient(nil)
	client.baseURL = srv.URL

	_, _ = client.GetDepartures("8500073", 5)
	_, _ = client.GetDepartures("8500073", 5)

	if callCount != 1 {
		t.Errorf("expected 1 API call (cached), got %d", callCount)
	}
}

func TestClient_GetDepartures_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", 500)
	}))
	defer srv.Close()

	client := NewClient(nil)
	client.baseURL = srv.URL

	_, err := client.GetDepartures("8500073", 5)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestClient_SearchStations_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", 500)
	}))
	defer srv.Close()

	client := NewClient(nil)
	client.baseURL = srv.URL

	_, err := client.SearchStations("Basel")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
