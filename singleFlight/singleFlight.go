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

// singleflight的机制就是要保证对于某个key，同一时间只有一个请求在执行，其他请求等待这个请求完成后直接返回结果。为了实现这个机制，我们需要一个结构体来存储正在执行的请求的信息，包括结果和错误。我们还需要一个互斥锁来保护这个结构体，以确保线程安全。
func (g *Group[T]) Do(key string, fn func() (T, error)) (T, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call[T])
	}

	if existingCall, ok := g.m[key]; ok {
		g.mu.Unlock()
		existingCall.wg.Wait()
		return existingCall.val, existingCall.err
	}

	c := &call[T]{}
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()
	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	// 删除key
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err

}
