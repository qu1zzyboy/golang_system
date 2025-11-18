package toUpbitBybitSymbol

import (
	"math"
	"upbitBnServer/internal/cal/u64Cal"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/quant/exchanges/bybit/account/bybitAccountAvailable"
	"upbitBnServer/internal/quant/exchanges/bybit/order/byBitOrderAppManager"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitParam"
	"upbitBnServer/pkg/utils/time2str"
)

/*
限制1: maxNotional 单账户单品种最大开仓上限
限制2：qtyTotal*0.3

退出循环:
1、账户没钱(不一定准)
*/

func (s *Single) PlacePostOnlyOrder() {
	limitPrice := u64Cal.FromF64(s.priceMaxBuy, s.pScale.Uint8())
	s.posTotalNeed = toUpbitParam.QtyTotal / s.priceMaxBuy

	// maker抽奖金额为 min(单账户单品种最大开仓上限,需开参数价值)
	maxOpenQty := math.Min(toUpbitParam.QtyTotal*toUpbitParam.F03, s.maxNotional)
	limitNum := u64Cal.FromF64(maxOpenQty/s.priceMaxBuy, s.qScale.Uint8())

	safex.SafeGo("to_upbit_bybit_limit_maker", func() {
		var accountIndex uint8
		var count int
		defer func() {
			toUpBitDataStatic.DyLog.GetLog().Infof("账户[%d],下单[%d]次 [%d,%d] 10ms POST_ONLY 协程结束", accountIndex, count, limitPrice, limitNum)
		}()
		// 遍历所有账户下单
		for accountIndex = 0; accountIndex < uint8(toUpbitParam.AccountLen); accountIndex++ {
			accountMaxQty := bybitAccountAvailable.GetManager().GetAvailable(accountIndex)
			// 可开仓金额太少
			if 4.0*accountMaxQty <= toUpbitParam.Dec500 {
				continue
			}

			//每个账户每秒最多下10次
			for i := 0; i <= 10; i++ {
				select {
				case <-s.ctxStop.Done():
					toUpBitDataStatic.DyLog.GetLog().Infof("收到关闭信号,退出POST_ONLY抽奖协程")
					return
				default:
					{
						//完全成交或者本轮挂单成功
						if s.hasAllFilled.Load() {
							return
						}
						if err := byBitOrderAppManager.GetTradeManager().SendPlaceOrder(accountIndex, orderModel.MyPlaceOrderReq{
							SymbolName:    s.symbolName,
							ClientOrderId: time2str.GetNowTimeStampMicroSlice16(),
							Pvalue:        limitPrice,
							Qvalue:        limitNum,
							Pscale:        s.pScale,
							Qscale:        s.qScale,
							OrderMode:     execute.BUY_OPEN_LIMIT_MAKER,
							SymbolIndex:   s.symbolIndex,
							SymbolLen:     s.symbolLen,
							ReqFrom:       from_bybit,
							UsageFrom:     to_upbit_main,
						}); err != nil {
							toUpBitDataStatic.DyLog.GetLog().Errorf("每秒limit_maker订单失败: %v", err)
						}
						count++
					}
				}
			}
		}
	})
}
