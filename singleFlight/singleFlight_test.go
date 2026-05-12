package singleflight

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGroupSuppressesDuplicateCalls(t *testing.T) {
	var group Group[string, int]
	var calls int32
	const goroutines = 10

	start := make(chan struct{})
	results := make(chan int, goroutines)
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			value, err, _ := group.Do("key", func() (int, error) {
				atomic.AddInt32(&calls, 1)
				time.Sleep(20 * time.Millisecond)
				return 7, nil
			})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			results <- value
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	if calls != 1 {
		t.Fatalf("got %d calls, want 1", calls)
	}
	for got := range results {
		if got != 7 {
			t.Fatalf("got %d, want 7", got)
		}
	}
}
