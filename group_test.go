package singleflight

import (
	"context"
	"errors"
	"github.com/itsabgr/fak"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCorrectness1(t *testing.T) {
	numGoroutines := 64
	g := Group[int, int]{}
	done := atomic.Bool{}

	wg := sync.WaitGroup{}
	wg.Add(numGoroutines)
	start := sync.WaitGroup{}
	start.Add(numGoroutines)

	fn := func() int {
		start.Wait()
		if done.CompareAndSwap(false, true) {
			return numGoroutines
		}
		t.FailNow()
		return 0
	}
	for range numGoroutines {
		go func() {
			defer wg.Done()
			start.Done()
			result := fak.Must(g.Do(context.Background(), 1, fn))
			fak.Assert(result == numGoroutines, nil)
		}()
	}
	wg.Wait()
}

func TestCorrectness2(t *testing.T) {
	g := Group[int, int]{}
	start := sync.WaitGroup{}
	start.Add(1)
	dead := sync.WaitGroup{}
	dead.Add(1)

	go g.Do(context.Background(), 1, func() int {
		start.Done()
		dead.Wait()
		return 2
	})
	start.Wait()
	res, err := fak.Timeout(context.Background(), time.Millisecond*100, func(timeoutCtx context.Context) (int, error) {
		return g.Do(timeoutCtx, 1, func() int {
			t.FailNow()
			return 1
		})
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("unexpected %s", err)
	}
	if res != 0 {
		t.Fatalf("unexpected %d", res)
	}

}

func TestCorrectness3(t *testing.T) {
	numGoroutines := 64
	g := Group[int, int]{}
	start := sync.WaitGroup{}
	start.Add(numGoroutines)

	dead := sync.WaitGroup{}
	dead.Add(1)
	defer dead.Done()
	count := atomic.Int64{}
	for i := range numGoroutines {
		go g.Do(context.Background(), i, func() int {
			count.Add(1)
			start.Done()
			dead.Wait()
			return i
		})
	}
	start.Wait()
	if count.Load() != int64(numGoroutines) {
		t.FailNow()
	}
}
