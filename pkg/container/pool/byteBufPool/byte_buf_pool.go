package byteBufPool

import "sync"

const (
	SIZE_64   = 64
	SIZE_128  = 128
	SIZE_256  = 256
	SIZE_512  = 512
	SIZE_1024 = 1024
)

//1、减少 GC 压力,降低延迟抖动
//2、复用内存,提高吞吐量

// Go 的 make([]byte, 0, N) 内存粒度是按 64、128、256、512、1024... 分配对齐的

var (
	pool64  = sync.Pool{New: func() any { buf := make([]byte, 0, SIZE_64); return &buf }}
	pool128 = sync.Pool{New: func() any { buf := make([]byte, 0, SIZE_128); return &buf }}
	pool256 = sync.Pool{New: func() any { buf := make([]byte, 0, SIZE_256); return &buf }}
	pool512 = sync.Pool{New: func() any { buf := make([]byte, 0, SIZE_512); return &buf }}
	pool1k  = sync.Pool{New: func() any { buf := make([]byte, 0, SIZE_1024); return &buf }}
)

// AcquireBuffer 获取指定大小的缓冲区(注意是指容量,不是长度)
func AcquireBuffer(size int) *[]byte {
	switch {
	case size <= SIZE_64:
		return pool64.Get().(*[]byte)
	case size <= SIZE_128:
		return pool128.Get().(*[]byte)
	case size <= SIZE_256:
		return pool256.Get().(*[]byte)
	case size <= SIZE_512:
		return pool512.Get().(*[]byte)
	case size <= SIZE_1024:
		return pool1k.Get().(*[]byte)
	default:
		buf := make([]byte, 0, size) // 超过默认池容量,直接分配
		return &buf
	}
}

// ReleaseBuffer 回收缓冲区到对应池
func ReleaseBuffer(buf *[]byte) {
	if buf == nil {
		return
	}
	capacity := cap(*buf)
	*buf = (*buf)[:0] // reset slice
	switch {
	case capacity <= SIZE_64:
		pool64.Put(buf)
	case capacity <= SIZE_128:
		pool128.Put(buf)
	case capacity <= SIZE_256:
		pool256.Put(buf)
	case capacity <= SIZE_512:
		pool512.Put(buf)
	case capacity <= SIZE_1024:
		pool1k.Put(buf)
		// 超出不回收,防止滥用内存
	}
}
