package bnDriveCommon

// GetPreIndex 给定一个整数 i,算出它在 1–10 的“循环序列”里的前2个下标
func GetPreIndex(i int32) int32 {
	res := i%10 + 8
	if res > 10 {
		res = res - 10
	}
	return res
}

// GetCurIndex 把任意整数 i 映射到 1–10 之间
func GetCurIndex(i int32) int32 {
	res := i % 10
	if res == 0 {
		res = 10
	}
	return res
}
