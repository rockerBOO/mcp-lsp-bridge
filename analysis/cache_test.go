package analysis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewAnalysisCache tests the creation of a new analysis cache
func TestNewAnalysisCache(t *testing.T) {
	cache := NewAnalysisCache(100, 1*time.Hour)
	
	assert.NotNil(t, cache)
	
	stats := cache.Stats()
	assert.Zero(t, stats.Size)
	assert.Zero(t, stats.Hits)
	assert.Zero(t, stats.Misses)
}

// TestCacheSetAndGet tests setting and retrieving items from the cache
func TestCacheSetAndGet(t *testing.T) {
	cache := NewAnalysisCache(100, 1*time.Hour)
	
	// Set an item
	testValue := "test_value"
	cache.Set("test_key", testValue, 0)
	
	// Retrieve the item
	value, found := cache.Get("test_key")
	
	assert.True(t, found)
	assert.Equal(t, testValue, value)
	
	// Check stats
	stats := cache.Stats()
	assert.Equal(t, 1, stats.Size)
	assert.Equal(t, int64(1), stats.Hits)
}

// TestCacheExpiration tests cache item expiration
func TestCacheExpiration(t *testing.T) {
	// Very short TTL for testing
	cache := NewAnalysisCache(100, 10*time.Millisecond)
	
	// Set an item
	testValue := "expiring_value"
	cache.Set("expiring_key", testValue, 0)
	
	// Initial cache size
	initialStats := cache.Stats()
	assert.Equal(t, 1, initialStats.Size)
	
	// Wait for expiration
	time.Sleep(20 * time.Millisecond)
	
	// Try to retrieve expired item
	value, found := cache.Get("expiring_key")
	
	assert.False(t, found)
	assert.Nil(t, value)
	
	// Check stats
	stats := cache.Stats()
	assert.Zero(t, stats.Size)
	assert.Equal(t, int64(1), stats.Misses)
}

// TestCacheLRUEviction tests Least Recently Used (LRU) eviction
func TestCacheLRUEviction(t *testing.T) {
	// Small cache size to force eviction
	cache := NewAnalysisCache(2, 1*time.Hour)
	
	// Add three items
	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)
	
	// Access first item to change its last access time
	cache.Get("key1")
	
	// Add third item (should evict key2)
	cache.Set("key3", "value3", 0)
	
	// Check item availability
	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	_, found3 := cache.Get("key3")
	
	assert.True(t, found1)
	assert.False(t, found2)
	assert.True(t, found3)
	
	// Check stats
	stats := cache.Stats()
	assert.Equal(t, 2, stats.Size)
}

// TestCacheClear tests clearing the entire cache
func TestCacheClear(t *testing.T) {
	cache := NewAnalysisCache(100, 1*time.Hour)
	
	// Add some items
	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)
	
	// Clear the cache
	cache.Clear()
	
	// Check stats
	stats := cache.Stats()
	assert.Zero(t, stats.Size)
	
	// Verify items are gone
	_, found1 := cache.Get("key1")
	_, found2 := cache.Get("key2")
	
	assert.False(t, found1)
	assert.False(t, found2)
}

// TestCacheDelete tests deleting a specific item
func TestCacheDelete(t *testing.T) {
	cache := NewAnalysisCache(100, 1*time.Hour)
	
	// Add an item
	cache.Set("key_to_delete", "value", 0)
	
	// Delete the item
	cache.Delete("key_to_delete")
	
	// Check stats
	stats := cache.Stats()
	assert.Zero(t, stats.Size)
	
	// Verify item is gone
	_, found := cache.Get("key_to_delete")
	assert.False(t, found)
}