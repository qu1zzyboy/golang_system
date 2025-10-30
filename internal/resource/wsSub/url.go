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
