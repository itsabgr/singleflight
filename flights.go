package singleflight

import (
	"errors"
	"sync"
)

type Flights[K comparable, V any] struct {
	mutex sync.Mutex
	_map  map[K]V
}

func (s *Flights[K, V]) TryAcquire(key K, val V) (_ V, swapped bool) {
	return s.tryAcquireFunc(key, func() V {
		return val
	})
}

func (s *Flights[K, V]) tryAcquireFunc(key K, fn func() V) (_ V, swapped bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s._map == nil {
		s._map = make(map[K]V)
	} else {
		v, lockExists := s._map[key]
		if lockExists {
			return v, false
		}
	}
	val := fn()
	s._map[key] = val
	return val, true
}

var ErrUnlocked = errors.New("sync: unlock of unlocked mutex")

func (s *Flights[K, V]) Release(key K) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s._map == nil {
		panic(ErrUnlocked)
	}
	_, lockExists := s._map[key]
	if !lockExists {
		panic(ErrUnlocked)
	}
	delete(s._map, key)
}

func (s *Flights[K, V]) release(key K) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s._map, key)
}
