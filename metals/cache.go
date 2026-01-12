package metals

import (
	"sync"
	"time"
)

// Cache stores metal prices with TTL
type Cache struct {
	mu      sync.RWMutex
	items   map[string]*cacheItem
	ttl     time.Duration
	enabled bool
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

var (
	globalCache     *Cache
	globalCacheLock sync.Mutex
)

// Default cache TTL for metal prices (1 hour - prices don't change that frequently)
const defaultCacheTTL = 1 * time.Hour

// getCache returns the global cache instance
func getCache() *Cache {
	globalCacheLock.Lock()
	defer globalCacheLock.Unlock()

	if globalCache == nil {
		globalCache = &Cache{
			items:   make(map[string]*cacheItem),
			ttl:     defaultCacheTTL,
			enabled: true,
		}
	}
	return globalCache
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	if !c.enabled {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(item.expiration) {
		return nil, false
	}

	return item.value, true
}

// Set stores a value in the cache
func (c *Cache) Set(key string, value interface{}) {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &cacheItem{
		value:      value,
		expiration: time.Now().Add(c.ttl),
	}
}

// Clear removes all items from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
}

// SetTTL sets the cache TTL
func (c *Cache) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ttl = ttl
}

// SetEnabled enables or disables the cache
func (c *Cache) SetEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.enabled = enabled
	if !enabled {
		c.items = make(map[string]*cacheItem)
	}
}

// SetCacheTTL sets the global cache TTL for metal prices
func SetCacheTTL(ttl time.Duration) {
	getCache().SetTTL(ttl)
}

// ClearCache clears the global cache
func ClearCache() {
	getCache().Clear()
}

// DisableCache disables caching globally
func DisableCache() {
	getCache().SetEnabled(false)
}

// EnableCache enables caching globally
func EnableCache() {
	getCache().SetEnabled(true)
}

