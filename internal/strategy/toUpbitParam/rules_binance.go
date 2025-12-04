package toUpbitParam

import "math"

var gainBucketsBinance = []gainBucket{
	{lower: 0, upper: 5, baseline: 40, cap: 55},
	{lower: 5, upper: 20, baseline: 40, cap: 55},
	{lower: 20, upper: 30, baseline: 40, cap: 55},
	{lower: 30, upper: 40, baseline: 30, cap: 40},
	{lower: 40, upper: 50, baseline: 29, cap: 38},
	{lower: 50, upper: 60, baseline: 28, cap: 36},
	{lower: 60, upper: 70, baseline: 26, cap: 31},
	{lower: 70, upper: 80, baseline: 24, cap: 28},
	{lower: 80, upper: 90, baseline: 22, cap: 26},
	{lower: 90, upper: 100, baseline: 22, cap: 26},
	{lower: 100, upper: 120, baseline: 19, cap: 24},
	{lower: 120, upper: 140, baseline: 20, cap: 27},
	{lower: 140, upper: 330, baseline: 20, cap: 27},
	{lower: 330, upper: 600, baseline: 0.05, cap: 0.1},
}

var gainFallbackBinance = gainBucket{lower: 600, upper: math.MaxFloat64, baseline: 0.1, cap: 0.2}

var twapBucketsBinance = []twapBucket{
	{lower: 0, upper: 5, baseline: 60, cap: 90},
	{lower: 5, upper: 100, baseline: 45, cap: 65},
	{lower: 100, upper: 120, baseline: 35, cap: 52},
	{lower: 120, upper: 200, baseline: 45, cap: 67},
	{lower: 200, upper: 330, baseline: 45, cap: 67},
	{lower: 330, upper: 600, baseline: 10, cap: 15},
}

var twapFallbackBinance = twapBucket{lower: 600, upper: math.MaxFloat64, baseline: 5, cap: 10}

const (
	fearGreedHighAddBinance    = fearGreedHighAdd
	fearGreedLowSubBinance     = fearGreedLowSub
	btc1dMultiplierBinance     = btc1dMultiplier
	btc7dMultiplierBinance     = btc7dMultiplier
	fearGreedHighSubSecBinance = fearGreedHighSubSec
	fearGreedLowSubSecBinance  = fearGreedLowSubSec
	btc1dSecondsPerPctBinance  = btc1dSecondsPerPct
	btc7dSecondsPerPctBinance  = btc7dSecondsPerPct
	memeExtraSecondsBinance    = memeExtraSeconds
	memecoinGainAddBinance     = memecoinGainAdd
)

func pickGainBucketBinance(marketCap float64) gainBucket {
	for _, b := range gainBucketsBinance {
		if marketCap >= b.lower && marketCap < b.upper {
			return b
		}
	}
	return gainFallbackBinance
}

func pickTwapBucketBinance(marketCap float64) twapBucket {
	for _, b := range twapBucketsBinance {
		if marketCap >= b.lower && marketCap < b.upper {
			return b
		}
	}
	return twapFallbackBinance
}

func expectedSplitGainBinance(marketCapM, fearGreedIndex, btc1d, btc7d float64, isMeme bool) float64 {
	bucket := pickGainBucketBinance(marketCapM)
	score := bucket.baseline
	switch {
	case fearGreedIndex > 70:
		score += fearGreedHighAddBinance
	case fearGreedIndex < 40:
		score -= fearGreedLowSubBinance
	}
	score += btc1d * btc1dMultiplierBinance
	score += btc7d * btc7dMultiplierBinance
	if isMeme {
		score += memecoinGainAddBinance
	}
	if score < bucket.baseline {
		return bucket.baseline
	}
	if score > bucket.cap {
		return bucket.cap
	}
	return score
}

func expectedTwapDurationBinance(marketCapM, fearGreedIndex, btc1d, btc7d float64, isMeme bool) float64 {
	bucket := pickTwapBucketBinance(marketCapM)
	seconds := bucket.baseline
	switch {
	case fearGreedIndex > 70:
		seconds -= fearGreedHighSubSecBinance
	case fearGreedIndex < 40:
		seconds += fearGreedLowSubSecBinance
	}
	seconds += btc1d * btc1dSecondsPerPctBinance
	seconds += btc7d * btc7dSecondsPerPctBinance
	if isMeme {
		seconds += memeExtraSecondsBinance
	}
	if seconds < bucket.baseline {
		return bucket.baseline
	}
	if seconds > bucket.cap {
		return bucket.cap
	}
	return seconds
}

func expectedSplitGainAndTwapDurationBinance(marketCapM, fearGreedIndex, btc1d, btc7d float64, isMeme bool) (float64, float64) {
	return expectedSplitGainBinance(marketCapM, fearGreedIndex, btc1d, btc7d, isMeme),
		expectedTwapDurationBinance(marketCapM, fearGreedIndex, btc1d, btc7d, isMeme)
}
