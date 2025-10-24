package timeUtils

import (
	"time"

	"upbitBnServer/internal/define/defineTime"
)

// GetSecTimeStrBy 根据时间戳获取秒级时间字符串
func GetSecTimeStrBy(timeStamp int64) string {
	return time.Unix(timeStamp/1000, 0).Format(defineTime.FormatSec)
}

// GetMillSecTimeStrBy 根据时间戳获取毫秒级时间字符串
func GetMillSecTimeStrBy(timeStamp int64) string {
	return time.UnixMilli(timeStamp).Format(defineTime.FormatMillSec)
}

func GetHourKey(ts int64) int64 {
	t := time.UnixMilli(ts)
	return int64(t.Year()*1000000 + int(t.Month())*10000 + t.Day()*100 + t.Hour())
}
