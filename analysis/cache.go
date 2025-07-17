package analysis

import (
	"sync"
	"time"
)

// AnalysisCache provides an interface for caching analysis results
type AnalysisCache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
	Stats() CacheStats
}

// CacheStats provides metrics about cache performance
type CacheStats struct {
	Hits       int64
	Misses     int64
	Size       int
	HitRate    float64
	MemoryUsed int64
}

// inMemoryCache is a thread-safe in-memory implementation of AnalysisCache
type inMemoryCache struct {
	mu       sync.RWMutex
	items    map[string]cacheItem
	stats    CacheStats
	maxSize  int
	ttl      time.Duration
}

// cacheItem represents a single cached item with expiration
type cacheItem struct {
	value     interface{}
	expiresAt time.Time
	accessedAt time.Time
}

// NewAnalysisCache creates a new in-memory cache with specified max size and default TTL
func NewAnalysisCache(maxSize int, defaultTTL time.Duration) AnalysisCache {
	cache := &inMemoryCache{
		items:   make(map[string]cacheItem),
		maxSize: maxSize,
		ttl:     defaultTTL,
	}
	
	// Start cleanup goroutine
	go cache.cleanup()
	
	return cache
}

// Get retrieves an item from the cache
func (c *inMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	item, exists := c.items[key]
	if !exists {
		c.stats.Misses++
		return nil, false
	}
	
	if time.Now().After(item.expiresAt) {
		delete(c.items, key)
		c.stats.Size = len(c.items)
		c.stats.Misses++
		return nil, false
	}
	
	// Update access time
	item.accessedAt = time.Now()
	c.items[key] = item
	
	c.stats.Hits++
	c.stats.HitRate = float64(c.stats.Hits) / float64(c.stats.Hits + c.stats.Misses)
	
	return item.value, true
}

// Set adds or updates an item in the cache
func (c *inMemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if ttl == 0 {
		ttl = c.ttl
	}
	
	// Evict if at max size
	if len(c.items) >= c.maxSize {
		c.evictLRU()
	}
	
	c.items[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
		accessedAt: time.Now(),
	}
	
	c.stats.Size = len(c.items)
}

// Delete removes an item from the cache
func (c *inMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	delete(c.items, key)
	c.stats.Size = len(c.items)
}

// Clear removes all items from the cache
func (c *inMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.items = make(map[string]cacheItem)
	c.stats.Size = 0
}

// Stats returns current cache statistics
func (c *inMemoryCache) Stats() CacheStats {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	return c.stats
}

// evictLRU removes the least recently used item
func (c *inMemoryCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time
	
	for key, item := range c.items {
		if oldestKey == "" || item.accessedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.accessedAt
		}
	}
	
	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}

// cleanup periodically removes expired items
func (c *inMemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.expiresAt) {
				delete(c.items, key)
			}
		}
		c.stats.Size = len(c.items)
		c.mu.Unlock()
	}
}