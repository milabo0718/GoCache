package gocache

import (
	"sync"

	"github.com/milabo0718/gocache/memory"
)

type Cache struct {
	mu    sync.RWMutex
	store memory.Store
}

func NewCache() *Cache {
	return &Cache{
		store: memory.NewCache("LRU", 100), // default to 100 items
	}
}

func (c *Cache) Get(key string) (ByteView, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if v, ok := c.store.Get(key); ok {
		return ByteView{b: v}, true
	}
	return ByteView{}, false
}

func (c *Cache) Set(key string, value []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.store.Set(key, value)
}

func (c *Cache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	flag, err := c.store.Delete(key)
	if !flag {
		return err
	}
	return nil
}
