package bnDriveSymbol

import (
	"upbitBnServer/internal/strategy/newsDrive/common/driverStatic"
	"upbitBnServer/internal/strategy/newsDrive/upbit/toUpbitParam"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListPos"

	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"upbitBnServer/internal/quant/execute/order/orderModel"

	"github.com/shopspring/decimal"
)

func (s *Single) PlacePostOnlyOrder(limit decimal.Decimal) {
	s.posTotalNeed = toUpbitParam.QtyTotal.Div(limit).Truncate(s.qScale)

	// maker抽奖金额为 min(单账户单品种最大开仓上限,需开参数价值)
	orderNum := decimal.Min(s.maxNotional, toUpbitParam.QtyTotal.Mul(toUpbitParam.Dec03)).Div(limit).Truncate(s.qScale)

	safex.SafeGo("to_upbit_bn_limit_maker", func() {
		s.secondArr[0].start()
		var i int
		defer func() {
			driverStatic.DyLog.GetLog().Infof("账户[%d],下单[%d,%s]次 10ms POST_ONLY 协程结束", 0, i+1, limit.String())
		}()
	OUTER:
		for i = 0; i <= 200; i++ {
			select {
			case <-s.ctxStop.Done():
				driverStatic.DyLog.GetLog().Infof("收到关闭信号,退出POST_ONLY抽奖协程")
				break OUTER
			default:
				{
					//完全成交或者本轮挂单成功
					if s.secondArr[0].loadStop() || s.hasAllFilled.Load() {
						break OUTER
					}
					if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(order_from, 0, s.symbolIndex,
						&orderModel.MyPlaceOrderReq{
							OrigPrice:     limit,
							OrigVol:       orderNum,
							ClientOrderId: driverStatic.GetMakerClientOrderId(),
							StaticMeta:    s.StMeta,
							OrderType:     execute.ORDER_TYPE_POST_ONLY,
							OrderMode:     execute.ORDER_BUY_OPEN,
						}); err != nil {
						driverStatic.DyLog.GetLog().Errorf("每秒limit_maker订单失败: %v", err)
					}
					// time.Sleep(40 * time.Microsecond) // 休眠 40 微秒
				}
			}
		}
	})
	s.pos = toUpbitListPos.NewPosCal()
	s.firstPriceBuy = limit.Mul(toUpbitParam.Dec103).Truncate(s.pScale)
}
