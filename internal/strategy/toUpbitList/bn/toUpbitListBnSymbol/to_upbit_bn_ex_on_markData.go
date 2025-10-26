package toUpbitListBnSymbol

import (
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitBnMode"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
)

func (s *Single) onBookTickExecute(f64 float64, ts int64) {
	s.bidPrice.Store(f64)

	// 已经触发过停止信号
	if s.hasReceiveStop {
		return
	}
	//价格涨到位,触发平仓
	if s.hasTreeNews && s.takeProfitPrice > 0 && f64 > s.takeProfitPrice {
		toUpBitListDataStatic.DyLog.GetLog().Infof("触发平仓价格: %.8f,当前价格: %.8f", s.takeProfitPrice, f64)
		s.receiveStop(StopByBtTakeProfit)
		return
	}

	// 还不允许移动止损
	if !s.isStopLossAble.Load() {
		return
	}
	// 止损判定
	tsSecond := ts / 1000
	// 只在最后100ms判断移动止损
	if ts >= tsSecond*1000+900 {
		if markPrice_u10, ok := s.trigPriceMax_10.Load(tsSecond); ok {
			if toUpbitBnMode.Mode.IsDynamicStopLossTrig(f64, float64(markPrice_u10)/1e10) {
				toUpBitListDataStatic.DyLog.GetLog().Infof("移动止损触发,价格上限:%d,bid: %.8f", markPrice_u10, f64)
				s.receiveStop(StopByMoveStopLoss)
				return
			}
		}
	}
}
