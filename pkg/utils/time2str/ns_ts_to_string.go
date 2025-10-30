package time2str

import (
	"time"
)

// 高性能固定 19 位 Uint64 到字符串（适用于毫秒级时间戳）

func I2str19(buf []byte, v int64) {
	buf[18] = '0' + byte(v%10)
	v /= 10
	buf[17] = '0' + byte(v%10)
	v /= 10
	buf[16] = '0' + byte(v%10)
	v /= 10
	buf[15] = '0' + byte(v%10)
	v /= 10
	buf[14] = '0' + byte(v%10)
	v /= 10
	buf[13] = '0' + byte(v%10)
	v /= 10
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

func GetNowTimeStampNanoStr() string {
	var buf [19]byte
	I2str19(buf[:], time.Now().UnixNano())
	return string(buf[:])
}

func GetTimeStampNanoStrBy(timeStamp int64) string {
	var buf [19]byte
	I2str19(buf[:], timeStamp)
	return string(buf[:])
}
