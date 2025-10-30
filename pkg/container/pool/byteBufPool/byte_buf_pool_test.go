package byteBufPool

import (
	"sync"
	"testing"
)

var p = sync.Pool{New: func() any {
	buf := make([]byte, 1024)
	return &buf
}}

func BenchmarkPointer(b *testing.B) {
	for b.Loop() {
		bufPtr := p.Get().(*[]byte)
		b := *bufPtr
		_ = b[:1]
		p.Put(bufPtr)
	}
}

var q = sync.Pool{New: func() any {
	buf := make([]byte, 1024)
	return buf
}}

func BenchmarkSlice(b *testing.B) {
	for b.Loop() {
		buf := q.Get().([]byte)
		_ = buf[:1]
		q.Put(buf)
	}
}
