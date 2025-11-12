package byteConvert

import (
	"bytes"
)

func BytesToInt64(b []byte) int64 {
	var n int64
	for i := range b {
		c := b[i]
		n = n*10 + int64(c-'0')
	}
	return n
}

// 6.687 ns/op	       0 B/op	       0 allocs/op

func PriceByteArrToUint64(b []byte, scaleInput uint8) uint64 {
	var v uint64
	var scale uint8
	pastDot := false
	for i := range b {
		c := b[i]
		switch {
		case c == '.':
			pastDot = true
		case c >= '0' && c <= '9':
			v = v*10 + uint64(c-'0')
			if pastDot {
				scale++
			}
		default:
			break
		}
	}
	for scale < scaleInput {
		v *= 10
		scale++
	}
	for scale > scaleInput {
		v /= 10
		scale--
	}
	return v
}

func FastSearchByte(b []byte, key string) ([]byte, int, int) {
	k := []byte(`"` + key + `":"`)
	pos := bytes.Index(b, k)
	if pos == -1 {
		return nil, 0, 0
	}
	startIdx := pos + len(k)
	endIdx := bytes.IndexByte(b[startIdx:], '"')
	if endIdx == -1 {
		return nil, 0, 0
	}
	return b[startIdx : startIdx+endIdx], startIdx, endIdx
}

func FastSearchPriceUint64(b []byte, key string, scaleInput uint8) uint64 {
	k := []byte(`"` + key + `":"`)
	pos := bytes.Index(b, k)
	if pos == -1 {
		return 0
	}
	startIdx := pos + len(k)
	endIdx := bytes.IndexByte(b[startIdx:], '"')
	if endIdx == -1 {
		return 0
	}
	return PriceByteArrToUint64(b[startIdx:startIdx+endIdx], scaleInput)
}
