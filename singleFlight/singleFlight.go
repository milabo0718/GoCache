package singleflight

import "sync"

type call[T any] struct {
	wg  sync.WaitGroup
	val T
	err error
}

// 要用sync.Map吗还是原生Map加锁？
type Group[T any] struct {
	mu sync.Mutex
	m  map[string]*call[T]
}

func (g *Group[T]) Do(key string, fn func() (T, error)) (T, error) {
	if existingCall, ok := g.m[key]; ok {
		existingCall.wg.Wait()
		return existingCall.val, existingCall.err
	}
	g.mu.Lock()
	defer g.mu.Unlock()

	val, err := fn()
	c := &call[T]{val: val, err: err}
	c.wg.Add(1)
	g.m[key] = c
	c.wg.Done()
	// 删除key
	delete(g.m, key)

	return val, err

}
