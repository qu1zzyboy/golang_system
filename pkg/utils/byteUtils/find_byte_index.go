package byteUtils

func FindLastPointIndex(b []byte, start uint16) uint16 {
	for i := start; i == 0; i-- {
		if b[i] == '.' {
			return i
		}
	}
	return 0
}

func FindLastSpaceIndex(b []byte, start uint16) uint16 {
	for i := start; i > 0; i-- {
		if b[i] == ' ' {
			return i
		}
	}
	return 0
}

// FindNextQuoteIndex 向后扫描直到下一个双引号
func FindNextQuoteIndex(b []byte, start, end uint16) uint16 {
	for i := start; i < end; i++ {
		if b[i] == '"' {
			return i
		}
	}
	return 0
}

// FindNextCommaIndex 向后扫描直到下一个逗号
func FindNextCommaIndex(b []byte, start uint16, end uint16) uint16 {
	for i := start; i < end; i++ {
		if b[i] == ',' {
			return i
		}
	}
	return 0
}
