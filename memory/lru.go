package memory

import (
	"sync"
)

type node struct {
	key   string
	value []byte
	prev  *node
	next  *node
}

type lruCache struct {
	mu       sync.Mutex
	head     *node
	tail     *node
	cache    map[string]*node
	size     int
	capacity int
}

func newLRUCache(capacity int) *lruCache {
	c := &lruCache{
		cache:    make(map[string]*node),
		capacity: capacity,
		head:     &node{},
		tail:     &node{},
		size:     0,
	}
	c.head.next = c.tail
	c.tail.prev = c.head
	return c
}

func (c *lruCache) addNode(n *node) {
	prev := c.head
	next := c.head.next
	prev.next = n
	n.prev = prev
	n.next = next
	next.prev = n
}

func (c *lruCache) removeNode(n *node) {
	prev := n.prev
	next := n.next
	prev.next = next
	next.prev = prev
}

func (c *lruCache) moveToFront(n *node) {
	c.removeNode(n)
	c.addNode(n)
}

func (c *lruCache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if n, ok := c.cache[key]; ok {
		c.moveToFront(n)
		return n.value, true
	}
	return nil, false
}

func (c *lruCache) Set(key string, value []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if n, ok := c.cache[key]; ok {
		n.value = value
		c.moveToFront(n)
		return nil
	}

	n := &node{
		key:   key,
		value: value,
	}
	c.cache[key] = n
	c.addNode(n)
	c.size++
	if c.size > c.capacity {
		lru := c.tail.prev
		c.removeNode(lru)
		delete(c.cache, lru.key)
		c.size--
	}
	return nil
}

func (c *lruCache) Delete(key string) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if n, ok := c.cache[key]; ok {
		c.removeNode(n)
		delete(c.cache, key)
		c.size--
		return true, nil
	}
	return false, nil
}
