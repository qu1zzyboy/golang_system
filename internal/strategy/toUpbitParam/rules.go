package toUpbitParam

import (
	"math"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
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

func ExpectedSplitGainAndTwapDurationWithExchange(exType exchangeEnum.ExchangeType, marketCapM, fearGreedIndex, btc1d, btc7d float64, isMeme bool) (float64, float64) {
	switch exType {
	case exchangeEnum.BINANCE:
		return expectedSplitGainAndTwapDurationBinance(marketCapM, fearGreedIndex, btc1d, btc7d, isMeme)
	default:
		return expectedSplitGainAndTwapDuration(marketCapM, fearGreedIndex, btc1d, btc7d, isMeme)
	}
}

func clipGain(exType exchangeEnum.ExchangeType, marketCapM, target float64) (float64, float64, float64) {
	bucket := pickGainBucketByExchange(exType, marketCapM)
	return clampFloat(target, bucket.baseline, bucket.cap), bucket.baseline, bucket.cap
}

func pickGainBucketByExchange(exType exchangeEnum.ExchangeType, marketCap float64) gainBucket {
	switch exType {
	case exchangeEnum.UPBIT:
		return pickGainBucketUpbitKrw(marketCap)
	case exchangeEnum.BINANCE:
		return pickGainBucketBinance(marketCap)
	default:
		return pickGainBucketUpbitKrw(marketCap)
	}
}

func clipTwap(exType exchangeEnum.ExchangeType, marketCapM, target float64) (float64, float64, float64) {
	bucket := pickTwapBucketByExchange(exType, marketCapM)
	return clampFloat(target, bucket.baseline, bucket.cap), bucket.baseline, bucket.cap
}

func pickTwapBucketByExchange(exType exchangeEnum.ExchangeType, marketCap float64) twapBucket {
	switch exType {
	case exchangeEnum.UPBIT:
		return pickTwapBucketUpbitKrw(marketCap)
	case exchangeEnum.BINANCE:
		return pickTwapBucketBinance(marketCap)
	default:
		return pickTwapBucketUpbitKrw(marketCap)
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
