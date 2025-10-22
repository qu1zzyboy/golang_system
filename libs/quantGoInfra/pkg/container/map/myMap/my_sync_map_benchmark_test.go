package myMap

import (
	"fmt"
	"testing"
)

// goos: windows
// goarch: amd64
// pkg: QuantityRR/pkg/myMap
// cpu: Intel(R) Core(TM) i7-14700KF
// Benchmark_MySyncMap_WithSize_Store-28           	 7075609	       194.8 ns/op	     106 B/op	       6 allocs/op
// Benchmark_MySyncMap_NoSize_Store-28             	 8343194	       189.3 ns/op	     105 B/op	       6 allocs/op
// Benchmark_MySyncMap_WithSize_Length-28          	1000000000	         0.1834 ns/op	       0 B/op	       0 allocs/op
// Benchmark_MySyncMap_NoSize_Length-28            	    1515	    782198 ns/op	       0 B/op	       0 allocs/op
// Benchmark_MySyncMap_WithSize_ReadExisting-28    	1000000000	         0.7851 ns/op	       0 B/op	       0 allocs/op
// Benchmark_MySyncMap_WithSize_ReadMissing-28     	1000000000	         0.3767 ns/op	       0 B/op	       0 allocs/op
// Benchmark_MySyncMap_NoSize_ReadExisting-28      	1000000000	         0.8432 ns/op	       0 B/op	       0 allocs/op
// Benchmark_MySyncMap_NoSize_ReadMissing-28       	1000000000	         0.3787 ns/op	       0 B/op	       0 allocs/op

func Benchmark_MySyncMap_WithSize_Store(b *testing.B) {
	m := NewMySyncMap[string, int]()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			m.Store(fmt.Sprintf("key%d", i), i)
		}
	})
}

func Benchmark_MySyncMap_WithSize_Length(b *testing.B) {
	m := NewMySyncMap[string, int]()
	for i := 0; i < 100000; i++ {
		m.Store(fmt.Sprintf("key%d", i), i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Length()
	}
}

func Benchmark_MySyncMap_WithSize_ReadExisting(b *testing.B) {
	m := NewMySyncMap[string, int]()
	m.Store("exist", 42)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Load("exist")
		}
	})
}

func Benchmark_MySyncMap_WithSize_ReadMissing(b *testing.B) {
	m := NewMySyncMap[string, int]()
	// 没有 Store key
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Load("not_found")
		}
	})
}
