package toUpbitListBnSymbol

import (
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListPos"

	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"

	"github.com/shopspring/decimal"
)

var (
	dec03    decimal.Decimal                     //首次下 0.3
	dec103   = decimal.RequireFromString("1.03") //第一次下单价格1.03倍
	qtyTotal decimal.Decimal                     //单次下单总金额
)

func SetParam(qty, dec003 float64) {
	qtyTotal = decimal.NewFromFloat(qty)
	dec03 = decimal.NewFromFloat(dec003)
}

/**
limit_maker协程,用到成员变量
posTotalNeed:在这里赋值,后面都只读
pScale: 多线程读安全
maxNotional: 初始化赋值之后都只读不写
secondArr: 不修改指针就安全
ctxStop:
hasAllFilled:
symbolIndex:
StMeta: 多线程读安全
firstPriceBuy:
thisOrderAccountId:

**/

/*
限制1: maxNotional 单账户单品种最大开仓上限
限制2：qtyTotal*0.3

退出循环:
1、账户没钱(不一定准)
*/

func (s *Single) PlacePostOnlyOrder(limit decimal.Decimal) {
	s.posTotalNeed = qtyTotal.Div(limit).Truncate(s.qScale)

	// maker抽奖金额为 min(单账户单品种最大开仓上限,需开参数价值)
	orderNum := decimal.Min(s.maxNotional, qtyTotal.Mul(dec03)).Div(limit).Truncate(s.qScale)

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
					if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(order_from, 0, s.symbolIndex,
						&orderModel.MyPlaceOrderReq{
							OrigPrice:     limit,
							OrigVol:       orderNum,
							ClientOrderId: toUpBitDataStatic.GetMakerClientOrderId(),
							StaticMeta:    s.StMeta,
							OrderType:     execute.ORDER_TYPE_POST_ONLY,
							OrderMode:     execute.ORDER_BUY_OPEN,
						}); err != nil {
						toUpBitDataStatic.DyLog.GetLog().Errorf("每秒limit_maker订单失败: %v", err)
					}
					// time.Sleep(40 * time.Microsecond) // 休眠 40 微秒
				}
			}
		}
	})
	s.pos = toUpbitListPos.NewPosCal()
	s.firstPriceBuy = limit.Mul(dec103).Truncate(s.pScale)
}
