package market

import (
	"testing"
	"time"
)

func TestCache_GetMiss(t *testing.T) {
	c := newCache()
	_, ok := c.get("AAPL", "1mo")
	if ok {
		t.Fatal("expected cache miss")
	}
}

func TestCache_SetAndGet(t *testing.T) {
	c := newCache()
	data := MarketData{Quote: Quote{Symbol: "AAPL", Price: 150.0}}
	c.set("AAPL", "1mo", data)
	got, ok := c.get("AAPL", "1mo")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if got.Quote.Price != 150.0 {
		t.Fatalf("expected price 150.0, got %f", got.Quote.Price)
	}
}

func TestCache_DifferentRanges(t *testing.T) {
	c := newCache()
	c.set("AAPL", "1d", MarketData{Quote: Quote{Price: 100}})
	c.set("AAPL", "1mo", MarketData{Quote: Quote{Price: 200}})
	got1d, _ := c.get("AAPL", "1d")
	got1mo, _ := c.get("AAPL", "1mo")
	if got1d.Quote.Price != 100 || got1mo.Quote.Price != 200 {
		t.Fatal("ranges should be independent cache keys")
	}
}

func TestCache_Expiry(t *testing.T) {
	c := newCache()
	c.entries[cacheKey("AAPL", "1mo")] = cacheEntry{
		data:   MarketData{Quote: Quote{Price: 1}},
		expiry: time.Now().Add(-1 * time.Second),
	}
	_, ok := c.get("AAPL", "1mo")
	if ok {
		t.Fatal("expected expired entry to be a miss")
	}
}

func TestCacheTTL(t *testing.T) {
	if cacheTTL("1d") != 15*time.Minute {
		t.Error("1d should be 15 minutes")
	}
	if cacheTTL("5d") != 1*time.Hour {
		t.Error("5d should be 1 hour")
	}
	if cacheTTL("1mo") != 24*time.Hour {
		t.Error("1mo should be 24 hours")
	}
	if cacheTTL("unknown") != 15*time.Minute {
		t.Error("unknown should default to 15 minutes")
	}
}
