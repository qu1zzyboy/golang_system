package myMap

import "sync"

type MyRWMap[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

func NewMyRWMap[K comparable, V any]() *MyRWMap[K, V] {
	return &MyRWMap[K, V]{
		data: make(map[K]V),
	}
}

func (m *MyRWMap[K, V]) Load(k K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.data[k]
	return v, ok
}

func (m *MyRWMap[K, V]) Store(k K, v V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[k] = v
}

func (m *MyRWMap[K, V]) Delete(k K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, k)
}

func (m *MyRWMap[K, V]) Length() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}
