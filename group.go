package gocache

import (
	"context"
	"errors"
	"sync"
)

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

type Getter interface {
	Get(ctx context.Context, key string) ([]byte, error)
}

type Group struct {
	name      string
	getter    Getter
	mainCache *Cache
}

func NewGroup(name string, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: NewCache(),
	}
	mu.Lock()
	defer mu.Unlock()
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, nil
	}

	if v, ok := g.mainCache.Get(key); ok {
		return v, nil
	}

	return ByteView{}, nil
}

func (g *Group) Set(key string, value []byte) error {
	if key == "" {
		return errors.New("key is required")
	}
	if len(value) == 0 {
		return errors.New("value is required")
	}

	return g.mainCache.Set(key, value)
}

func (g *Group) Delete(key string) error {
	if key == "" {
		return errors.New("key is required")
	}

	return g.mainCache.Delete(key)
}
