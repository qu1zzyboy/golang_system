package snapShot

import (
	"fmt"
	"testing"
	"time"
)

// BenchmarkIndexManager_Add-28
//   224199              4944 ns/op            1509 B/op         32 allocs/op

// BenchmarkIndexManager_Add 插入快照数据的性能测试
func BenchmarkIndexManager_Add(b *testing.B) {
	manager, err := NewIndexManager()
	if err != nil {
		b.Fatalf("failed to create IndexManager: %v", err)
	}
	defer manager.db.Close()

	startTs := time.Now().UnixMilli()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts := startTs + int64(i)
		val := []byte(fmt.Sprintf("snapshot-data-%d", i))

		if err := manager.Add(ts, val, 5*time.Minute); err != nil {
			b.Errorf("failed to add snapshot at %d: %v", ts, err)
		}
	}
}

// BenchmarkIndexManager_GetRange-28
//	15295             73464 ns/op           38282 B/op        825 allocs/op

// BenchmarkIndexManager_GetRange 查询快照范围的性能测试
func BenchmarkIndexManager_GetRange(b *testing.B) {
	manager, err := NewIndexManager()
	if err != nil {
		b.Fatalf("failed to create IndexManager: %v", err)
	}
	defer manager.db.Close()

	baseTs := time.Now().UnixMilli()

	// 预写入固定数量数据
	total := 5000
	for i := 0; i < total; i++ {
		ts := baseTs + int64(i)
		val := []byte(fmt.Sprintf("payload-%d", i))
		_ = manager.Add(ts, val, 10*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := baseTs + 1000
		end := baseTs + 1100

		results, err := manager.GetRange(start, end)
		if err != nil {
			b.Errorf("GetRange error: %v", err)
		}
		if len(results) == 0 {
			b.Errorf("unexpected empty result at iteration %d", i)
		}
	}
}
