# GoCache

GoCache is a small Go cache project scaffold.

## Current Features

- Generic cache interface
- Goroutine-safe in-memory cache
- Optional per-key TTL
- Duplicate-call suppression with `singleflight.Group`

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/yourname/gocache"
)

func main() {
	cache := gocache.NewMemoryCache[string, string]()
	cache.Set("hello", "world", 0)

	value, ok := cache.Get("hello")
	fmt.Println(value, ok)
}
```

## Development

```sh
go test ./...
```
