package singleflight

import (
	"github.com/itsabgr/go-completer"
)

type Single[R any] struct {
	flight Group[struct{}, R]
}

func (s *Single[R]) Do(fn func() R) completer.WaitContextFunc[R] {
	return s.flight.Do(struct{}{}, fn)
}

func (s *Single[R]) Release() {
	s.flight.Release()
}
