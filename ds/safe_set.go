package ds

import "sync"

type empty struct{}

type SafeSet[T any] struct {
	m sync.Map
}

func NewSafeSet[T any]() *SafeSet[T] {
	return &SafeSet[T]{}
}

func (s *SafeSet[T]) Add(item T) {
	s.m.Store(item, empty{})
}

func (s *SafeSet[T]) Remove(item T) {
	s.m.Delete(item)
}

func (s *SafeSet[T]) Contains(item T) bool {
	_, ok := s.m.Load(item)

	return ok
}

func (s *SafeSet[T]) Range(f func(item T) bool) {
	s.m.Range(func(key, value any) bool {
		return f(key.(T))
	})
}
