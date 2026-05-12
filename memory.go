package gocache

import (
	"sync"
	"time"
)

type entry[V any] struct {
	value     V
	expiresAt time.Time
}

// MemoryCache is a goroutine-safe in-memory cache with optional per-key TTL.
type MemoryCache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]entry[V]
	now   func() time.Time
}

// NewMemoryCache creates an empty in-memory cache.
func NewMemoryCache[K comparable, V any]() *MemoryCache[K, V] {
	return &MemoryCache[K, V]{
		items: make(map[K]entry[V]),
		now:   time.Now,
	}
}

// Get returns a cached value when it exists and has not expired.
func (c *MemoryCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()

	var zero V
	if !ok {
		return zero, false
	}
	if !item.expiresAt.IsZero() && !c.now().Before(item.expiresAt) {
		c.Delete(key)
		return zero, false
	}
	return item.value, true
}

// Set stores a value. A ttl <= 0 means the value does not expire.
func (c *MemoryCache[K, V]) Set(key K, value V, ttl time.Duration) {
	item := entry[V]{value: value}
	if ttl > 0 {
		item.expiresAt = c.now().Add(ttl)
	}

	c.mu.Lock()
	c.items[key] = item
	c.mu.Unlock()
}

// Delete removes a key from the cache.
func (c *MemoryCache[K, V]) Delete(key K) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// Clear removes all keys from the cache.
func (c *MemoryCache[K, V]) Clear() {
	c.mu.Lock()
	c.items = make(map[K]entry[V])
	c.mu.Unlock()
}

// Len returns the number of currently live entries.
func (c *MemoryCache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.items {
		if !item.expiresAt.IsZero() && !c.now().Before(item.expiresAt) {
			delete(c.items, key)
		}
	}
	return len(c.items)
}
