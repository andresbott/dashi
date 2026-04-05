package market

import (
	"sync"
	"time"
)

type cacheEntry struct {
	data   MarketData
	expiry time.Time
}

type cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
}

func newCache() *cache {
	return &cache{entries: make(map[string]cacheEntry)}
}

func cacheKey(symbol, rangeID string) string {
	return symbol + ":" + rangeID
}

func cacheTTL(rangeID string) time.Duration {
	switch rangeID {
	case "1d":
		return 15 * time.Minute
	case "5d":
		return 1 * time.Hour
	default:
		switch rangeID {
		case "1mo", "3mo", "6mo", "1y":
			return 24 * time.Hour
		default:
			return 15 * time.Minute
		}
	}
}

func (c *cache) get(symbol, rangeID string) (MarketData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	key := cacheKey(symbol, rangeID)
	entry, ok := c.entries[key]
	if !ok || time.Now().After(entry.expiry) {
		return MarketData{}, false
	}
	return entry.data, true
}

func (c *cache) set(symbol, rangeID string, data MarketData) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := cacheKey(symbol, rangeID)
	c.entries[key] = cacheEntry{
		data:   data,
		expiry: time.Now().Add(cacheTTL(rangeID)),
	}
}
