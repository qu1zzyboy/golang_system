package klineEnum

const (
	p_KLINE_1s  = "1s"
	p_KLINE_1m  = "1m"
	p_KLINE_3m  = "3m"
	p_KLINE_5m  = "5m"
	p_KLINE_15m = "15m"
	p_KLINE_30m = "30m"
	p_KLINE_1h  = "1h"
	p_KLINE_2h  = "2h"
	p_KLINE_4h  = "4h"
	p_KLINE_1d  = "1d"
	p_KLINE_1w  = "1w"
)

type Interval uint8

func (s Interval) String() string {
	switch s {
	case KLINE_1s:
		return p_KLINE_1s
	case KLINE_1m:
		return p_KLINE_1m
	case KLINE_3m:
		return p_KLINE_3m
	case KLINE_5m:
		return p_KLINE_5m
	case KLINE_15m:
		return p_KLINE_15m
	case KLINE_30m:
		return p_KLINE_30m
	case KLINE_1h:
		return p_KLINE_1h
	case KLINE_2h:
		return p_KLINE_2h
	case KLINE_4h:
		return p_KLINE_4h
	case KLINE_1d:
		return p_KLINE_1d
	case KLINE_1w:
		return p_KLINE_1w
	default:
		return "UNSET"
	}
}

const (
	KLINE_1s Interval = iota
	KLINE_1m
	KLINE_3m
	KLINE_5m
	KLINE_15m
	KLINE_30m
	KLINE_1h
	KLINE_2h
	KLINE_4h
	KLINE_1d
	KLINE_1w
)

// GetIntervalTimeStep 获取时间间隔的毫秒数
func GetIntervalTimeStep(interval Interval) int64 {
	switch interval {
	case KLINE_1s:
		return 1000
	case KLINE_1m:
		return 60 * 1000
	case KLINE_3m:
		return 180 * 1000
	case KLINE_5m:
		return 300 * 1000
	case KLINE_15m:
		return 900 * 1000
	case KLINE_30m:
		return 1800 * 1000
	case KLINE_1h:
		return 3600 * 1000
	case KLINE_2h:
		return 7200 * 1000
	case KLINE_4h:
		return 14400 * 1000
	case KLINE_1d:
		return 86400 * 1000
	case KLINE_1w:
		return 604800 * 1000
	default:
		return -1
	}
}
