package gocache

import (
	"testing"
	"time"
)

func TestMemoryCacheSetGet(t *testing.T) {
	cache := NewMemoryCache[string, int]()

	cache.Set("answer", 42, 0)

	got, ok := cache.Get("answer")
	if !ok {
		t.Fatal("expected key to exist")
	}
	if got != 42 {
		t.Fatalf("got %d, want 42", got)
	}
}

func TestMemoryCacheExpiration(t *testing.T) {
	cache := NewMemoryCache[string, string]()
	now := time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC)
	cache.now = func() time.Time { return now }

	cache.Set("token", "abc", time.Minute)
	now = now.Add(time.Minute)

	if _, ok := cache.Get("token"); ok {
		t.Fatal("expected key to expire")
	}
	if got := cache.Len(); got != 0 {
		t.Fatalf("got len %d, want 0", got)
	}
}

func TestMemoryCacheDeleteAndClear(t *testing.T) {
	cache := NewMemoryCache[string, string]()
	cache.Set("a", "1", 0)
	cache.Set("b", "2", 0)

	cache.Delete("a")
	if _, ok := cache.Get("a"); ok {
		t.Fatal("expected deleted key to be missing")
	}

	cache.Clear()
	if got := cache.Len(); got != 0 {
		t.Fatalf("got len %d, want 0", got)
	}
}
