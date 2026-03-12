package transcript

import (
	"context"
	"sync"
	"time"
)

type cacheItem struct {
	Value      string
	Expiration int64
}

type InMemoryCache struct {
	items sync.Map
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{}
}

func (c *InMemoryCache) Get(ctx context.Context, key string) (string, error) {
	val, ok := c.items.Load(key)
	if !ok {
		return "", nil
	}

	item := val.(cacheItem)
	if item.Expiration != 0 && time.Now().UnixMilli() > item.Expiration {
		c.items.Delete(key)
		return "", nil
	}

	return item.Value, nil
}

func (c *InMemoryCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	var expiration int64
	if ttl != 0 {
		expiration = time.Now().Add(ttl).UnixMilli()
	}

	c.items.Store(key, cacheItem{
		Value:      value,
		Expiration: expiration,
	})

	return nil
}
