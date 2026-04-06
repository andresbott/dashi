package swisstransport

import (
	"fmt"
	"sync"
	"time"
)

type cacheEntry struct {
	data   []Departure
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

func cacheKey(stationID string, limit int) string {
	return fmt.Sprintf("%s:%d", stationID, limit)
}

func (c *cache) get(stationID string, limit int) ([]Departure, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey(stationID, limit)
	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expiry) {
		delete(c.entries, key)
		return nil, false
	}
	return entry.data, true
}

func (c *cache) set(stationID string, limit int, data []Departure) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey(stationID, limit)
	c.entries[key] = cacheEntry{
		data:   data,
		expiry: time.Now().Add(c.ttl),
	}
}
