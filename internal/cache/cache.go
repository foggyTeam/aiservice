package cache

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// CacheItem represents a cached item with expiration
type CacheItem struct {
	Value      any
	Expiration time.Time
}

// IsExpired checks if the cache item has expired
func (item *CacheItem) IsExpired() bool {
	return time.Now().After(item.Expiration)
}

// Cache interface defines the contract for cache implementations
type Cache interface {
	Get(key string) (any, bool)
	Set(key string, value any, ttl time.Duration)
	Delete(key string)
	Clear()
}

// InMemoryCache is a simple in-memory cache implementation
type InMemoryCache struct {
	items map[string]*CacheItem
	mutex sync.RWMutex
}

// NewInMemoryCache creates a new in-memory cache instance
func NewInMemoryCache() *InMemoryCache {
	cache := &InMemoryCache{
		items: make(map[string]*CacheItem),
	}

	// Start a goroutine to clean up expired items periodically
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a value from the cache
func (c *InMemoryCache) Get(key string) (any, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	if item.IsExpired() {
		// Don't remove here to avoid race conditions, cleanup will handle it
		return nil, false
	}

	return item.Value, true
}

// Set stores a value in the cache with a TTL
func (c *InMemoryCache) Set(key string, value any, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	expiration := time.Now().Add(ttl)
	c.items[key] = &CacheItem{
		Value:      value,
		Expiration: expiration,
	}
}

// Delete removes a key from the cache
func (c *InMemoryCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
}

// Clear removes all items from the cache
func (c *InMemoryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*CacheItem)
}

// cleanupExpired periodically removes expired items from the cache
func (c *InMemoryCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute) // Clean up every 5 minutes
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.Expiration) {
				delete(c.items, key)
			}
		}
		c.mutex.Unlock()
	}
}

// GenerateKey creates a cache key from input data
func GenerateKey(data ...string) string {
	hash := sha256.New()
	for _, d := range data {
		hash.Write([]byte(d))
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}
