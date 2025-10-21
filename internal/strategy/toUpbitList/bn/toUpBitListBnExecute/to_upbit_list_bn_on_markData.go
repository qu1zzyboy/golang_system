package toUpBitListBnExecute

import (
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
)

func (s *Execute) OnBookTick(f64 float64, ts int64) {
	//价格涨到位,触发平仓
	if s.hasTreeNews.Load() && s.takeProfitPrice > 0 && f64 > s.takeProfitPrice {
		toUpBitListDataStatic.DyLog.GetLog().Infof("触发平仓价格: %.8f,当前价格: %.8f", s.takeProfitPrice, f64)
		s.ReceiveStop(StopByBtTakeProfit)
		return
	}
	// 止损判定
	tsSecond := ts / 1000
	// 只在最后100ms判断移动止损
	if ts >= tsSecond*1000+900 {
		if markPrice_u10, ok := toUpBitListDataAfter.TrigPriceMax_10.Load(tsSecond); ok {
			maxPriceF64 := float64(markPrice_u10) / 1e10
			if toUpBitListDataStatic.IsDebug {
				if f64 < maxPriceF64*0.86 {
					toUpBitListDataStatic.DyLog.GetLog().Infof("测试移动止损触发,价格上限:%d,bid: %.8f", markPrice_u10, f64)
					s.ReceiveStop(StopByMoveStopLoss)
					return
				}
			} else {
				if f64 < maxPriceF64*0.95 {
					toUpBitListDataStatic.DyLog.GetLog().Infof("实盘移动止损触发,价格上限:%d,bid: %.8f", markPrice_u10, f64)
					s.ReceiveStop(StopByMoveStopLoss)
					return
				}
			}
		}
	}
}
