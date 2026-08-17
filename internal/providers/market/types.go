package market

import "time"

type Quote struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Currency      string  `json:"currency"`
	Price         float64 `json:"price"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"changePercent"`
	MarketState   string  `json:"marketState"`
}

type PricePoint struct {
	Time  time.Time `json:"time"`
	Close float64   `json:"close"`
}

type MarketData struct {
	Quote  Quote        `json:"quote"`
	Points []PricePoint `json:"points"`
}

var validRanges = map[string]struct {
	Range    string
	Interval string
}{
	"1d":  {Range: "1d", Interval: "5m"},
	"5d":  {Range: "5d", Interval: "30m"},
	"1mo": {Range: "1mo", Interval: "1d"},
	"3mo": {Range: "3mo", Interval: "1d"},
	"6mo": {Range: "6mo", Interval: "1d"},
	"1y":  {Range: "1y", Interval: "1wk"},
}

func ValidRange(rangeID string) (yahooRange, interval string, ok bool) {
	r, found := validRanges[rangeID]
	if !found {
		return "", "", false
	}
	return r.Range, r.Interval, true
}
