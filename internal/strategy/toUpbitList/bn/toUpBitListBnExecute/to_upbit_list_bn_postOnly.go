package toUpBitListBnExecute

import (
	"time"

	"github.com/hhh500/quantGoInfra/infra/safex"
	"github.com/hhh500/upbitBnServer/internal/quant/execute"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderModel"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"github.com/shopspring/decimal"
)

var (
	dec03    decimal.Decimal                     //首次下 0.3
	dec103   = decimal.RequireFromString("1.03") //第一次下单价格1.03倍
	qtyTotal decimal.Decimal                     //单次下单总金额
)

func (s *Execute) SetParam(qty, dec003 float64) {
	qtyTotal = decimal.NewFromFloat(qty)
	dec03 = decimal.NewFromFloat(dec003)
}

/*
限制1: maxNotional 单账户单品种最大开仓上限
限制2：qtyTotal*0.3

退出循环:
1、账户没钱(不一定准)
*/

func (s *Execute) PlacePostOnlyOrder(limit decimal.Decimal) {
	s.posTotalNeed = qtyTotal.Div(limit).Truncate(s.QScale)

	// maker抽奖金额为 min(单账户单品种最大开仓上限,需开参数价值)
	orderNum := decimal.Min(s.maxNotional, qtyTotal.Mul(dec03)).Div(limit).Truncate(s.QScale)

	safex.SafeGo("to_upbit_bn_limit_maker", func() {
		s.stopThisSecondPerArr[0].Store(false)   // 开启本轮抽奖信号
		s.hasInToSecondPerLoopArr[0].Store(true) // 确认进入了每秒抽奖循环
		var i int
		defer func() {
			toUpBitListDataStatic.DyLog.GetLog().Infof("账户[%d],下单[%d]次 10ms POST_ONLY 协程结束", 0, i+1)
		}()
	OUTER:
		for i = 0; i <= 200; i++ {
			select {
			case <-s.ctxStop.Done():
				toUpBitListDataStatic.DyLog.GetLog().Infof("收到关闭信号,退出POST_ONLY抽奖协程")
				break OUTER
			default:
				{
					//完全成交或者本轮挂单成功
					if s.stopThisSecondPerArr[0].Load() || s.hasAllFilled.Load() {
						break OUTER
					}
					if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(order_from, 0, s.symbolIndex,
						&orderModel.MyPlaceOrderReq{
							OrigPrice:     limit,
							OrigVol:       orderNum,
							ClientOrderId: toUpBitListDataStatic.GetMakerClientOrderId(),
							StaticMeta:    s.StMeta,
							OrderType:     execute.ORDER_TYPE_POST_ONLY,
							OrderMode:     execute.ORDER_BUY_OPEN,
						}); err != nil {
						toUpBitListDataStatic.DyLog.GetLog().Errorf("每秒limit_maker订单失败: %v", err)
					}
					time.Sleep(40 * time.Microsecond) // 休眠 40 微秒
				}
			}
		}
	})
	s.firstPriceBuy = limit.Mul(dec103).Truncate(s.PScale)
	s.thisOrderAccountId.Store(0) // 当前订单使用的资金账户ID
}
