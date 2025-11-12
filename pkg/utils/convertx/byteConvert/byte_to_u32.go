package byteConvert

import "upbitBnServer/internal/infra/systemx"

// 6.687 ns/op	       0 B/op	       0 allocs/op

func PriceByteArrToUint32(b []byte, scaleInput uint8) systemx.ScaledValue {
	var v systemx.ScaledValue
	var scale uint8
	pastDot := false
	for i := range b {
		c := b[i]
		switch {
		case c == '.':
			pastDot = true
		case c >= '0' && c <= '9':
			v = v*10 + systemx.ScaledValue(c-'0')
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
