package sortedArray

import (
	"cmp"
	"sort"
)

// 默认顺序就是 升序排列

type RiseSortedArray[T cmp.Ordered] struct {
	data []T
}

// 创建有序数组,支持预分配容量
func NewRise[T cmp.Ordered](cap int) *RiseSortedArray[T] {
	return &RiseSortedArray[T]{
		data: make([]T, 0, cap),
	}
}

// 插入一个元素,保持有序
func (sa *RiseSortedArray[T]) Insert(x T) {
	// 二分查找插入位置
	idx := sort.Search(len(sa.data), func(i int) bool {
		return sa.data[i] >= x
	})

	// 插入
	sa.data = append(sa.data, *new(T)) // 扩容一位
	copy(sa.data[idx+1:], sa.data[idx:])
	sa.data[idx] = x
}

// 查找某个值，返回索引和是否存在
func (sa *RiseSortedArray[T]) Find(x T) (int, bool) {
	idx := sort.Search(len(sa.data), func(i int) bool {
		return sa.data[i] >= x
	})
	if idx < len(sa.data) && sa.data[idx] == x {
		return idx, true
	}
	return -1, false
}

// 删除某个值,返回是否删除成功
func (sa *RiseSortedArray[T]) Delete(x T) bool {
	idx, ok := sa.Find(x)
	if !ok {
		return false
	}
	copy(sa.data[idx:], sa.data[idx+1:])
	sa.data = sa.data[:len(sa.data)-1]
	return true
}

// 获取所有值(只读)
func (sa *RiseSortedArray[T]) Values() []T {
	return sa.data
}

// 获取最大值
func (sa *RiseSortedArray[T]) Max() (T, bool) {
	if len(sa.data) == 0 {
		var zero T
		return zero, false
	}
	return sa.data[len(sa.data)-1], true
}
