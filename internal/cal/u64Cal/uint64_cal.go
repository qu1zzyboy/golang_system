package u64Cal

import "upbitBnServer/pkg/utils/pow10Utils"

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

func FromF64(f float64, scale uint8) uint64 {
	return uint64(f * pow10Utils.ToPowF64(scale))
}
