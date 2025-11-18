package toUpbitBybitSymbol

import (
	"time"
	"upbitBnServer/internal/cal/u64Cal"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/infra/systemx/usageEnum"
	"upbitBnServer/internal/quant/exchanges/bybit/order/byBitOrderAppManager"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitDefine"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitParam"
	"upbitBnServer/pkg/utils/time2str"
)

const (
	to_upbit_main = usageEnum.TO_UPBIT_MAIN
	from_bybit    = instanceEnum.TO_UPBIT_LIST_BYBIT
)

func (s *Single) clear() {
	s.posTotalNeed = 0
	//清空持仓统计
	if s.posLong != nil {
		s.posLong.Clear()
	}
	s.takeProfitPrice = 0
	s.hasAllFilled.Store(false)
}

func (s *Single) getPosLong() float64 {
	if s.posLong == nil {
		return 0.0
	}
	return s.posLong.GetTotal()
}

func (s *Single) receiveStop(stopType toUpbitDefine.StopType) {
	if s.hasReceiveStop {
		return
	}
	s.hasReceiveStop = true
	toUpBitDataStatic.DyLog.GetLog().Infof("收到停止信号==> %s", toUpbitDefine.StopReasonArr[stopType])
	s.cancel()
	//开启平仓线程
	safex.SafeGo("to_upbit_bn_close", func() {
		defer func() {
			toUpBitDataStatic.DyLog.GetLog().Info(" 平仓协程结束")
			time.Sleep(2 * time.Second)
			s.clear()
			toUpBitListDataAfter.ClearTrig()
		}()
		// 撤销全部订单
		s.clientOrderIds.Range(func(clientOrderId systemx.WsId16B, accountKeyId uint8) bool {
			byBitOrderAppManager.GetTradeManager().SendCancelOrder(accountKeyId, orderModel.MyQueryOrderReq{
				SymbolName:    s.symbolName,
				ClientOrderId: clientOrderId,
				ReqFrom:       from_bybit,
				UsageFrom:     to_upbit_main,
			})
			return true
		})

		// 判断有没有持仓
		use := s.getPosLong()
		if use <= 0 {
			toUpBitDataStatic.DyLog.GetLog().Infof("没有可用的平仓数量,取消平仓")
			return
		}
		if use*s.priceMaxBuy <= toUpbitParam.Dec500 {
			toUpBitDataStatic.DyLog.GetLog().Infof("没有足够的平仓数量,取消平仓")
			return
		}
		//每秒平一次
		var closeDecArr [toUpbitParam.MaxAccount]float64 // 每个账户每秒应该止盈的数量
		perDec := 1 / s.twapSec
		copyMap := s.posLong.GetAllAccountPos()
		for accountKeyId, vol := range copyMap {
			closeDecArr[accountKeyId] = perDec * vol
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
					bid64 := val.(float64)
					posLeft := s.posLong.GetTotal()
					if posLeft*bid64 <= toUpbitParam.Dec500 {
						toUpBitDataStatic.DyLog.GetLog().Infof("平仓完全成交,开始清理资源")
						ticker.Stop()
						return
					}
					toUpBitDataStatic.DyLog.GetLog().Infof("============开始平仓,剩余:%.8f============", posLeft)
					// 最新的每个账户的仓位情况
					for accountKeyId, vol := range s.posLong.GetAllAccountPos() {
						// 已经完全平完了
						if vol <= 0 {
							continue
						}
						// 不够就全平
						num := closeDecArr[accountKeyId]
						if vol < num {
							num = vol
						}
						// 发送平仓信号
						if err := byBitOrderAppManager.GetTradeManager().SendPlaceOrder(accountKeyId, orderModel.MyPlaceOrderReq{
							SymbolName:    s.symbolName,
							ClientOrderId: time2str.GetNowTimeStampMicroSlice16(),
							Pvalue:        u64Cal.FromF64(bid64, s.pScale.Uint8()),
							Qvalue:        u64Cal.FromF64(num, s.qScale.Uint8()),
							Pscale:        s.pScale,
							Qscale:        s.qScale,
							OrderMode:     execute.SELL_CLOSE_LIMIT,
							SymbolIndex:   s.symbolIndex,
							SymbolLen:     s.symbolLen,
							ReqFrom:       from_bybit,
							UsageFrom:     to_upbit_main,
						}); err != nil {
							toUpBitDataStatic.DyLog.GetLog().Errorf("每秒平仓创建订单失败: %v", err)
						}
					}
				}
			case <-timeout:
				toUpBitDataStatic.DyLog.GetLog().Infof("平仓时间结束,开始清理资源")
				ticker.Stop()
				return
			}
		}
	})
}
