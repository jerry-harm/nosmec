package memory

import (
	"runtime"
	"sync"
	"time"
)

type cacheValue[V any] struct {
	Value      V
	Expiration int64
}

type Cache[V any] struct {
	data map[[32]byte]cacheValue[V]
	mu   sync.RWMutex
}

func New[V any](maxSize int) *Cache[V] {
	c := &Cache[V]{
		data: make(map[[32]byte]cacheValue[V], maxSize),
	}
	go c.runJanitor()
	return c
}

func (c *Cache[V]) runJanitor() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		c.cleanup()
	}
}

func (c *Cache[V]) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now().UnixNano()
	for k, v := range c.data {
		if v.Expiration > 0 && v.Expiration < now {
			delete(c.data, k)
		}
	}
	runtime.GC()
}

func (c *Cache[V]) Get(k [32]byte) (v V, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.data[k]
	if !found {
		return *new(V), false
	}
	if item.Expiration > 0 && item.Expiration < time.Now().UnixNano() {
		return *new(V), false
	}
	return item.Value, true
}

func (c *Cache[V]) Delete(k [32]byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, k)
}

func (c *Cache[V]) Set(k [32]byte, v V) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[k] = cacheValue[V]{Value: v}
	return true
}

func (c *Cache[V]) SetWithTTL(k [32]byte, v V, d time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[k] = cacheValue[V]{
		Value:      v,
		Expiration: time.Now().Add(d).UnixNano(),
	}
	return true
}