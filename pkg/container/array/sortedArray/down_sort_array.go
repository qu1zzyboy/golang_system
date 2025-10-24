package sortedArray

import (
	"cmp"
	"sort"
)

type DownSortedArray[T cmp.Ordered] struct {
	data []T
}

// 创建有序数组（降序），支持预分配容量
func NewDown[T cmp.Ordered](cap int) *DownSortedArray[T] {
	return &DownSortedArray[T]{
		data: make([]T, 0, cap),
	}
}

// 插入一个元素,保持降序
func (sa *DownSortedArray[T]) Insert(x T) {
	// 二分查找插入位置
	idx := sort.Search(len(sa.data), func(i int) bool {
		return sa.data[i] <= x // 找到第一个 <= x 的位置
	})

	// 插入
	sa.data = append(sa.data, *new(T)) // 扩容一位
	copy(sa.data[idx+1:], sa.data[idx:])
	sa.data[idx] = x
}

// 查找某个值,返回索引和是否存在
func (sa *DownSortedArray[T]) Find(x T) (int, bool) {
	idx := sort.Search(len(sa.data), func(i int) bool {
		return sa.data[i] <= x
	})
	if idx < len(sa.data) && sa.data[idx] == x {
		return idx, true
	}
	return -1, false
}

// 删除某个值,返回是否删除成功
func (sa *DownSortedArray[T]) Delete(x T) bool {
	idx, ok := sa.Find(x)
	if !ok {
		return false
	}
	copy(sa.data[idx:], sa.data[idx+1:])
	sa.data = sa.data[:len(sa.data)-1]
	return true
}

// 获取所有值(只读,降序排列)
func (sa *DownSortedArray[T]) Values() []T {
	return sa.data
}

// 获取最小值
func (sa *DownSortedArray[T]) Min() (T, bool) {
	if len(sa.data) == 0 {
		var zero T
		return zero, false
	}
	return sa.data[len(sa.data)-1], true
}
