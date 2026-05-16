package memory

type Store interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte) error
	Delete(key string) (bool, error)
}

type CacheType string

const (
	LRU  CacheType = "LRU"
	LRU2 CacheType = "LRU2"
)

func NewCache(cacheType CacheType, capacity int) Store {
	switch cacheType {
	case LRU:
		return newLRUCache(capacity)
	default:
		return nil
	}
}
