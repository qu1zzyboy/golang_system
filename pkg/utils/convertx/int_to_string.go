package convertx

// fallthrough 让所有case连续执行,实现从n执行到1,连续无分支指令,极大减少 CPU 跳转、循环开销
func Itoa(buf []byte, n int, v int64) int {
	if len(buf) < n+1 {
		panic("buffer too small")
	}
	neg := v < 0
	if neg {
		v = -v
	}
	p := n - 1
	switch n {
	case 20:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 19:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 18:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 17:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 16:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 15:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 14:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 13:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 12:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 11:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 10:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 9:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 8:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 7:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 6:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 5:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 4:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 3:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 2:
		buf[p] = '0' + byte(v%10)
		v /= 10
		p--
		fallthrough
	case 1:
		buf[p] = '0' + byte(v%10)
	}
	if neg {
		// 左移一位，加上负号
		copy(buf[1:], buf[:n])
		buf[0] = '-'
		return n + 1
	}
	return n
}

// 封装版,返回 string
func ItoaString(v int64, width int) string {
	var buf [21]byte
	n := Itoa(buf[:], width, v)
	return string(buf[:n])
}
