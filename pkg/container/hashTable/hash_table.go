package hashTable

// 低延迟、固定容量、开放寻址哈希表（通用泛型版）

import "sync/atomic"

// Key 表示用户自定义的、可比较的键类型（可作为 map 的 key）
type Key interface{ comparable }

// HashFunc 是用户提供的 64 位哈希函数，用于将 K 映射为 uint64
type HashFunc[K Key] func(K) uint64

// Slot 表示开放寻址哈希表中的一个槽位。
// Used 为原子标志位：0 表示空槽；1 表示槽已被占用。性能最高（内联原子操作） Slot 最小（cache 效率极高）内存布局最紧凑（无 noCopy）
type Slot[K Key] struct {
	Key  K
	Used uint32 // uint32 原子标记：0 空 / 1 已使用
}

// Table 是固定容量的开放寻址哈希表，使用线性探测解决冲突。
// - slots 长度固定，capacity 必须是 2 的幂
// - mask = capacity - 1，用于快速取模（与运算）
// - size 仅用于统计当前使用槽数量，不参与功能逻辑
// - hash 由调用方提供，保证同一 K 总是映射到相同的 uint64

type Table[K Key] struct {
	slots []Slot[K]   // 槽位数组
	mask  uint64      // 掩码 = len(slots) - 1
	size  uint64      // 已使用槽数量（近似元素个数）
	hash  HashFunc[K] // 用户提供的哈希函数
}

// NewTable 创建一个固定容量的哈希表。
// 注意：capacity 必须是 2 的幂（例如 1<<20）。
func NewTable[K Key](capacity int, hash HashFunc[K]) *Table[K] {
	return &Table[K]{
		slots: make([]Slot[K], capacity),
		mask:  uint64(capacity - 1),
		hash:  hash,
	}
}

// Exists 判断键是否存在。
// 这是无锁方法，适合在多协程环境下并发调用（仅读取原子位和只读 Key）。
func (t *Table[K]) Exists(k K) bool {
	index := t.hash(k) & t.mask
	for {
		slot := &t.slots[index]
		if atomic.LoadUint32(&slot.Used) == 0 {
			// 遇到空槽，说明探测链到此结束，键不存在
			return false
		}
		if slot.Key == k {
			return true
		}
		index = (index + 1) & t.mask
	}
}

// Insert 插入键：
// - 若探测到空槽，则通过 CAS 抢占该槽；
// - 若发现槽中已是相同 Key，则视为已存在，直接返回（幂等）；
// - 在高并发下可能需要多次重试不同槽位。
func (t *Table[K]) Insert(k K) {
	index := t.hash(k) & t.mask
	for {
		slot := &t.slots[index]
		used := atomic.LoadUint32(&slot.Used)
		if used == 0 {
			// 先抢占槽位，再写入 Key，保证其他协程看到 Used=1 时，Key 已经写入
			if atomic.CompareAndSwapUint32(&slot.Used, 0, 1) {
				slot.Key = k
				atomic.AddUint64(&t.size, 1)
				return
			}
			// CAS 失败，说明有其他协程抢先占用该槽，继续线性探测
		} else if slot.Key == k {
			// 已存在相同 Key，直接返回，保持幂等
			return
		}
		index = (index + 1) & t.mask
	}
}

// Clear 清空表，将所有槽标记为空。
// 一般用于“滑动窗口轮换”时，将最旧窗口的数据整体丢弃。
func (t *Table[K]) Clear() {
	for i := range t.slots {
		atomic.StoreUint32(&t.slots[i].Used, 0)
	}
	atomic.StoreUint64(&t.size, 0)
}

// Size 返回当前表中已使用的槽数量（近似元素数量，仅供监控/统计使用）。
func (t *Table[K]) Size() uint64 {
	return atomic.LoadUint64(&t.size)
}
