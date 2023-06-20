package cache

import (
	"context"
	"time"

	"github.com/allegro/bigcache/v3"
)

func NewBigCache() *bigcache.BigCache {
	cache, err := bigcache.New(context.Background(), bigcache.DefaultConfig(24*time.Hour))
	if err != nil {
		panic("cannot instantiate bigcache")
	}
	return cache
}
