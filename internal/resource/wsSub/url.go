package wsSub

import (
	"fmt"
	"strings"
	"upbitBnServer/internal/quant/market/kline/klineEnum"
)

func getBookTickFuSubParam(symbol string) string {
	return fmt.Sprintf("%s@bookTicker", strings.ToLower(symbol))
}

func getAggTradeFuSubParam(symbol string) string {
	return fmt.Sprintf("%s@aggTrade", strings.ToLower(symbol))
}

func getMarkPriceFuSubParam(symbol string) string {
	return fmt.Sprintf("%s@markPrice@1s", strings.ToLower(symbol))
}

func getKlineFuSubParam(interval klineEnum.Interval) func(symbol string) string {
	return func(symbol string) string {
		return fmt.Sprintf("%s@kline_%s", strings.ToLower(symbol), interval.String())
	}
}

func getContinuousKlineFuSubParam(interval klineEnum.Interval) func(symbol string) string {
	return func(symbol string) string {
		// 币安连续 K 线格式: <pair>_<contractType>@continuousKline_<interval>
		// 合约类型通常为 perpetual（永续合约，小写）
		return fmt.Sprintf("%s_perpetual@continuousKline_%s", strings.ToLower(symbol), interval.String())
	}
}
