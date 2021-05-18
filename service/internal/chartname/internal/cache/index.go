package cache

import (
	"context"

	gocache "github.com/patrickmn/go-cache"
)

type Index struct {
	cache *gocache.Cache
}

func NewIndex() *Index {
	i := &Index{
		cache: gocache.New(expiration, expiration/2),
	}

	return i
}

func (i *Index) Get(ctx context.Context, key string) ([]byte, bool) {
	val, ok := i.cache.Get(key)
	if ok {
		return val.([]byte), true
	}

	return nil, false
}

func (i *Index) Set(ctx context.Context, key string, val []byte) {
	i.cache.SetDefault(key, val)
}
