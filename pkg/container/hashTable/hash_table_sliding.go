package hashTable

import "sync/atomic"

// SlidingWindow 使用 N 个轮换的哈希表，实现“最近 N 个时间片”的唯一性判定。
// 完全无锁+并发安全
// 典型用法：
//   - num = 2：最近 2 秒内唯一
//   - num = 3：最近 3 秒内唯一
//
// 调用方通常使用定时器（例如每秒）调用 Rotate，使窗口向前滑动。
type SlidingWindow[K Key] struct {
	tables []*Table[K]   // 环形数组,tables[0..count-1]
	index  atomic.Uint32 // 当前最新窗口的索引
	count  int           // 窗口总数
}

// NewSlidingWindow 创建一个滑动窗口唯一性检查器。
// num 为窗口表个数；capacity 为每个表的容量（必须为 2 的幂）。
func NewSlidingWindow[K Key](num int, capacity int, hash HashFunc[K]) *SlidingWindow[K] {
	sw := &SlidingWindow[K]{count: num}
	sw.tables = make([]*Table[K], num)
	for i := range num {
		sw.tables[i] = NewTable(capacity, hash)
	}
	sw.index.Store(0)
	return sw
}

// ExistsOrInsert 会在所有窗口中检查键是否存在：
// - 如在任意窗口中找到该键，则返回 true（表示重复）；
// - 否则，将该键插入到当前最新窗口，并返回 false（首次出现）。
func (sw *SlidingWindow[K]) ExistsOrInsert(k K) bool {
	newest := int(sw.index.Load())

	// 从最新窗口开始，向历史窗口逐个检查
	for i := 0; i < sw.count; i++ {
		// 倒序遍历+避免负数
		idx := (newest - i + sw.count) % sw.count
		if sw.tables[idx].Exists(k) {
			return true
		}
	}

	sw.tables[newest].Insert(k)
	return false
}

// Rotate 将滑动窗口向前移动一格：
// - 找到下一个窗口索引
// - 清空该窗口对应的表
// - 将其标记为最新窗口
// 一般由定时任务按照固定时间片（如 1 秒）调用。
func (sw *SlidingWindow[K]) Rotate() {
	next := (int(sw.index.Load()) + 1) % sw.count
	sw.tables[next].Clear()
	sw.index.Store(uint32(next))
}

// Capacity 返回单个窗口表的容量。
func (sw *SlidingWindow[K]) Capacity() int {
	if len(sw.tables) == 0 {
		return 0
	}
	return len(sw.tables[0].slots)
}
