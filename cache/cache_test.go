package cache

import (
	"testing"
	"time"
)

func TestCache_GetSet(t *testing.T) {
	c := New(DefaultConfig())
	
	// Test set and get
	c.Set("key1", "value1")
	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("Expected to find key1")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}
	
	// Test missing key
	_, ok = c.Get("missing")
	if ok {
		t.Error("Expected missing key to return false")
	}
}

func TestCache_Expiration(t *testing.T) {
	config := DefaultConfig()
	config.TTL = 100 * time.Millisecond
	c := New(config)
	
	c.Set("key1", "value1")
	
	// Should exist immediately
	_, ok := c.Get("key1")
	if !ok {
		t.Error("Expected key to exist")
	}
	
	// Wait for expiration
	time.Sleep(150 * time.Millisecond)
	
	// Should be expired
	_, ok = c.Get("key1")
	if ok {
		t.Error("Expected key to be expired")
	}
	
	// Check stats
	stats := c.GetStats()
	if stats.Expirations == 0 {
		t.Error("Expected expiration to be recorded")
	}
}

func TestCache_MaxSize(t *testing.T) {
	config := DefaultConfig()
	config.MaxSize = 3
	c := New(config)
	
	// Add 4 items (should evict oldest)
	c.Set("key1", "value1")
	time.Sleep(10 * time.Millisecond) // Ensure different access times
	c.Set("key2", "value2")
	time.Sleep(10 * time.Millisecond)
	c.Set("key3", "value3")
	time.Sleep(10 * time.Millisecond)
	c.Set("key4", "value4")
	
	// key1 should be evicted (oldest)
	_, ok := c.Get("key1")
	if ok {
		t.Error("Expected key1 to be evicted")
	}
	
	// Others should exist
	if _, ok := c.Get("key2"); !ok {
		t.Error("Expected key2 to exist")
	}
	if _, ok := c.Get("key3"); !ok {
		t.Error("Expected key3 to exist")
	}
	if _, ok := c.Get("key4"); !ok {
		t.Error("Expected key4 to exist")
	}
	
	// Check stats
	stats := c.GetStats()
	if stats.Evictions == 0 {
		t.Error("Expected eviction to be recorded")
	}
}

func TestCache_Delete(t *testing.T) {
	c := New(DefaultConfig())
	
	c.Set("key1", "value1")
	c.Delete("key1")
	
	_, ok := c.Get("key1")
	if ok {
		t.Error("Expected key to be deleted")
	}
}

func TestCache_Clear(t *testing.T) {
	c := New(DefaultConfig())
	
	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Clear()
	
	if c.Size() != 0 {
		t.Errorf("Expected size 0, got %d", c.Size())
	}
}

func TestCache_Disabled(t *testing.T) {
	config := DefaultConfig()
	config.Enabled = false
	c := New(config)
	
	c.Set("key1", "value1")
	_, ok := c.Get("key1")
	if ok {
		t.Error("Expected cache to be disabled")
	}
}

func TestCache_Stats(t *testing.T) {
	c := New(DefaultConfig())
	
	c.Set("key1", "value1")
	c.Get("key1") // Hit
	c.Get("key1") // Hit
	c.Get("missing") // Miss
	
	stats := c.GetStats()
	if stats.Hits != 2 {
		t.Errorf("Expected 2 hits, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
	
	hitRate := stats.HitRate()
	expected := 2.0 / 3.0
	if hitRate < expected-0.01 || hitRate > expected+0.01 {
		t.Errorf("Expected hit rate ~%.2f, got %.2f", expected, hitRate)
	}
}

func TestCache_Cleanup(t *testing.T) {
	config := DefaultConfig()
	config.TTL = 50 * time.Millisecond
	config.CleanupInterval = 100 * time.Millisecond
	c := New(config)
	
	c.Set("key1", "value1")
	c.Set("key2", "value2")
	
	// Wait for expiration and cleanup
	time.Sleep(200 * time.Millisecond)
	
	// Size should be 0 after cleanup
	if c.Size() != 0 {
		t.Errorf("Expected size 0 after cleanup, got %d", c.Size())
	}
}

