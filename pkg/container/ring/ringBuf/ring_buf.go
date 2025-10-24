package ringBuf

type Capacity int

const (
	Cap2     Capacity = 2
	Cap4     Capacity = 4
	Cap8     Capacity = 8
	Cap16    Capacity = 16
	Cap32    Capacity = 32
	Cap64    Capacity = 64
	Cap128   Capacity = 128
	Cap256   Capacity = 256
	Cap512   Capacity = 512
	Cap1024  Capacity = 1024
	Cap2048  Capacity = 2048
	Cap4096  Capacity = 4096
	Cap8192  Capacity = 8192
	Cap16384 Capacity = 16384
	Cap32768 Capacity = 32768
	Cap65536 Capacity = 65536
)

type Ring[T any] struct {
	buf    []T
	head   int // 下一次写入位置
	filled int // 有效元素数
	mask   int // = cap-1,要求 cap 为 2 的幂
}

func NewPow2[T any](n Capacity) *Ring[T] {
	return &Ring[T]{buf: make([]T, n), mask: int(n) - 1}
}

func (r *Ring[T]) Capacity() int { return len(r.buf) } //r.buf的len一直是固定的n
func (r *Ring[T]) Size() int     { return r.filled }   //size是当前有效元素数
func (r *Ring[T]) Full() bool    { return r.filled == len(r.buf) }

func (r *Ring[T]) Clear() {
	// 如需更快让 GC 释放引用，可取消注释置零；默认只重置游标更快
	// for i := 0; i < r.filled; i++ { var z T; r.buf[i] = z }
	r.head = 0
	r.filled = 0
}

// Push 覆盖最老的,用按位与代替取模
func (r *Ring[T]) Push(v T) {
	// 把新元素 v 写到当前写入位置 head
	r.buf[r.head] = v
	// 将 head 向前移动 1 个位置,如果走到末尾，需要“回绕”到 0
	r.head = (r.head + 1) & r.mask
	// 更新当前有效元素数
	if r.filled < len(r.buf) {
		r.filled++
	}
}

// --- 读路径优化,两段式复制,避免逐元素循环 ---

// ToSlice 返回“最旧→最新”的快照(分配一次,最多两次 copy)
func (r *Ring[T]) ToSlice() []T {
	sz := r.filled
	if sz == 0 {
		return nil
	}
	out := make([]T, sz)

	// oldest 的起点 = head - sz(可能为负,用加容量修正)
	start := r.head - sz
	if start < 0 {
		start += len(r.buf)
	}

	// 第一段：从 start 到尾部
	n1 := len(r.buf) - start
	if n1 > sz {
		n1 = sz
	}
	copy(out, r.buf[start:start+n1])

	// 第二段：从 0 到剩余
	if n2 := sz - n1; n2 > 0 {
		copy(out[n1:], r.buf[:n2])
	}
	return out
}

// CopyTo 将内容(最旧→最新)追加到 dst,并返回 dst
func (r *Ring[T]) CopyTo(dst []T) []T {
	sz := r.filled
	if sz == 0 {
		return dst
	}
	dst = append(dst, make([]T, sz)...)
	base := len(dst) - sz

	start := r.head - sz
	if start < 0 {
		start += len(r.buf)
	}
	n1 := len(r.buf) - start
	if n1 > sz {
		n1 = sz
	}
	copy(dst[base:], r.buf[start:start+n1])
	if n2 := sz - n1; n2 > 0 {
		copy(dst[base+n1:], r.buf[:n2])
	}
	return dst
}

// ForEach 仍然提供，但比起 ToSlice/CopyTo 略慢(函数调用/边界检查更多)
func (r *Ring[T]) ForEach(f func(T)) {
	sz := r.filled
	if sz == 0 {
		return
	}
	start := r.head - sz
	if start < 0 {
		start += len(r.buf)
	}
	// 第一段
	for i := start; i < len(r.buf) && sz > 0; i++ {
		f(r.buf[i])
		sz--
	}
	// 第二段
	for i := 0; i < sz; i++ {
		f(r.buf[i])
	}
}
