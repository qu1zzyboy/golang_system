package toUpbitListBnSymbol

import (
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitBnMode"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitDefine"
)

func (s *Single) onBookTickExecute(bid float64, ts int64) {
	s.bidPrice.Store(bid)

	// 已经触发过停止信号
	if s.hasReceiveStop {
		return
	}

	// 还不允许移动止损
	if !s.isStopLossAble.Load() {
		return
	}

	//价格涨到位,触发平仓
	if s.hasTreeNews && s.takeProfitPrice > 0 && bid > s.takeProfitPrice {
		toUpBitDataStatic.DyLog.GetLog().Infof("触发平仓价格: %.8f,当前价格: %.8f", s.takeProfitPrice, bid)
		s.receiveStop(toUpbitDefine.StopByBtTakeProfit)
		return
	}

	// 止损判定
	tsSecond := ts / 1000
	// 只在最后100ms判断移动止损
	if ts >= tsSecond*1000+900 {
		if buyMax, ok := s.trigPriceMax.Load(tsSecond); ok {
			if toUpbitBnMode.Mode.IsDynamicStopLossTrig(bid, buyMax) {
				toUpBitDataStatic.DyLog.GetLog().Infof("移动止损触发,价格上限:%.8f,bid: %.8f", buyMax, bid)
				s.receiveStop(toUpbitDefine.StopByMoveStopLoss)
				return
			}
		}
	}
}
