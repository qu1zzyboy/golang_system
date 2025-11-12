package byteUtils

import (
	"strconv"
)

// AppendScaledValue 将定点整数值 val 按 scale 小数位输出到 dst。
// 10.59 ns/op	       0 B/op	       0 allocs/op
// 例如：
//
//	val=12345678, scale=4 => "1234.5678"
//	val=100,      scale=8 => "0.00000100"
//	val=123,      scale=0 => "123"
func AppendScaledValue(dst []byte, val uint64, scale int) []byte {
	//现代 64 位 CPU 的算术逻辑单元（ALU）和寄存器都是 64 位宽
	//即使用 uint8 或 uint32，CPU 实际仍会把它们“零扩展”为 64 位来运算
	//相比在 CPU 上多次零扩展、小寄存器搬运,直接在 Go 里用 uint64 / int 运算更快
	if scale == 0 {
		return strconv.AppendUint(dst, val, 10)
	}
	// 小数部分固定 8 字节缓冲（因为 scale≤8）
	var frac [8]byte
	for i := scale - 1; i >= 0; i-- {
		frac[i] = byte('0' + val%10)
		val /= 10
	}
	// 没有整数部分
	if val == 0 {
		dst = append(dst, "0."...)
		dst = append(dst, frac[:scale]...)
		return dst
	}
	// 有整数部分
	dst = strconv.AppendUint(dst, val, 10)
	dst = append(dst, '.')
	dst = append(dst, frac[:scale]...)
	return dst
}
