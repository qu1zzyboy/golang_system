package toUpbitListBnSymbol

import (
	"time"
	"upbitBnServer/internal/strategy/newsDrive/driverDefine"
	"upbitBnServer/internal/strategy/twapLimitClose"

	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"

	"github.com/shopspring/decimal"
)

const (
	order_from = orderBelongEnum.TO_UPBIT_LIST_LOOP
)

func (s *Single) clear() {
	s.PosTotalNeed = decimal.Zero
	if s.Pos != nil {
		s.Pos.Clear() //清空持仓统计
	}
	s.takeProfitPrice = 0
	for i := range s.SecondArr {
		s.SecondArr[i].clear()
	}
	s.hasAllFilled.Store(false)
	s.thisOrderAccountId.Store(0)
	s.TrigExType = exchangeEnum.UNKNOWN
	s.bnSpotPerNum = decimal.Zero
	s.bnBeginTwapBuy = 0
	s.stopLossPrice = 0
	s.takeProfitPrice = 0
	s.hasTreeNews.Store(false)
	s.hasReceiveStop = false
	s.bnAlreadyTwapBuy = false
	toUpBitListDataAfter.ClearTrig()
}

func (s *Single) receiveStop(stopType driverDefine.StopType) {
	if s.hasReceiveStop {
		return
	}
	s.hasReceiveStop = true
	toUpBitDataStatic.DyLog.GetLog().Infof("收到停止信号==> %s", driverDefine.StopReasonArr[stopType])
	s.cancel()
	//开启平仓线程
	safex.SafeGo("to_upbit_bn_close", func() {
		defer func() {
			toUpBitDataStatic.DyLog.GetLog().Infof("当前账户id[%d] 平仓协程结束", s.thisOrderAccountId.Load())
			time.Sleep(20 * time.Millisecond)
			s.clear()
		}()
		// 撤销全部订单
		s.clientOrderIds.Range(func(clientOrderId string, accountKeyId uint8) bool {
			bnOrderAppManager.GetTradeManager().SendCancelOrder(order_from, accountKeyId, &orderModel.MyQueryOrderReq{
				ClientOrderId: clientOrderId,
				StaticMeta:    s.StMeta,
			})
			return true
		})

		// 判断有没有持仓
		use := s.Pos.GetTotal()
		if use.LessThanOrEqual(decimal.Zero) {
			toUpBitDataStatic.DyLog.GetLog().Infof("没有可用的平仓数量,取消平仓")
			return
		}
		if use.LessThanOrEqual(toUpBitDataStatic.Dec500) {
			toUpBitDataStatic.DyLog.GetLog().Infof("没有足够的平仓数量,取消平仓")
			return
		}
		val := s.bidPrice.Load()
		if val == nil {
			return
		}
		priceDec := decimal.NewFromFloat(val.(float64)).Truncate(s.pScale)
		//每秒平一次
		//var closeDecArr [11]decimal.Decimal // 每个账户每秒应该止盈的数量
		perDec := decimal.NewFromFloat(1 / s.twapSec)
		copyMap := s.Pos.GetAllAccountPos()
		for accountKeyId, vol := range copyMap {
			if vol.LessThanOrEqual(decimal.Zero) {
				continue
			}
			twapLimitClose.InitPerSecondBegin(accountKeyId, s.SymbolIndex, s.pScale, s.QScale, s.StMeta, s.closeMap[accountKeyId], priceDec, vol, vol.Mul(perDec).Truncate(s.QScale))
			//closeDecArr[accountKeyId] = vol.Mul(perDec).Truncate(s.QScale) //每秒应该止盈的数量
		}
		ticker := time.NewTicker(time.Second)
		timeout := time.After(s.closeDuration)
		for {
			select {
			case <-ticker.C:
				{
					val = s.bidPrice.Load()
					if val == nil {
						continue
					}
					priceDec = decimal.NewFromFloat(val.(float64)).Truncate(s.pScale)
					posLeft := s.Pos.GetTotal()
					if s.Pos.GetTotal().Mul(priceDec).LessThanOrEqual(toUpBitDataStatic.Dec500) {
						toUpBitDataStatic.DyLog.GetLog().Infof("平仓完全成交,开始清理资源")
						ticker.Stop()
						return
					}
					toUpBitDataStatic.DyLog.GetLog().Infof("============开始平仓,剩余:%s============", posLeft)
					for accountKeyId, closeOrderMap := range s.closeMap {
						twapLimitClose.RefreshPerSecondEnd(uint8(accountKeyId), s.StMeta, closeOrderMap, priceDec)
						twapLimitClose.RefreshPerSecondBegin(uint8(accountKeyId), s.pScale, s.StMeta, closeOrderMap, priceDec)
					}
					//// 最新的每个账户的仓位情况
					//copyMap := s.Pos.GetAllAccountPos()
					//for accountKeyId, vol := range copyMap {
					//	// 已经完全平完了
					//	if vol.LessThanOrEqual(decimal.Zero) {
					//		continue
					//	}
					//	// 不够就全平
					//	num := closeDecArr[accountKeyId]
					//	if vol.LessThan(num) {
					//		num = vol.Truncate(s.QScale)
					//	}
					//	// 发送平仓信号
					//	if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(order_from, accountKeyId, s.SymbolIndex,
					//		&orderModel.MyPlaceOrderReq{
					//			OrigPrice:     priceDec,
					//			OrigVol:       num,
					//			ClientOrderId: toUpBitDataStatic.GetClientOrderIdBy("server_close"),
					//			StaticMeta:    s.StMeta,
					//			OrderType:     execute.ORDER_TYPE_LIMIT,
					//			OrderMode:     execute.ORDER_SELL_CLOSE,
					//		}); err != nil {
					//		toUpBitDataStatic.DyLog.GetLog().Errorf("每秒平仓创建订单失败: %v", err)
					//	}
					//}
				}
			case <-timeout:
				toUpBitDataStatic.DyLog.GetLog().Infof("平仓时间结束,开始清理资源")
				ticker.Stop()
				return
			}
		}
	})
}
