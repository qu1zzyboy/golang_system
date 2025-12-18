package toUpbitListBnSymbol

import (
	"time"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/strategy/newsDrive/driverDefine"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitBnMode"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
)

func (s *Single) onBookTickExecute(f64 float64, ts int64) {
	s.bidPrice.Store(f64)

	// 已经触发过停止信号
	if s.hasReceiveStop {
		return
	}

	// 还不允许移动止损
	if !s.isStopLossAble.Load() {
		return
	}

	switch s.TrigExType {
	case exchangeEnum.UPBIT:
		// 止损判定
		tsSecond := ts / 1000
		// 只在最后100ms判断移动止损
		if ts >= tsSecond*1000+900 {
			if markPrice_u10, ok := s.trigPriceMax_10.Load(tsSecond); ok {
				if toUpbitBnMode.Mode.IsDynamicStopLossTrig(f64, float64(markPrice_u10)/1e10) {
					toUpBitDataStatic.DyLog.GetLog().Infof("移动止损触发,价格上限:%d,bid: %.8f", markPrice_u10, f64)
					s.receiveStop(driverDefine.StopByMoveStopLoss)
					return
				}
			}
		}

	case exchangeEnum.BINANCE:
		//价格涨到位,触发平仓
		if s.bnBeginTwapBuy > 0 && f64 > s.bnBeginTwapBuy {
			toUpBitDataStatic.DyLog.GetLog().Infof("触发平仓价格: %.8f,当前价格: %.8f", s.bnBeginTwapBuy, f64)
			if !s.bnAlreadyTwapBuy {
				s.bnAlreadyTwapBuy = true
				s.bnSpotCancel()
				safex.SafeGo("bnSpot_initBuyOpen", func() {
					s.initBnSpotBuyOpen()
					s.buyLoop()
				})
				safex.SafeGo("bnSpot_timeLine", func() {
					time.Sleep(25 * time.Second)
					s.receiveStop(driverDefine.StopByTimeLine)
				})
			}
			return
		}
		if s.stopLossPrice > 0 && f64 < s.stopLossPrice {
			toUpBitDataStatic.DyLog.GetLog().Infof("触发止损价格: %.8f,当前价格: %.8f", s.stopLossPrice, f64)
			s.receiveStop(driverDefine.StopByStopLoss)
			return
		}
	case exchangeEnum.UNKNOWN:
		return
	}

	// 通用止盈判定
	if s.takeProfitPrice > 0 && f64 > s.takeProfitPrice {
		toUpBitDataStatic.DyLog.GetLog().Infof("触发平仓价格: %.8f,当前价格: %.8f", s.takeProfitPrice, f64)
		s.receiveStop(driverDefine.StopByTakeProfit)
		return
	}
}
