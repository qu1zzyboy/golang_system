package time2str

const (
	TSNano  = 0      //= 0,二进制:0000
	TSSecs  = 1 << 0 //= 1,二进制:0001
	TSMilli = 1 << 1 //= 2,二进制:0010
	TSMicro = 1 << 2 //= 4,二进制:0100
)

func GetNumDigits(v int64) int {
	n := 1
	if v >= 100000000000000000 {
		v /= 100000000000000000
		n += 17
	}
	if v >= 100000000 {
		v /= 100000000
		n += 8
	}
	if v >= 10000 {
		v /= 10000
		n += 4
	}
	if v >= 100 {
		v /= 100
		n += 2
	}
	if v >= 10 {
		n++
	}
	return n
}

// Utoa fallthrough 让所有case连续执行,实现从n执行到1,连续无分支指令,极大减少 CPU 跳转、循环开销
func Utoa(buf []byte, n int, v int64) {
	if len(buf) < n {
		panic("buffer too small")
	}
	// '0' + byte(1) = '1',也就是ASCII码 49
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
	// 	p:12,v:174159463055,res:3
	// p:11,v:17415946305,res:53
	// p:10,v:1741594630,res:553
	// p:9,v:174159463,res:0553
	// p:8,v:17415946,res:30553
	// p:7,v:1741594,res:630553
	// p:6,v:174159,res:4630553
	// p:5,v:17415,res:94630553
	// p:4,v:1741,res:594630553
	// p:3,v:174,res:1594630553
	// p:2,v:17,res:41594630553
	// p:1,v:1,res:741594630553
	// p:0,v:1,res:1741594630553
}

func Uint64ToString(v int64) string {
	n := GetNumDigits(v)
	if n == 13 {
		var buf [13]byte
		i2slice13(buf[:], v)
		return string(buf[:])
	}
	var buf [21]byte
	Utoa(buf[:], n, v)
	return string(buf[:n])
}

// TimestampToChars 将时间戳转换为十进制字符串,写入buf中,返回写入的长度
func TimestampToChars(buf []byte, ts int64, flag int) int {
	//按位与比较,二进制数都为1则为1,否则为0
	switch {
	case flag&TSMicro != 0:
		ts /= 1e3
	case flag&TSMilli != 0:
		ts /= 1e6
	case flag&TSSecs != 0:
		ts /= 1e9
	}
	n := GetNumDigits(ts)
	Utoa(buf, n, ts)
	return n
}

// TimestampToString 返回时间戳的字符串表示(string 类型)
func TimestampToString(ts int64, flag int) string {
	var buf [21]byte // uint64 最大 20 位 + \0
	n := TimestampToChars(buf[:], ts, flag)
	return string(buf[:n])
}
