package dataloader

import (
	"context"
	"errors"
	"sync"
	"time"
)

var NoValueError = errors.New("<dataloader: no value>")

type BatchFunc[K comparable, V any] func([]context.Context, []K) map[K]Result[V]

type Result[V any] struct {
	Data  V
	Error error
}

type Loader[K comparable, V any] struct {
	batchFn  BatchFunc[K, V]
	batchCap int
	wait     time.Duration
	batchLock sync.Mutex
	curBatcher *batcher[K, V]
}

type batchRequest[K comparable, V any] struct {
	ctx     context.Context
	key     K
	channel chan Result[V]
}

type Options struct {
	Wait         time.Duration
	MaxThreshold uint
}

func NewBatchedLoader[K comparable, V any](batchFn BatchFunc[K, V], opts Options) *Loader[K, V] {
	loader := &Loader[K, V]{
		batchFn:  batchFn,
		batchCap: int(opts.MaxThreshold),
		wait:     opts.Wait,
	}
	loader.curBatcher = loader.newBatcher()
	return loader
}

func (l *Loader[K, V]) Load(ctx context.Context, key K) (value V, err error) {
	c := make(chan Result[V], 1)
	req := batchRequest[K, V]{ctx, key, c}

	l.batchLock.Lock()

	if len(l.curBatcher.requests) == 0 {
		go func(b *batcher[K, V]) {
			select {
			case <-b.thresholdReached:
			case <-time.After(l.wait):
				l.batchLock.Lock()
				if l.curBatcher == b {
					l.curBatcher = l.newBatcher()
				}
				l.batchLock.Unlock()
			}

			ctxs := make([]context.Context, 0, len(b.requests))
			keys := make([]K, 0, len(b.requests))

			for _, item := range b.requests {
				ctxs = append(ctxs, item.ctx)
				keys = append(keys, item.key)
			}

			res := l.batchFn(ctxs, keys)

			for _, req := range b.requests {
				if r, ok := res[req.key]; ok {
					req.channel <- r
				}
				close(req.channel)
			}
		}(l.curBatcher)
	}

	l.curBatcher.requests = append(l.curBatcher.requests, req)
	if len(l.curBatcher.requests) == l.batchCap {
		close(l.curBatcher.thresholdReached)
		l.curBatcher = l.newBatcher()
	}

	l.batchLock.Unlock()

	select {
	case v, ok := <-c:
		if ok {
			return v.Data, v.Error
		}
		return value, NoValueError
	case <-ctx.Done():
		return value, ctx.Err()
	}
}

type batcher[K comparable, V any] struct {
	thresholdReached chan struct{}
	requests         []batchRequest[K, V]
	batchFn          BatchFunc[K, V]
}

func (l *Loader[K, V]) newBatcher() *batcher[K, V] {
	return &batcher[K, V]{
		thresholdReached: make(chan struct{}),
		requests:         make([]batchRequest[K, V], 0, l.batchCap),
		batchFn:          l.batchFn,
	}
}