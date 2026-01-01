package toUpbitParam

import "math"

var gainBucketsBinanceLife = []gainBucket{
	{lower: 0, upper: 50, baseline: 40, cap: 55},
	{lower: 50, upper: 70, baseline: 40, cap: 55},
	{lower: 70, upper: 90, baseline: 40, cap: 55},
	{lower: 90, upper: 105, baseline: 40, cap: 55},
	{lower: 105, upper: 120, baseline: 40, cap: 55},
	{lower: 120, upper: 130, baseline: 40, cap: 50},
	{lower: 130, upper: 140, baseline: 32, cap: 42},
	{lower: 140, upper: 150, baseline: 25, cap: 32},
	{lower: 150, upper: 160, baseline: 25, cap: 32},
	{lower: 160, upper: 170, baseline: 25, cap: 32},
	{lower: 170, upper: 180, baseline: 25, cap: 32},
	{lower: 180, upper: 190, baseline: 25, cap: 32},
	{lower: 190, upper: 200, baseline: 25, cap: 32},
	{lower: 200, upper: 330, baseline: 25, cap: 32},
	{lower: 330, upper: 600, baseline: 0.05, cap: 0.1},
}

var gainFallbackBinanceLife = gainBucket{lower: 600, upper: math.MaxFloat64, baseline: 0.1, cap: 0.2}

var twapBucketsBinanceLife = []twapBucket{
	{lower: 0, upper: 5, baseline: 150, cap: 160},
	{lower: 5, upper: 100, baseline: 150, cap: 160},
	{lower: 100, upper: 120, baseline: 150, cap: 160},
	{lower: 120, upper: 200, baseline: 150, cap: 160},
	{lower: 200, upper: 330, baseline: 30, cap: 60},
	{lower: 330, upper: 600, baseline: 10, cap: 20},
}

var twapFallbackBinanceLife = twapBucket{lower: 600, upper: math.MaxFloat64, baseline: 5, cap: 10}

func pickGainBucketBinanceLife(marketCap float64) gainBucket {
	for _, b := range gainBucketsBinanceLife {
		if marketCap >= b.lower && marketCap < b.upper {
			return b
		}
	}
	return gainFallbackBinanceLife
}

func pickTwapBucketBinanceLife(marketCap float64) twapBucket {
	for _, b := range twapBucketsBinanceLife {
		if marketCap >= b.lower && marketCap < b.upper {
			return b
		}
	}
	return twapFallbackBinanceLife
}

func expectedSplitGainBinanceLife(marketCapM, fearGreedIndex, btc1d, btc7d float64, isMeme bool) float64 {
	bucket := pickGainBucketBinanceLife(marketCapM)
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

func expectedTwapDurationBinanceLife(marketCapM, fearGreedIndex, btc1d, btc7d float64, isMeme bool) float64 {
	bucket := pickTwapBucketBinanceLife(marketCapM)
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

func expectedSplitGainAndTwapDurationBinanceLife(marketCapM, fearGreedIndex, btc1d, btc7d float64, isMeme bool) (float64, float64) {
	return expectedSplitGainBinanceLife(marketCapM, fearGreedIndex, btc1d, btc7d, isMeme),
		expectedTwapDurationBinanceLife(marketCapM, fearGreedIndex, btc1d, btc7d, isMeme)
}
