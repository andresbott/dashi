package weather

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type cacheEntry struct {
	data   WeatherData
	expiry time.Time
}

type cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
}

func newCache(ttl time.Duration) *cache {
	return &cache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
	}
}

func cacheKey(lat, lon float64) string {
	rlat := math.Round(lat*100) / 100
	rlon := math.Round(lon*100) / 100
	return fmt.Sprintf("%.2f,%.2f", rlat, rlon)
}

func (c *cache) get(lat, lon float64) (WeatherData, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey(lat, lon)
	entry, ok := c.entries[key]
	if !ok {
		return WeatherData{}, false
	}
	if time.Now().After(entry.expiry) {
		delete(c.entries, key)
		return WeatherData{}, false
	}
	return entry.data, true
}

func (c *cache) set(lat, lon float64, data WeatherData) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey(lat, lon)
	c.entries[key] = cacheEntry{
		data:   data,
		expiry: time.Now().Add(c.ttl),
	}
}
