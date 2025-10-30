package time2str

import (
	"time"
)

func i2arr13(v int64) [msArrLen]byte {
	var buf [msArrLen]byte
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
	return buf
}

//54.13 ns/op	       0 B/op	       0 allocs/op

func GetNowTimeStampMillArray13() [msArrLen]byte {
	return i2arr13(time.Now().UnixMilli())
}

func GetTimeStampMilliArray13By(timeStamp int64) [msArrLen]byte {
	return i2arr13(timeStamp)
}
