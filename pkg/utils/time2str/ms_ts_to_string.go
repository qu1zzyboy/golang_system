package time2str

import (
	"time"
)

const msArrLen = 15

// 高性能固定 13 位 Uint64 到字符串（适用于毫秒级时间戳）
func i2slice13(buf []byte, v int64) {
	buf[12] = '0' + byte(v%10)
	v /= 10
	buf[11] = '0' + byte(v%10)
	v /= 10
	buf[10] = '0' + byte(v%10)
	v /= 10
	buf[9] = '0' + byte(v%10)
	v /= 10
	buf[8] = '0' + byte(v%10)
	v /= 10
	buf[7] = '0' + byte(v%10)
	v /= 10
	buf[6] = '0' + byte(v%10)
	v /= 10
	buf[5] = '0' + byte(v%10)
	v /= 10
	buf[4] = '0' + byte(v%10)
	v /= 10
	buf[3] = '0' + byte(v%10)
	v /= 10
	buf[2] = '0' + byte(v%10)
	v /= 10
	buf[1] = '0' + byte(v%10)
	v /= 10
	buf[0] = '0' + byte(v%10)
}

//	我的:19.33 ns/op           16 B/op          1 allocs/op
//	获取时间戳 2.217 ns/op           0 B/op          0 allocs/op
//
// 标准库:21.25 ns/op           16 B/op          1 allocs/op
func GetNowTimeStampMilliStr() string {
	var buf [msArrLen]byte
	i2slice13(buf[:], time.Now().UnixMilli())
	return string(buf[:])
}

//47.56 ns/op	       0 B/op	       0 allocs/op

func GetNowTimeStampMillSlice14() [msArrLen]byte {
	var buf [msArrLen]byte
	i2slice13(buf[:], time.Now().UnixMilli())
	return buf
}

func GetTimeStampMilliStrBy(timeStamp int64) string {
	var buf [msArrLen]byte
	i2slice13(buf[:], timeStamp)
	return string(buf[:])
}
