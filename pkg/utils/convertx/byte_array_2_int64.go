package convertx

func BytesToInt64(b []byte) int64 {
	var n int64
	for i := range b {
		c := b[i]
		n = n*10 + int64(c-'0')
	}
	return n
}

func PriceByteArrToUint64(b []byte, scaleInput int32) uint64 {
	var v uint64
	var scale int32
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
