package toUpbitBybitSymbol

import (
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"
)

func (s *Single) onTradeLite(data toUpbitListChan.TrigOrderInfo) {
	// 处理预挂单成交
	s.pre.OnPreFilled(s.symbolName, data.ClientOrderId, s.pScale, s.qScale)

	if toUpBitListDataAfter.LoadTrig() {
		if s.symbolIndex == toUpBitListDataAfter.TrigSymbolIndex {
			toUpBitDataStatic.SendToUpBitMsg("发送bybit二次确认失败", map[string]string{
				"symbol": s.symbolName,
				"op":     "bybit_二次上涨确认",
			})
			toUpBitDataStatic.DyLog.GetLog().Infof("触发后二次确认:%s", s.symbolName)
		}
	} else {
		/*********************上币还未触发**************************/
		go func() {
			toUpBitDataStatic.SendToUpBitMsg("发送bybit预挂单成交失败", map[string]string{
				"symbol": s.symbolName,
				"op":     "bybit_预挂单成交",
			})
		}()
		s.onOrderPriceCheck(data.T, data.P)
	}
}
