package toUpbitParam

import (
	"math"

	exchangeEnum "upbitBnServer/internal/quant/exchanges/exchangeEnum"
)

type gainBucket struct {
	lower    float64
	upper    float64
	baseline float64
	cap      float64
}

type twapBucket struct {
	lower    float64
	upper    float64
	baseline float64
	cap      float64
}

var gainBuckets = []gainBucket{
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

var gainFallback = gainBucket{lower: 600, upper: math.MaxFloat64, baseline: 0.1, cap: 0.2}

var twapBuckets = []twapBucket{
	{lower: 0, upper: 5, baseline: 60, cap: 90},
	{lower: 5, upper: 100, baseline: 45, cap: 65},
	{lower: 100, upper: 120, baseline: 35, cap: 52},
	{lower: 120, upper: 200, baseline: 45, cap: 67},
	{lower: 200, upper: 330, baseline: 45, cap: 67},
	{lower: 330, upper: 600, baseline: 10, cap: 15},
}

var twapFallback = twapBucket{lower: 600, upper: math.MaxFloat64, baseline: 5, cap: 10}

const (
	fearGreedHighAdd    = 3.0
	fearGreedLowSub     = 2.0
	btc1dMultiplier     = 1.0
	btc7dMultiplier     = 0.2
	fearGreedHighSubSec = 5.0
	fearGreedLowSubSec  = 5.0
	btc1dSecondsPerPct  = 0.6
	btc7dSecondsPerPct  = 0.3
	memeExtraSeconds    = 5.0
	memecoinGainAdd     = 0.0
)

func pickGainBucket(marketCap float64) gainBucket {
	for _, b := range gainBuckets {
		if marketCap >= b.lower && marketCap < b.upper {
			return b
		}
	}
	return gainFallback
}

func pickTwapBucket(marketCap float64) twapBucket {
	for _, b := range twapBuckets {
		if marketCap >= b.lower && marketCap < b.upper {
			return b
		}
	}
	return twapFallback
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

func expectedSplitGainAndTwapDuration(marketCapM, fearGreedIndex, btc1d, btc7d float64, isMeme bool) (float64, float64) {
	return expectedSplitGain(marketCapM, fearGreedIndex, btc1d, btc7d, isMeme),
		expectedTwapDuration(marketCapM, fearGreedIndex, btc1d, btc7d, isMeme)
}

func ExpectedSplitGainAndTwapDurationWithExchange(exchange exchangeEnum.ExchangeType, marketCapM, fearGreedIndex, btc1d, btc7d float64, isMeme bool) (float64, float64) {
	switch normalizeExchange(exchange) {
	case exchangeEnum.BINANCE:
		return expectedSplitGainAndTwapDurationBinance(marketCapM, fearGreedIndex, btc1d, btc7d, isMeme)
	default:
		return expectedSplitGainAndTwapDuration(marketCapM, fearGreedIndex, btc1d, btc7d, isMeme)
	}
}

func clipGain(exchange exchangeEnum.ExchangeType, marketCapM, target float64) (float64, float64, float64) {
	bucket := pickGainBucketByExchange(exchange, marketCapM)
	return clampFloat(target, bucket.baseline, bucket.cap), bucket.baseline, bucket.cap
}

func clipTwap(exchange exchangeEnum.ExchangeType, marketCapM, target float64) (float64, float64, float64) {
	bucket := pickTwapBucketByExchange(exchange, marketCapM)
	return clampFloat(target, bucket.baseline, bucket.cap), bucket.baseline, bucket.cap
}

func pickGainBucketByExchange(exchange exchangeEnum.ExchangeType, marketCap float64) gainBucket {
	switch normalizeExchange(exchange) {
	case exchangeEnum.BINANCE:
		return pickGainBucketBinance(marketCap)
	default:
		return pickGainBucket(marketCap)
	}
}

func pickTwapBucketByExchange(exchange exchangeEnum.ExchangeType, marketCap float64) twapBucket {
	switch normalizeExchange(exchange) {
	case exchangeEnum.BINANCE:
		return pickTwapBucketBinance(marketCap)
	default:
		return pickTwapBucket(marketCap)
	}
}

func normalizeExchange(ex exchangeEnum.ExchangeType) exchangeEnum.ExchangeType {
	switch ex {
	case exchangeEnum.BINANCE:
		return exchangeEnum.BINANCE
	case exchangeEnum.UPBIT:
		return exchangeEnum.UPBIT
	default:
		return exchangeEnum.UPBIT
	}
}

const (
	sLow          = 0.03
	sHigh         = 0.1
	gainOiMax     = 5.0
	twapOiMaxSecs = 10.0
)

func computeOIContribs(oiNotional float64, marketCapM float64) (gainAdd, twapAdd float64, strength, norm float64) {
	if oiNotional <= 0 {
		return 0, 0, 0, 0
	}
	marketCapUSD := marketCapM * 1e6
	if marketCapUSD <= 0 {
		return 0, 0, 0, 0
	}
	s := oiNotional / marketCapUSD
	n := normalizeS(s)
	gain := gainOiMax * n
	twap := twapOiMaxSecs * n
	return gain, twap, s, n
}

func normalizeS(s float64) float64 {
	switch {
	case s <= sLow:
		return 1.0
	case s >= sHigh:
		return -1.0
	default:
		return 1.0 - 2.0*(s-sLow)/(sHigh-sLow)
	}
}

func clampFloat(value, lower, upper float64) float64 {
	if value < lower {
		return lower
	}
	if value > upper {
		return upper
	}
	return value
}
