package convertx

import (
	"math"
)

// goos: windows
// goarch: amd64
// pkg: QuantGo/pkg/utils/convertx
// cpu: Intel(R) Core(TM) i7-14700KF
// BenchmarkPriceStringToInt64_Original-28 87233392 13.31 ns/op 0 B/op 0 allocs/op
// BenchmarkPriceStringToInt64_Optimized-28 120938542 9.938 ns/op 0 B/op 0 allocs/op
// BenchmarkPriceStringToInt64_ParseFloat-28 59633254 20.19 ns/op 0 B/op 0 allocs/op

func PriceF64ToInt64(f float64, scale int32) uint64 {
	return uint64(f * math.Pow10(int(scale)))
}

func PriceStringToUint64(s string, scaleInput int32) uint64 {
	var v uint64
	var scale int32
	pastDot := false

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '.' {
			pastDot = true
			continue
		}
		if c < '0' || c > '9' {
			continue
		}
		v = v*10 + uint64(c-'0')
		if pastDot {
			scale++
		}
	}

	for scale < scaleInput {
		v *= 10
		scale++
	}

	for scale > scaleInput && v%10 == 0 {
		v /= 10
		scale--
	}
	return v
}
