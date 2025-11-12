package byteUtils

// copyScaledValue 将定点整数 val 按 scale 小数位直接写入 dst[start:]
// 例如 val=12345678, scale=4 => 写入 "1234.5678"

func CopyScaledValue(dst []byte, start uint16, val uint64, scale int) uint16 {
	// fast path: scale==0，无小数
	if scale == 0 {
		return writeUint(dst, start, val)
	}
	// 小数部分（右对齐,scale <= 8
	var frac [8]byte
	for i := scale - 1; i >= 0; i-- {
		frac[i] = byte('0' + val%10)
		val /= 10
	}
	// 整数部分
	end := writeUint(dst, start, val)
	dst[end] = '.'
	copy(dst[end+1:], frac[:scale])
	return end + 1 + uint16(scale)
}

// writeUint 把整数 val 转为十进制写入 dst[start:], 返回写入后的位置。
func writeUint(dst []byte, start uint16, val uint64) uint16 {
	if val == 0 {
		dst[start] = '0'
		return start + 1
	}
	// 临时缓冲反向写
	var tmp [9]byte // ✅ 足够安全（支持到 999,999,999）
	i := len(tmp)
	for val > 0 {
		i--
		tmp[i] = byte('0' + val%10)
		val /= 10
	}
	//1	19	'4'	[ '4' ]
	//2	18	'3'	[ '3','4' ]
	//3	17	'2'	[ '2','3','4' ]
	//4	16	'1'	[ '1','2','3','4' ]
	n := uint16(len(tmp) - i)
	copy(dst[start:], tmp[i:]) //自动只会复制最短的长度,不需要写end
	return start + n
}
