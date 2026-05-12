package gocache

import "time"

// Cache defines the core operations shared by cache implementations.
type Cache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V, ttl time.Duration)
	Delete(key K)
	Clear()
	Len() int
}
