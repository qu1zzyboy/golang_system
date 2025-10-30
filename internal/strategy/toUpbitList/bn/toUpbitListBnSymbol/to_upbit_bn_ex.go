package toUpbitListBnSymbol

import (
	"time"

	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitDefine"

	"github.com/shopspring/decimal"
)

const (
	order_from = orderBelongEnum.TO_UPBIT_LIST_LOOP
)

func (s *Single) clear() {
	s.posTotalNeed = decimal.Zero
	s.pos.Clear() //清空持仓统计
	s.takeProfitPrice = 0
	for i := range s.secondArr {
		s.secondArr[i].clear()
	}
	s.hasAllFilled.Store(false)
	s.thisOrderAccountId.Store(0)
	toUpBitListDataAfter.ClearTrig()
}

func (s *Single) receiveStop(stopType toUpbitDefine.StopType) {
	if s.hasReceiveStop {
		return
	}
	s.hasReceiveStop = true
	toUpBitListDataStatic.DyLog.GetLog().Infof("收到停止信号==> %s", toUpbitDefine.StopReasonArr[stopType])
	s.cancel()
	//开启平仓线程
	safex.SafeGo("to_upbit_bn_close", func() {
		defer func() {
			toUpBitListDataStatic.DyLog.GetLog().Infof("当前账户id[%d] 平仓协程结束", s.thisOrderAccountId.Load())
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
		use := s.pos.GetTotalVol()
		if use.LessThanOrEqual(decimal.Zero) {
			toUpBitListDataStatic.DyLog.GetLog().Infof("没有可用的平仓数量,取消平仓")
			return
		}
		if use.LessThanOrEqual(toUpBitListDataStatic.Dec500) {
			toUpBitListDataStatic.DyLog.GetLog().Infof("没有足够的平仓数量,取消平仓")
			return
		}
		//每秒平一次
		var closeDecArr [11]decimal.Decimal // 每个账户每秒应该止盈的数量
		perDec := decimal.NewFromFloat(1 / s.twapSec)
		copyMap := s.pos.GetAllAccountPos()
		for accountKeyId, vol := range copyMap {
			closeDecArr[accountKeyId] = vol.Mul(perDec).Truncate(s.qScale) //每秒应该止盈的数量
		}
		ticker := time.NewTicker(time.Second)
		timeout := time.After(s.closeDuration)
		for {
			select {
			case <-ticker.C:
				{
					val := s.bidPrice.Load()
					if val == nil {
						continue
					}
					priceDec := decimal.NewFromFloat(val.(float64)).Truncate(s.pScale)
					posLeft := s.pos.GetTotalVol()
					if s.pos.GetTotalVol().Mul(priceDec).LessThanOrEqual(toUpBitListDataStatic.Dec500) {
						toUpBitListDataStatic.DyLog.GetLog().Infof("平仓完全成交,开始清理资源")
						ticker.Stop()
						return
					}
					toUpBitListDataStatic.DyLog.GetLog().Infof("============开始平仓,剩余:%s============", posLeft)
					// 最新的每个账户的仓位情况
					copyMap := s.pos.GetAllAccountPos()
					for accountKeyId, vol := range copyMap {
						// 已经完全平完了
						if vol.LessThanOrEqual(decimal.Zero) {
							continue
						}
						// 不够就全平
						num := closeDecArr[accountKeyId]
						if vol.LessThan(num) {
							num = vol.Truncate(s.qScale)
						}
						// 发送平仓信号
						if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(order_from, accountKeyId, s.symbolIndex,
							&orderModel.MyPlaceOrderReq{
								OrigPrice:     priceDec,
								OrigVol:       num,
								ClientOrderId: toUpBitListDataStatic.GetClientOrderIdBy("close"),
								StaticMeta:    s.StMeta,
								OrderType:     execute.ORDER_TYPE_LIMIT,
								OrderMode:     execute.ORDER_SELL_CLOSE,
							}); err != nil {
							toUpBitListDataStatic.DyLog.GetLog().Errorf("每秒平仓创建订单失败: %v", err)
						}
					}
				}
			case <-timeout:
				toUpBitListDataStatic.DyLog.GetLog().Infof("平仓时间结束,开始清理资源")
				ticker.Stop()
				return
			}
		}
	})
}
