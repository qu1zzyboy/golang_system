package toUpbitListBnSymbol

import (
	"fmt"

	"upbitBnServer/internal/quant/exchanges/binance/bnConst"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitBnMode"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"

	"github.com/shopspring/decimal"
)

//to do

func (s *Single) onOrderPriceCheck(tradeTs int64, priceU64_8 uint64) {
	// minBId>=0.95*markPrice
	if float64(s.minPriceAfterMp) >= toUpBitDataStatic.PriceRiceTrig*float64(s.markPrice_8) {
		s.IntoExecuteNoCheck(tradeTs, "preOrder", priceU64_8)
	} else {
		toUpBitDataStatic.SendToUpBitMsg("成交但不满足上市check消息失败", map[string]string{
			"msg":  fmt.Sprintf("成交但不满足上市check,成交价:%d", priceU64_8),
			"bn品种": s.StMeta.SymbolName,
			"上涨幅度": fmt.Sprintf("%.2f%%", s.lastRiseValue*100),
		})
	}
}

func (s *Single) IntoExecuteNoCheck(eventTs int64, trigFlag string, priceTrig_8 uint64) {
	s.hasTreeNews.Store(toUpbitBnMode.Mode.GetTreeNewsFlag())
	toUpBitListDataAfter.Trig(s.SymbolIndex)
	s.startTrig()
	limit := decimal.New(int64(s.priceMaxBuy_10), -bnConst.PScale_10).Truncate(s.pScale)
	s.checkTreeNews()
	s.placePostOnlyOrder(limit)
	s.tryBuyLoopBeforeNews()
	toUpBitDataStatic.DyLog.GetLog().Infof("%s->[%s]价格触发,最新价格: %d,涨幅: %f%%,事件时间:%d", trigFlag, s.StMeta.SymbolName, priceTrig_8, s.lastRiseValue*100, eventTs)
}
