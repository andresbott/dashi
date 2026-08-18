package market

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetMarketData(t *testing.T) {
	response := map[string]any{
		"chart": map[string]any{
			"result": []any{
				map[string]any{
					"meta": map[string]any{
						"symbol":             "AAPL",
						"shortName":          "Apple Inc.",
						"currency":           "USD",
						"regularMarketPrice": 178.52,
						"chartPreviousClose": 176.18,
						"regularMarketState": "REGULAR",
					},
					"timestamp": []any{1712000000, 1712086400, 1712172800},
					"indicators": map[string]any{
						"quote": []any{
							map[string]any{
								"close": []any{174.5, 176.2, 178.52},
							},
						},
					},
				},
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("range") == "" {
			http.Error(w, "missing range", 400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("encode error: %v", err)
		}
	}))
	defer srv.Close()

	client := NewClient(nil)
	client.BaseURL = srv.URL

	data, err := client.GetMarketData("AAPL", "1mo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.Quote.Symbol != "AAPL" {
		t.Errorf("symbol = %q, want AAPL", data.Quote.Symbol)
	}
	if data.Quote.Price != 178.52 {
		t.Errorf("price = %f, want 178.52", data.Quote.Price)
	}
	if data.Quote.Name != "Apple Inc." {
		t.Errorf("name = %q, want Apple Inc.", data.Quote.Name)
	}
	if len(data.Points) != 3 {
		t.Fatalf("expected 3 points, got %d", len(data.Points))
	}
	if data.Points[2].Close != 178.52 {
		t.Errorf("last close = %f, want 178.52", data.Points[2].Close)
	}
}

func TestClient_GetMarketData_InvalidRange(t *testing.T) {
	client := NewClient(nil)
	_, err := client.GetMarketData("AAPL", "invalid")
	if err == nil {
		t.Fatal("expected error for invalid range")
	}
}

func TestClient_GetMarketData_Cache(t *testing.T) {
	callCount := 0
	response := map[string]any{
		"chart": map[string]any{
			"result": []any{
				map[string]any{
					"meta": map[string]any{
						"symbol":             "AAPL",
						"regularMarketPrice": 100.0,
						"chartPreviousClose": 99.0,
						"regularMarketState": "REGULAR",
					},
					"timestamp":  []any{},
					"indicators": map[string]any{"quote": []any{map[string]any{"close": []any{}}}},
				},
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatalf("encode error: %v", err)
		}
	}))
	defer srv.Close()

	client := NewClient(nil)
	client.BaseURL = srv.URL

	_, _ = client.GetMarketData("AAPL", "1mo")
	_, _ = client.GetMarketData("AAPL", "1mo")

	if callCount != 1 {
		t.Errorf("expected 1 API call (cached), got %d", callCount)
	}
}

func TestClient_GetMarketData_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", 500)
	}))
	defer srv.Close()

	client := NewClient(nil)
	client.BaseURL = srv.URL

	_, err := client.GetMarketData("AAPL", "1mo")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
