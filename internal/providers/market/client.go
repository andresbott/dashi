package market

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"
)

const defaultBaseURL = "https://query1.finance.yahoo.com/v8/finance/chart"

type Client struct {
	httpClient *http.Client
	BaseURL    string
	cache      *cache
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		httpClient: httpClient,
		BaseURL:    defaultBaseURL,
		cache:      newCache(),
	}
}

func (c *Client) GetMarketData(symbol, rangeID string) (MarketData, error) {
	yahooRange, interval, ok := ValidRange(rangeID)
	if !ok {
		return MarketData{}, fmt.Errorf("invalid range: %q", rangeID)
	}

	if data, hit := c.cache.get(symbol, rangeID); hit {
		return data, nil
	}

	reqURL := fmt.Sprintf("%s/%s?range=%s&interval=%s", c.BaseURL, symbol, yahooRange, interval)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return MarketData{}, fmt.Errorf("market API request failed: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return MarketData{}, fmt.Errorf("market API request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return MarketData{}, fmt.Errorf("market API returned status %d", resp.StatusCode)
	}

	var raw yahooChartResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return MarketData{}, fmt.Errorf("decoding market response: %w", err)
	}

	data, err := parseYahooResponse(raw)
	if err != nil {
		return MarketData{}, err
	}

	c.cache.set(symbol, rangeID, data)
	return data, nil
}

// WarmupSymbols pre-fetches market data for the given symbols,
// populating the cache so subsequent requests are instant.
// It respects context cancellation between iterations.
func (c *Client) WarmupSymbols(ctx context.Context, symbols []struct{ Symbol, Range string }) {
	for _, s := range symbols {
		if ctx.Err() != nil {
			return
		}
		_, _ = c.GetMarketData(s.Symbol, s.Range)
	}
}

type yahooChartResponse struct {
	Chart struct {
		Result []yahooResult `json:"result"`
		Error  *yahooError   `json:"error"`
	} `json:"chart"`
}

type yahooError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type yahooResult struct {
	Meta       yahooMeta       `json:"meta"`
	Timestamps []int64         `json:"timestamp"`
	Indicators yahooIndicators `json:"indicators"`
}

type yahooMeta struct {
	Symbol             string  `json:"symbol"`
	ShortName          string  `json:"shortName"`
	Currency           string  `json:"currency"`
	RegularMarketPrice float64 `json:"regularMarketPrice"`
	ChartPreviousClose float64 `json:"chartPreviousClose"`
	RegularMarketState string  `json:"regularMarketState"`
}

type yahooIndicators struct {
	Quote []yahooQuote `json:"quote"`
}

type yahooQuote struct {
	Close []interface{} `json:"close"`
}

func parseYahooResponse(raw yahooChartResponse) (MarketData, error) {
	if raw.Chart.Error != nil {
		return MarketData{}, fmt.Errorf("yahoo API error: %s", raw.Chart.Error.Description)
	}
	if len(raw.Chart.Result) == 0 {
		return MarketData{}, fmt.Errorf("no results from yahoo API")
	}

	r := raw.Chart.Result[0]
	prevClose := r.Meta.ChartPreviousClose
	price := r.Meta.RegularMarketPrice
	change := price - prevClose
	changePct := 0.0
	if prevClose != 0 {
		changePct = (change / prevClose) * 100
	}

	quote := Quote{
		Symbol:        r.Meta.Symbol,
		Name:          r.Meta.ShortName,
		Currency:      r.Meta.Currency,
		Price:         price,
		Change:        math.Round(change*100) / 100,
		ChangePercent: math.Round(changePct*100) / 100,
		MarketState:   r.Meta.RegularMarketState,
	}

	var points []PricePoint
	closes := []float64{}
	if len(r.Indicators.Quote) > 0 {
		for _, v := range r.Indicators.Quote[0].Close {
			if v == nil {
				closes = append(closes, 0)
				continue
			}
			switch n := v.(type) {
			case float64:
				closes = append(closes, n)
			case json.Number:
				f, _ := n.Float64()
				closes = append(closes, f)
			default:
				closes = append(closes, 0)
			}
		}
	}

	for i, ts := range r.Timestamps {
		if i >= len(closes) {
			break
		}
		if closes[i] == 0 {
			continue
		}
		points = append(points, PricePoint{
			Time:  time.Unix(ts, 0).UTC(),
			Close: math.Round(closes[i]*100) / 100,
		})
	}

	return MarketData{Quote: quote, Points: points}, nil
}
