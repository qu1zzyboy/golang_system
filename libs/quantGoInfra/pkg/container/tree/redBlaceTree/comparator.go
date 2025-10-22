package redBlaceTree

func BnDepthComparator(x, y string) int {
	if len(x) < len(y) {
		return -1
	} else if len(x) > len(y) {
		return 1
	} else {
		if x < y {
			return -1
		} else if x > y {
			return 1
		}
	}
	return 0
}
