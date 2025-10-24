package timeUtils

import (
	"time"

	"upbitBnServer/internal/define/defineTime"
)

// 获取当前时间的日期形式的时间str(2025-07-07)
func GetNowDateStr() string {
	return time.Now().Format(defineTime.FormatDate)
}

func GetNowDateHourStr() string {
	return time.Now().Format(defineTime.FormatHour)
}

// 获取当前时间的毫秒级时间戳
func GetNowTimeUnixMilli() int64 {
	return time.Now().UnixMilli()
}

// 获取当前时间的微秒级时间戳
func GetNowTimeUnixMicro() int64 {
	return time.Now().UnixMicro()
}

// 我的 33.24 ns/op           24 B/op          1 allocs/op
// 标准库 95.88 ns/op	      24 B/op	       1 allocs/op
func GetNowMillisDateStrFast() string {
	var buf [23]byte // "YYYY-MM-DD HH:MM:SS.mmm" 固定 23 位
	t := time.Now()
	y, m, d := t.Date()
	h, min_, s := t.Clock()
	ms := t.Nanosecond() / 1e6
	write4Digits_(buf[0:], y) // YYYY
	buf[4] = '-'
	write2Digits_(buf[5:], int(m)) // MM
	buf[7] = '-'
	write2Digits_(buf[8:], d) // DD
	buf[10] = ' '
	write2Digits_(buf[11:], h) // HH
	buf[13] = ':'
	write2Digits_(buf[14:], min_) // mm
	buf[16] = ':'
	write2Digits_(buf[17:], s) // ss
	buf[19] = '.'
	write3Digits_(buf[20:], ms) // mmm
	return string(buf[:])
}

func write2Digits_(buf []byte, v int) {
	buf[0] = '0' + byte(v/10)
	buf[1] = '0' + byte(v%10)
}

func write3Digits_(buf []byte, v int) {
	buf[0] = '0' + byte((v/100)%10)
	buf[1] = '0' + byte((v/10)%10)
	buf[2] = '0' + byte(v%10)
}

func write4Digits_(buf []byte, v int) {
	buf[0] = '0' + byte((v/1000)%10)
	buf[1] = '0' + byte((v/100)%10)
	buf[2] = '0' + byte((v/10)%10)
	buf[3] = '0' + byte(v%10)
}
