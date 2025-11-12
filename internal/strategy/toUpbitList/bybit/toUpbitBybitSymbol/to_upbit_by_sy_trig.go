package toUpbitBybitSymbol

import (
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitBnMode"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
)

//to do

func (s *Single) onOrderPriceCheck(tradeTs int64, priceU64_8 uint64) {
	toUpBitDataStatic.SendToUpBitMsg("发送bn快速上涨消息失败", map[string]string{
		"msg":  "预挂单成交",
		"bn品种": s.symbolName,
	})
	s.IntoExecuteNoCheck(tradeTs, priceU64_8)
}

func (s *Single) IntoExecuteNoCheck(eventTs int64, priceTrig_8 uint64) {
	s.hasTreeNews = toUpbitBnMode.Mode.GetTreeNewsFlag()
	toUpBitListDataAfter.Trig(s.symbolIndex)
	s.startTrig()
	s.checkTreeNews()
	s.PlacePostOnlyOrder()
	s.tryBuyLoop(20)
	toUpBitDataStatic.DyLog.GetLog().Infof("[%s]探针成交,最新价格: %d,事件时间:%d", s.symbolName, priceTrig_8, eventTs)
}

func (s *Single) intoExecuteByMsg() {
	s.hasTreeNews = true
	toUpBitListDataAfter.Trig(s.symbolIndex)
	s.startTrig()
	s.tryBuyLoop(20)
	toUpBitDataStatic.DyLog.GetLog().Infof("treeNewsSub->[%s]触发", s.symbolName)
}

func (s *Single) checkMarket(eventTs int64, priceU8 uint64) {
	// 更新两分钟之前的价格
	minuteId := eventTs / (60000)
	if minuteId > s.thisMinTs {
		s.thisMinTs = minuteId
		s.last2MinClose_8 = s.last1MinClose_8
		s.last1MinClose_8 = s.thisMinClose_8
	} else {
		s.thisMinClose_8 = priceU8
	}
}
