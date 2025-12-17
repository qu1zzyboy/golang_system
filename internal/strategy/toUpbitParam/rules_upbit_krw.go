package toUpbitParam

import "math"

var (
	gainBuckets = []gainBucket{
		{lower: 0, upper: 5, baseline: 40, cap: 55},
		{lower: 5, upper: 20, baseline: 38, cap: 55},
		{lower: 20, upper: 30, baseline: 34, cap: 55},
		{lower: 30, upper: 40, baseline: 30, cap: 40},
		{lower: 40, upper: 50, baseline: 29, cap: 33},
		{lower: 50, upper: 60, baseline: 28, cap: 32},
		{lower: 60, upper: 70, baseline: 26, cap: 31},
		{lower: 70, upper: 80, baseline: 24, cap: 28},
		{lower: 80, upper: 90, baseline: 22, cap: 26},
		{lower: 90, upper: 100, baseline: 22, cap: 26},
		{lower: 100, upper: 120, baseline: 19, cap: 24},
		{lower: 120, upper: 140, baseline: 20, cap: 27},
		{lower: 140, upper: 330, baseline: 20, cap: 27},
		{lower: 330, upper: 600, baseline: 0.05, cap: 0.1},
	}
	gainFallback = gainBucket{lower: 600, upper: math.MaxFloat64, baseline: 0.1, cap: 0.2}

	twapBuckets = []twapBucket{
		{lower: 0, upper: 5, baseline: 60, cap: 90},
		{lower: 5, upper: 20, baseline: 60, cap: 90},
		{lower: 20, upper: 30, baseline: 56, cap: 90},
		{lower: 30, upper: 40, baseline: 55, cap: 88},
		{lower: 40, upper: 50, baseline: 54, cap: 86},
		{lower: 50, upper: 60, baseline: 53, cap: 84},
		{lower: 60, upper: 70, baseline: 52, cap: 82},
		{lower: 70, upper: 80, baseline: 51, cap: 80},
		{lower: 80, upper: 90, baseline: 49, cap: 75},
		{lower: 90, upper: 100, baseline: 47, cap: 70},
		{lower: 100, upper: 120, baseline: 45, cap: 65},
		{lower: 120, upper: 140, baseline: 35, cap: 52},
		{lower: 140, upper: 330, baseline: 45, cap: 67},
		{lower: 330, upper: 600, baseline: 10, cap: 15},
	}
)

func pickGainBucketUpbitKrw(marketCap float64) gainBucket {
	for _, b := range gainBuckets {
		if marketCap >= b.lower && marketCap < b.upper {
			return b
		}
	}
	return gainFallback
}

func pickTwapBucketUpbitKrw(marketCap float64) twapBucket {
	for _, b := range twapBuckets {
		if marketCap >= b.lower && marketCap < b.upper {
			return b
		}
	}
	return twapFallback
}

func expectedSplitGainAndTwapDuration(marketCapM, fearGreedIndex, btc1d, btc7d float64, isMeme bool) (float64, float64) {
	return expectedSplitGain(marketCapM, fearGreedIndex, btc1d, btc7d, isMeme),
		expectedTwapDuration(marketCapM, fearGreedIndex, btc1d, btc7d, isMeme)
}

func expectedSplitGain(marketCapM, fearGreedIndex, btc1d, btc7d float64, isMeme bool) float64 {
	bucket := pickGainBucket(marketCapM)
	score := bucket.baseline
	switch {
	case fearGreedIndex > 70:
		score += fearGreedHighAdd
	case fearGreedIndex < 40:
		score -= fearGreedLowSub
	}
	score += btc1d * btc1dMultiplier
	score += btc7d * btc7dMultiplier
	if isMeme {
		score += memecoinGainAdd
	}
	if score < bucket.baseline {
		return bucket.baseline
	}
	if score > bucket.cap {
		return bucket.cap
	}
	return score
}

func expectedTwapDuration(marketCapM, fearGreedIndex, btc1d, btc7d float64, isMeme bool) float64 {
	bucket := pickTwapBucket(marketCapM)
	seconds := bucket.baseline
	switch {
	case fearGreedIndex > 70:
		seconds -= fearGreedHighSubSec
	case fearGreedIndex < 40:
		seconds += fearGreedLowSubSec
	}
	seconds += btc1d * btc1dSecondsPerPct
	seconds += btc7d * btc7dSecondsPerPct
	if isMeme {
		seconds += memeExtraSeconds
	}
	if seconds < bucket.baseline {
		return bucket.baseline
	}
	if seconds > bucket.cap {
		return bucket.cap
	}
	return seconds
}
