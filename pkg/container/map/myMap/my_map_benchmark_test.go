package myMap

import (
	"strconv"
	"testing"
)

// goos: windows
// goarch: amd64
// pkg: QuantityRR/pkg/myMap
// cpu: Intel(R) Core(TM) i7-14700KF
// BenchmarkMySyncMap_ReadOnly-28       	1000000000	         0.7806 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMySyncMap_WriteOnly-28      	 9017050	       171.9 ns/op	     100 B/op	       6 allocs/op
// BenchmarkMySyncMap_ReadHotKey-28     	12232440	        89.23 ns/op	      63 B/op	       4 allocs/op
// BenchmarkMySyncMap_MixedAccess-28    	38033418	        29.91 ns/op	      72 B/op	       5 allocs/op

// BenchmarkMyRWMap_ReadOnly-28         	34055305	        35.05 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMyRWMap_WriteOnly-28        	 9168164	       127.9 ns/op	      19 B/op	       2 allocs/op
// BenchmarkMyRWMap_ReadHotKey-28       	21558382	        56.09 ns/op	       0 B/op	       0 allocs/op
// BenchmarkMyRWMap_MixedAccess-28      	 8490231	       140.6 ns/op	       5 B/op	       1 allocs/op
// -28 表示使用了 28 线程(逻辑 CPU)并行跑测试

// 只读配置、指标只读			✅ MySyncMap
// 热点 key 频繁更新			✅ MyRWMap
// 批量初始化、一次性写入		✅ MyRWMap
// 高频写 + 多 key 并发读写缓存	✅ MySyncMap
// 对内存 & GC 压力特别敏感		✅ MyRWMap(alloc更少)

func BenchmarkMySyncMap_ReadOnly(b *testing.B) {
	m := NewMySyncMap[string, int]()
	m.Store("hotkey", 42)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Load("hotkey")
		}
	})
}

func BenchmarkMyRWMap_ReadOnly(b *testing.B) {
	m := NewMyRWMap[string, int]()
	m.Store("hotkey", 42)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Load("hotkey")
		}
	})
}

func BenchmarkMySyncMap_WriteOnly(b *testing.B) {
	m := NewMySyncMap[string, int]()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			m.Store("key"+strconv.Itoa(i), i)
		}
	})
}

func BenchmarkMyRWMap_WriteOnly(b *testing.B) {
	m := NewMyRWMap[string, int]()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			m.Store("key"+strconv.Itoa(i), i)
		}
	})
}

/*********热点读写(同一 key)*********/

func BenchmarkMySyncMap_ReadHotKey(b *testing.B) {
	m := NewMySyncMap[string, int]()
	m.Store("hotkey", 0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			m.Load("hotkey")
			m.Store("hotkey", i)
		}
	})
}

func BenchmarkMyRWMap_ReadHotKey(b *testing.B) {
	m := NewMyRWMap[string, int]()
	m.Store("hotkey", 0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			m.Load("hotkey")
			m.Store("hotkey", i)
		}
	})
}

/*****************混合并发访问(高并发读写不同 key)**********************/

func BenchmarkMySyncMap_MixedAccess(b *testing.B) {
	m := NewMySyncMap[string, int]()
	for i := 0; i < 100; i++ {
		m.Store("key"+strconv.Itoa(i), i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			key := "key" + strconv.Itoa(i%100)
			m.Load(key)
			m.Store(key, i)
		}
	})
}

func BenchmarkMyRWMap_MixedAccess(b *testing.B) {
	m := NewMyRWMap[string, int]()
	for i := 0; i < 100; i++ {
		m.Store("key"+strconv.Itoa(i), i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for i := 0; pb.Next(); i++ {
			key := "key" + strconv.Itoa(i%100)
			m.Load(key)
			m.Store(key, i)
		}
	})
}
