package singleflight

import (
	"context"
	"github.com/itsabgr/go-completer"
)

type Group[K comparable, R any] struct {
	locks Flights[K, completer.WaitContextFunc[R]]
}

func (g *Group[K, R]) Do(ctx context.Context, key K, fn func() R) (result R, err error) {
	var complete completer.CompleteFunc[R]
	wait, acquired := g.locks.tryAcquireFunc(key, func() (wait completer.WaitContextFunc[R]) {
		wait, complete = completer.WithContext[R]()
		return wait
	})
	if !acquired {
		return wait(ctx)
	}
	defer func() {
		g.locks.release(key)
		complete(result)
	}()
	result = fn()
	return result, nil
}
