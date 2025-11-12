package toUpbitListBnSymbol

import (
	"math"
	"upbitBnServer/internal/cal/u64Cal"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitParam"
	"upbitBnServer/pkg/utils/time2str"

	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
)

/*
限制1: maxNotional 单账户单品种最大开仓上限
限制2：qtyTotal*0.3

退出循环:
1、账户没钱(不一定准)
*/

func (s *Single) PlacePostOnlyOrder() {
	limit_p := u64Cal.FromF64(s.priceMaxBuy, s.pScale.Uint8())
	s.posTotalNeed = toUpbitParam.QtyTotal / s.priceMaxBuy
	maxOpenQty := math.Min(toUpbitParam.QtyTotal*toUpbitParam.F03, s.maxNotional)
	// maker抽奖金额为 min(单账户单品种最大开仓上限,需开参数价值)
	orderNum := u64Cal.FromF64(maxOpenQty/s.priceMaxBuy, s.qScale.Uint8())

	safex.SafeGo("to_upbit_bn_limit_maker", func() {
		s.secondArr[0].start()
		var i int
		defer func() {
			toUpBitDataStatic.DyLog.GetLog().Infof("账户[%d],下单[%d]次 10ms POST_ONLY 协程结束", 0, i+1)
		}()
	OUTER:
		for i = 0; i <= 200; i++ {
			select {
			case <-s.ctxStop.Done():
				toUpBitDataStatic.DyLog.GetLog().Infof("收到关闭信号,退出POST_ONLY抽奖协程")
				break OUTER
			default:
				{
					//完全成交或者本轮挂单成功
					if s.secondArr[0].loadStop() || s.hasAllFilled.Load() {
						break OUTER
					}
					if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(0, orderModel.MyPlaceOrderReq{
						SymbolName:    s.symbolName,
						ClientOrderId: time2str.GetNowTimeStampMicroSlice16(),
						Pvalue:        limit_p,
						Qvalue:        orderNum,
						Pscale:        s.pScale,
						Qscale:        s.qScale,
						OrderMode:     execute.BUY_OPEN_LIMIT_MAKER,
						SymbolIndex:   s.symbolIndex,
						SymbolLen:     s.symbolLen,
						ReqFrom:       instanceEnum.TO_UPBIT_LIST_BN,
						UsageFrom:     to_upbit_main,
					}); err != nil {
						toUpBitDataStatic.DyLog.GetLog().Errorf("每秒limit_maker订单失败: %v", err)
					}
					// time.Sleep(40 * time.Microsecond) // 休眠 40 微秒
				}
			}
		}
	})
}
