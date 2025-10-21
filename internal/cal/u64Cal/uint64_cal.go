package u64Cal

// IsDiffMoreThanPercent100 returns true if |a-b| > 1% of max(a,b)
func IsDiffMoreThanPercent100(a, b uint64, percent uint64) bool {
	if a == b {
		return false
	}
	// 特判0,避免除零或无意义比较
	if a == 0 || b == 0 {
		return false
	}
	// 获取最大值和差值(仅1个分支)
	if a < b {
		a, b = b, a
	}
	diff := a - b
	return diff*100 > a*percent
}
