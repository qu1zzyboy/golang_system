package toUpbitListBnSymbol

import (
	"time"

	"upbitBnServer/internal/cal/u64Cal"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/infra/systemx/usageEnum"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitDefine"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitParam"
	"upbitBnServer/pkg/utils/time2str"
)

const (
	to_upbit_main = usageEnum.TO_UPBIT_MAIN
)

func (s *Single) Clear() {
	s.posTotalNeed = 0
	//清空持仓统计
	if s.pos != nil {
		s.pos.Clear()
	}
	s.takeProfitPrice = 0
	for i := range s.secondArr {
		s.secondArr[i].clear()
	}
	s.hasAllFilled.Store(false)
	s.thisOrderAccountId.Store(0)
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
			toUpBitDataStatic.DyLog.GetLog().Infof("当前账户id[%d] 平仓协程结束", s.thisOrderAccountId.Load())
			time.Sleep(2 * time.Second)
			s.Clear()
			toUpBitListDataAfter.ClearTrig()
		}()
		// 撤销全部订单
		s.clientOrderIds.Range(func(clientOrderId systemx.WsId16B, accountKeyId uint8) bool {
			s.can.RefreshClientOrderId(clientOrderId)
			bnOrderAppManager.GetTradeManager().SendCancelOrderBy(s.can, instanceEnum.TO_UPBIT_LIST_BN, to_upbit_main, accountKeyId)
			return true
		})

		if s.pos == nil {
			toUpBitDataStatic.DyLog.GetLog().Infof("s.pos is nil,取消平仓")
			return
		}
		// 判断有没有持仓
		use := s.pos.GetTotal()
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
		copyMap := s.pos.GetAllAccountPos()
		for accountKeyId, vol := range copyMap {
			closeDecArr[accountKeyId] = vol * perDec //每秒应该止盈的数量
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
					posLeft := s.pos.GetTotal()
					if posLeft*bid64 <= toUpbitParam.Dec500 {
						toUpBitDataStatic.DyLog.GetLog().Infof("平仓完全成交,开始清理资源")
						ticker.Stop()
						return
					}
					toUpBitDataStatic.DyLog.GetLog().Infof("============开始平仓,剩余:%s============", posLeft)
					// 最新的每个账户的仓位情况
					copyMap := s.pos.GetAllAccountPos()
					for accountKeyId, vol := range copyMap {
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
						if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(accountKeyId, orderModel.MyPlaceOrderReq{
							SymbolName:    s.symbolName,
							ClientOrderId: time2str.GetNowTimeStampMicroSlice16(),
							Pvalue:        u64Cal.FromF64(bid64, s.pScale.Uint8()),
							Qvalue:        u64Cal.FromF64(num, s.qScale.Uint8()),
							Pscale:        s.pScale,
							Qscale:        s.qScale,
							OrderMode:     execute.SELL_CLOSE_LIMIT,
							SymbolIndex:   s.symbolIndex,
							SymbolLen:     s.symbolLen,
							ReqFrom:       instanceEnum.TO_UPBIT_LIST_BN,
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
