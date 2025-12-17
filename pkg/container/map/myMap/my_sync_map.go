package myMap

import (
	"sync"
	"sync/atomic"
)

// MySyncMap 是线程安全、类型安全的泛型 map
type MySyncMap[K comparable, V any] struct {
	smap sync.Map
	size atomic.Int64
}

func NewMySyncMap[K comparable, V any]() *MySyncMap[K, V] {
	return &MySyncMap[K, V]{}
}

func (m *MySyncMap[K, V]) Load(k K) (V, bool) {
	v, ok := m.smap.Load(k)
	if ok {
		return v.(V), true
	}
	var zero V
	return zero, false
}

// Store 插入或更新指定键的值,并自动维护 size
func (m *MySyncMap[K, V]) Store(k K, v V) {
	_, loaded := m.smap.LoadOrStore(k, v)
	if loaded {
		m.smap.Store(k, v) // 替换旧值,不变更 size
	} else {
		m.size.Add(1) // 新增 key，增加 size
	}
}

// Delete 移除指定键,并自动更新 size
func (m *MySyncMap[K, V]) Delete(k K) {
	_, existed := m.smap.LoadAndDelete(k)
	if existed {
		m.size.Add(-1)
	}
}

func (m *MySyncMap[K, V]) Range(f func(k K, v V) bool) {
	m.smap.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}

func (m *MySyncMap[K, V]) Length() int64 {
	return m.size.Load()
}

func (m *MySyncMap[K, V]) Clear() {
	m.smap.Range(func(k, _ any) bool {
		m.smap.Delete(k)
		return true
	})
	m.size.Store(0)
}
