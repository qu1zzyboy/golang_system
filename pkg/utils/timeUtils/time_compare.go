package timeUtils

// IsTimeOut 给定的时间戳是否超时(ms),超时为true,未超时为false
func IsTimeOut(timeOutStepMilli int64, compareTimeStamp int64) bool {
	return GetNowTimeUnixMilli()-compareTimeStamp > timeOutStepMilli
}
