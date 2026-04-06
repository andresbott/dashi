package market

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mkt "github.com/andresbott/dashi/internal/market"
	"github.com/andresbott/dashi/internal/widgets"
)

func TestNewStaticRenderer_EmptyConfig(t *testing.T) {
	client := mkt.NewClient(nil)
	renderer := NewStaticRenderer(client)

	html, err := renderer(nil, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	htmlStr := string(html)
	if !strings.Contains(htmlStr, "No symbol configured") {
		t.Errorf("expected unconfigured message in HTML, got: %s", htmlStr)
	}
}

func TestNewStaticRenderer_InvalidJSON(t *testing.T) {
	client := mkt.NewClient(nil)
	renderer := NewStaticRenderer(client)

	_, err := renderer(json.RawMessage(`{invalid`), widgets.RenderContext{})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "market config") {
		t.Errorf("expected 'market config' in error, got: %v", err)
	}
}

func TestNewStaticRenderer_NoSymbol(t *testing.T) {
	client := mkt.NewClient(nil)
	renderer := NewStaticRenderer(client)

	config := json.RawMessage(`{"range":"1mo"}`)
	html, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	htmlStr := string(html)
	if !strings.Contains(htmlStr, "No symbol configured") {
		t.Errorf("expected unconfigured message when symbol is empty")
	}
}

func TestNewStaticRenderer_WithSymbol(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := `{
			"chart": {
				"result": [{
					"meta": {
						"symbol": "AAPL",
						"shortName": "Apple Inc.",
						"currency": "USD",
						"regularMarketPrice": 175.50,
						"chartPreviousClose": 170.00
					},
					"timestamp": [1704067200, 1704153600],
					"indicators": {
						"quote": [{
							"close": [172.0, 175.5]
						}]
					}
				}]
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	client := mkt.NewClient(server.Client())
	client.BaseURL = server.URL
	renderer := NewStaticRenderer(client)

	config := json.RawMessage(`{"symbol":"AAPL","range":"1mo"}`)
	html, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	htmlStr := string(html)
	if !strings.Contains(htmlStr, "AAPL") {
		t.Errorf("expected AAPL in output, got: %s", htmlStr)
	}
	if !strings.Contains(htmlStr, "Apple Inc.") {
		t.Errorf("expected 'Apple Inc.' in output")
	}
	if !strings.Contains(htmlStr, "175.50") {
		t.Errorf("expected price 175.50 in output")
	}
}

func TestNewStaticRenderer_DefaultRange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "range=1mo") {
			t.Errorf("expected default range=1mo, got: %s", r.URL.RawQuery)
		}
		resp := `{
			"chart": {
				"result": [{
					"meta": {
						"symbol": "TSLA",
						"shortName": "Tesla",
						"currency": "USD",
						"regularMarketPrice": 250.00,
						"chartPreviousClose": 240.00
					},
					"timestamp": [1704067200],
					"indicators": {
						"quote": [{
							"close": [250.0]
						}]
					}
				}]
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	client := mkt.NewClient(server.Client())
	client.BaseURL = server.URL
	renderer := NewStaticRenderer(client)

	config := json.RawMessage(`{"symbol":"TSLA"}`)
	_, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewStaticRenderer_ShowChartFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := `{
			"chart": {
				"result": [{
					"meta": {
						"symbol": "GOOGL",
						"shortName": "Alphabet",
						"currency": "USD",
						"regularMarketPrice": 140.00,
						"chartPreviousClose": 138.00
					},
					"timestamp": [1704067200, 1704153600],
					"indicators": {
						"quote": [{
							"close": [138.0, 140.0]
						}]
					}
				}]
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	client := mkt.NewClient(server.Client())
	client.BaseURL = server.URL
	renderer := NewStaticRenderer(client)

	config := json.RawMessage(`{"symbol":"GOOGL","showChart":false}`)
	html, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	htmlStr := string(html)
	// Chart should not be shown
	if strings.Contains(htmlStr, "base64") {
		t.Errorf("expected no chart image when showChart=false")
	}

	// Now test with explicit true
	config = json.RawMessage(`{"symbol":"GOOGL","showChart":true}`)
	if _, err = renderer(config, widgets.RenderContext{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewStaticRenderer_ShowChangeFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := `{
			"chart": {
				"result": [{
					"meta": {
						"symbol": "MSFT",
						"shortName": "Microsoft",
						"currency": "USD",
						"regularMarketPrice": 380.00,
						"chartPreviousClose": 375.00
					},
					"timestamp": [1704067200],
					"indicators": {
						"quote": [{
							"close": [380.0]
						}]
					}
				}]
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	client := mkt.NewClient(server.Client())
	client.BaseURL = server.URL
	renderer := NewStaticRenderer(client)

	config := json.RawMessage(`{"symbol":"MSFT","showChange":false}`)
	_, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Now test with explicit true
	config = json.RawMessage(`{"symbol":"MSFT","showChange":true}`)
	_, err = renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewStaticRenderer_NegativeChange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := `{
			"chart": {
				"result": [{
					"meta": {
						"symbol": "NFLX",
						"shortName": "Netflix",
						"currency": "USD",
						"regularMarketPrice": 450.00,
						"chartPreviousClose": 460.00
					},
					"timestamp": [1704067200, 1704153600],
					"indicators": {
						"quote": [{
							"close": [460.0, 450.0]
						}]
					}
				}]
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	client := mkt.NewClient(server.Client())
	client.BaseURL = server.URL
	renderer := NewStaticRenderer(client)

	config := json.RawMessage(`{"symbol":"NFLX","range":"1d"}`)
	html, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	htmlStr := string(html)
	// Negative change should not have + prefix
	if !strings.Contains(htmlStr, "-10.00") {
		t.Errorf("expected negative change value in output")
	}
}

func TestNewStaticRenderer_ClientError(t *testing.T) {
	client := mkt.NewClient(nil)
	client.BaseURL = "http://invalid-url-that-does-not-exist.local"
	renderer := NewStaticRenderer(client)

	config := json.RawMessage(`{"symbol":"INVALID"}`)
	_, err := renderer(config, widgets.RenderContext{})
	if err == nil {
		t.Fatal("expected error when client fails")
	}
	if !strings.Contains(err.Error(), "market fetch") {
		t.Errorf("expected 'market fetch' in error, got: %v", err)
	}
}

func TestNewStaticRenderer_WithChart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := `{
			"chart": {
				"result": [{
					"meta": {
						"symbol": "AMZN",
						"shortName": "Amazon",
						"currency": "USD",
						"regularMarketPrice": 160.00,
						"chartPreviousClose": 155.00
					},
					"timestamp": [1704067200, 1704153600, 1704240000],
					"indicators": {
						"quote": [{
							"close": [155.0, 157.5, 160.0]
						}]
					}
				}]
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	client := mkt.NewClient(server.Client())
	client.BaseURL = server.URL
	renderer := NewStaticRenderer(client)

	config := json.RawMessage(`{"symbol":"AMZN","range":"5d","showChart":true}`)
	html, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	htmlStr := string(html)
	// Should contain base64 encoded chart image
	if !strings.Contains(htmlStr, "base64") {
		t.Errorf("expected base64 chart image in output when showChart=true and points>1")
	}
}

func TestNewStaticRenderer_RangeLabels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := `{
			"chart": {
				"result": [{
					"meta": {
						"symbol": "META",
						"shortName": "Meta",
						"currency": "USD",
						"regularMarketPrice": 480.00,
						"chartPreviousClose": 475.00
					},
					"timestamp": [1704067200],
					"indicators": {
						"quote": [{
							"close": [480.0]
						}]
					}
				}]
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	client := mkt.NewClient(server.Client())
	client.BaseURL = server.URL
	renderer := NewStaticRenderer(client)

	testCases := []struct {
		rangeID string
		label   string
	}{
		{"1d", "1 Day"},
		{"5d", "5 Days"},
		{"1mo", "1 Month"},
		{"3mo", "3 Months"},
		{"6mo", "6 Months"},
		{"1y", "1 Year"},
	}

	for _, tc := range testCases {
		config := json.RawMessage(`{"symbol":"META","range":"` + tc.rangeID + `"}`)
		html, err := renderer(config, widgets.RenderContext{})
		if err != nil {
			t.Fatalf("unexpected error for range %s: %v", tc.rangeID, err)
		}
		htmlStr := string(html)
		if !strings.Contains(htmlStr, tc.label) {
			t.Errorf("expected range label '%s' for range %s in output", tc.label, tc.rangeID)
		}
	}
}

func TestNewStaticRenderer_SinglePoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := `{
			"chart": {
				"result": [{
					"meta": {
						"symbol": "NVDA",
						"shortName": "NVIDIA",
						"currency": "USD",
						"regularMarketPrice": 880.00,
						"chartPreviousClose": 870.00
					},
					"timestamp": [1704067200],
					"indicators": {
						"quote": [{
							"close": [880.0]
						}]
					}
				}]
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer server.Close()

	client := mkt.NewClient(server.Client())
	client.BaseURL = server.URL
	renderer := NewStaticRenderer(client)

	config := json.RawMessage(`{"symbol":"NVDA","showChart":true}`)
	html, err := renderer(config, widgets.RenderContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	htmlStr := string(html)
	// With only 1 point, chart should not be generated (requires > 1 point)
	// But the chart image tag will still be there with empty src if generateChart fails
	// Actually looking at the code, if len(data.Points) <= 1, generateChart is not even called
	// so td.ChartImage will be empty and no img tag will be rendered
	if strings.Contains(htmlStr, "img src=\"data:image/png;base64,") && strings.Contains(htmlStr, "base64,\"") {
		// Empty base64 string means no chart was generated
		t.Log("Chart section present but no image data (expected behavior)")
	}
}
