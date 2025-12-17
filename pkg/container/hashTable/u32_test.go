package hashTable

import (
	"sync/atomic"
	"testing"
)

//
// 方式 A：裸 uint32 + atomic.*Uint32
//

type RawSlot struct {
	Used uint32
}

func BenchmarkRawLoad(b *testing.B) {
	var s RawSlot
	s.Used = 1
	for i := 0; i < b.N; i++ {
		_ = atomic.LoadUint32(&s.Used)
	}
}

func BenchmarkRawStore(b *testing.B) {
	var s RawSlot
	for i := 0; i < b.N; i++ {
		atomic.StoreUint32(&s.Used, uint32(i))
	}
}

func BenchmarkRawCAS(b *testing.B) {
	var s RawSlot
	s.Used = 0
	for i := 0; i < b.N; i++ {
		atomic.CompareAndSwapUint32(&s.Used, 0, 1)
		atomic.StoreUint32(&s.Used, 0)
	}
}

//
// 方式 B：atomic.Uint32 的方法调用
//

type AtomicSlot struct {
	Used atomic.Uint32
}

func BenchmarkAtomicLoad(b *testing.B) {
	var s AtomicSlot
	s.Used.Store(1)
	for i := 0; i < b.N; i++ {
		_ = s.Used.Load()
	}
}

func BenchmarkAtomicStore(b *testing.B) {
	var s AtomicSlot
	for i := 0; i < b.N; i++ {
		s.Used.Store(uint32(i))
	}
}

func BenchmarkAtomicCAS(b *testing.B) {
	var s AtomicSlot
	s.Used.Store(0)
	for i := 0; i < b.N; i++ {
		s.Used.CompareAndSwap(0, 1)
		s.Used.Store(0)
	}
}
