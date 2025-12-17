package bnDrive_upbitKrw

import (
	"lowLatencyServer/internal/strategy/newsDrive/common/driverDefine"
	"lowLatencyServer/internal/strategy/newsDrive/common/driverStatic"
)

func (s *Single) OnBid(bid float64) {
	// 已经触发过停止信号
	if s.hasReceiveStop {
		return
	}
	//价格涨到位,触发平仓
	if s.takeProfitPrice > 0 && bid > s.takeProfitPrice {
		driverStatic.DyLog.GetLog().Infof("触发平仓价格: %.8f,当前价格: %.8f", s.takeProfitPrice, bid)
		s.receiveStop(driverDefine.StopByBtTakeProfit)
		return
	}
}
