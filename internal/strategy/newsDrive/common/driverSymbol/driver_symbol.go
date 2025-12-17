package driverSymbol

import (
	"sync/atomic"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/strategy/newsDrive/common/driverStatic"

	"github.com/tidwall/gjson"
)

type Symbol struct {
	BidPrice       atomic.Value           // 买一价,平仓和计算仓位价值用到
	SymbolName     string                 // 交易对名称
	TrigMarkPrice  float64                // 最新标记价格
	PriceMaxBuy    float64                // 价格上限
	PScale         systemx.PScale         // 价格小数位
	QScale         systemx.QScale         // 数量小数位
	upLimitPercent float32                // 涨停百分比
	HasTreeNews    atomic.Bool            // 是否已经接受到treeNews
	SymbolIndex    systemx.SymbolIndex16I // 交易对下标
}

func (s *Symbol) Clear() {
	s.HasTreeNews.Store(false)
}

func (s *Symbol) OnMarkPriceBefore(b []byte) float64 {
	thisMarkPrice := gjson.GetBytes(b, "p").Float()
	// 2、计算价格上限
	s.TrigMarkPrice = thisMarkPrice
	s.PriceMaxBuy = thisMarkPrice * float64(s.upLimitPercent)
	return thisMarkPrice
}

func (s *Symbol) OnMarkPriceAfter(b []byte) (float64, bool) {
	if s.SymbolIndex != driverStatic.TrigSymbolIndex {
		return 0, false
	}
	thisMarkPrice := gjson.GetBytes(b, "p").Float()
	priceMaxBuy := thisMarkPrice * float64(s.upLimitPercent)
	return priceMaxBuy, true
}

func (s *Symbol) OnBookTickAfter(byteLen uint16, b []byte) (float64, bool) {
	if s.SymbolIndex != driverStatic.TrigSymbolIndex {
		return 0, false
	}
	bid := gjson.GetBytes(b, "b").Float()
	return bid, true
}
