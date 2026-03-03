// Package cache provides caching functionality for findata-go
package cache

import (
	"sync"
	"time"
)

// Entry represents a cached entry with expiration
type Entry struct {
	Value      any
	ExpiresAt  time.Time
	AccessedAt time.Time
}

// IsExpired checks if the entry has expired
func (e *Entry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Cache is a thread-safe in-memory cache with TTL support
type Cache struct {
	mu      sync.RWMutex
	entries map[string]*Entry
	config  Config
	stats   Stats
}

// Config holds cache configuration
type Config struct {
	// TTL is the time-to-live for cache entries
	TTL time.Duration
	// MaxSize is the maximum number of entries (0 = unlimited)
	MaxSize int
	// Enabled controls whether caching is active
	Enabled bool
	// CleanupInterval is how often to clean expired entries
	CleanupInterval time.Duration
}

// Stats holds cache statistics
type Stats struct {
	mu          sync.RWMutex
	Hits        int64
	Misses      int64
	Evictions   int64
	Expirations int64
}

// DefaultConfig returns the default cache configuration
func DefaultConfig() Config {
	return Config{
		TTL:             5 * time.Minute,
		MaxSize:         1000,
		Enabled:         true,
		CleanupInterval: 10 * time.Minute,
	}
}

// New creates a new cache with the given configuration
func New(config Config) *Cache {
	c := &Cache{
		entries: make(map[string]*Entry),
		config:  config,
	}

	// Start cleanup goroutine if enabled
	if config.Enabled && config.CleanupInterval > 0 {
		go c.cleanupLoop()
	}

	return c
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (any, bool) {
	if !c.config.Enabled {
		c.stats.recordMiss()
		return nil, false
	}

	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()

	if !ok {
		c.stats.recordMiss()
		return nil, false
	}

	if entry.IsExpired() {
		c.stats.recordMiss()
		c.stats.recordExpiration()
		c.Delete(key)
		return nil, false
	}

	// Update access time
	c.mu.Lock()
	entry.AccessedAt = time.Now()
	c.mu.Unlock()

	c.stats.recordHit()
	return entry.Value, true
}

// Set stores a value in the cache
func (c *Cache) Set(key string, value any) {
	if !c.config.Enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict entries
	if c.config.MaxSize > 0 && len(c.entries) >= c.config.MaxSize {
		c.evictOldest()
	}

	c.entries[key] = &Entry{
		Value:      value,
		ExpiresAt:  time.Now().Add(c.config.TTL),
		AccessedAt: time.Now(),
	}
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*Entry)
}

// Size returns the number of entries in the cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// GetStats returns cache statistics
func (c *Cache) GetStats() Stats {
	return c.stats.copy()
}

// evictOldest removes the oldest accessed entry (LRU eviction)
func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.AccessedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.AccessedAt
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
		c.stats.recordEviction()
	}
}

// cleanupLoop periodically removes expired entries
func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup removes all expired entries
func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.entries {
		if entry.IsExpired() {
			delete(c.entries, key)
			c.stats.recordExpiration()
		}
	}
}

// Stats methods
func (s *Stats) recordHit() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Hits++
}

func (s *Stats) recordMiss() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Misses++
}

func (s *Stats) recordEviction() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Evictions++
}

func (s *Stats) recordExpiration() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Expirations++
}

func (s *Stats) copy() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return Stats{
		Hits:        s.Hits,
		Misses:      s.Misses,
		Evictions:   s.Evictions,
		Expirations: s.Expirations,
	}
}

// HitRate returns the cache hit rate (0.0 to 1.0)
func (s *Stats) HitRate() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := s.Hits + s.Misses
	if total == 0 {
		return 0.0
	}
	return float64(s.Hits) / float64(total)
}
