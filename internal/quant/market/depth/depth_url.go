package depth

import (
	"fmt"
	"strings"
)

var (
	BN    bn
	BYBIT bybit
)

type bn struct {
}

type bybit struct {
}

// GetFuSubParam 获取币安期货订单簿深度订阅参数
// levels: 深度档位（如20表示20档）
// updateSpeed: 更新速度（毫秒，如500表示500ms，100表示100ms）
func (s bn) GetFuSubParam(levels int, updateSpeed int) func(symbol string) string {
	return func(symbol string) string {
		// 币安订单簿深度格式: <symbol>@depth<levels>@<updateSpeed>ms
		// 例如: btcusdt@depth20@500ms 表示20档深度，500ms更新
		return fmt.Sprintf("%s@depth%d@%dms", strings.ToLower(symbol), levels, updateSpeed)
	}
}

func (s bybit) GetFuSubParam(symbol string) string {
	// Bybit 订单簿深度格式（待实现）
	return fmt.Sprintf("orderbook.1.%s", strings.ToUpper(symbol))
}
