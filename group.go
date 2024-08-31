package singleflight

import (
	"context"
	"github.com/itsabgr/go-completer"
	"sync"
)

type Group[K comparable, R any] struct {
	mutex sync.Mutex
	_map  map[K]completer.WaitContextFunc[R]
}

func (g *Group[K, R]) Do(key K, fn func() R) completer.WaitContextFunc[R] {
	var complete completer.CompleteFunc[R]
	wait, acquired := g.tryAcquireFunc(key, func() (wait completer.WaitContextFunc[R]) {
		wait, complete = completer.WithContext[R]()
		return wait
	})
	if !acquired {
		return wait
	}
	var result R
	defer func() {
		g.release(key)
		complete(result)
	}()
	result = fn()
	return func(ctx context.Context) (R, error) {
		return result, ctx.Err()
	}
}

func (g *Group[K, R]) tryAcquireFunc(key K, fn func() completer.WaitContextFunc[R]) (_ completer.WaitContextFunc[R], swapped bool) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	if g._map == nil {
		g._map = make(map[K]completer.WaitContextFunc[R])
	} else {
		v, lockExists := g._map[key]
		if lockExists {
			return v, false
		}
	}
	val := fn()
	g._map[key] = val
	return val, true
}

func (g *Group[K, V]) release(key K) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	delete(g._map, key)
}
func (g *Group[K, V]) Release() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	clear(g._map)
}
