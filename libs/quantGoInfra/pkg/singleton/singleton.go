package singleton

import (
	"sync"
)

// 泛型 + 工厂模式

type Singleton[T any] struct {
	once sync.Once
	val  T
	init func() T
}

func NewSingleton[T any](initFunc func() T) *Singleton[T] {
	return &Singleton[T]{init: initFunc}
}

func (s *Singleton[T]) Get() T {
	s.once.Do(func() {
		s.val = s.init()
	})
	return s.val
}
