package toUpbitListBnSymbol

import (
	"time"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListPos"

	"github.com/shopspring/decimal"
)

func (s *Single) placePostOnlyOrder(limit decimal.Decimal) {
	s.PosTotalNeed = qtyTotal.Div(limit).Truncate(s.QScale)

	// maker抽奖金额为 min(单账户单品种最大开仓上限,需开参数价值)
	orderNum := decimal.Min(s.MaxNotional, qtyTotal.Mul(dec03)).Div(limit).Truncate(s.QScale)

	safex.SafeGo("to_upbit_bn_limit_maker", func() {
		s.SecondArr[0].start()
		var i int
		defer func() {
			toUpBitDataStatic.DyLog.GetLog().Infof("账户[%d],下单[%d,%s]次 10ms POST_ONLY 协程结束", 0, i+1, limit.String())
		}()
	OUTER:
		for i = 0; i <= toUpBitDataStatic.MAX_BUY_COUNT_PER; i++ {
			select {
			case <-s.ctxStop.Done():
				toUpBitDataStatic.DyLog.GetLog().Infof("收到关闭信号,退出POST_ONLY抽奖协程")
				break OUTER
			default:
				{
					//完全成交或者本轮挂单成功
					if s.SecondArr[0].loadStop() || s.hasAllFilled.Load() {
						break OUTER
					}
					if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(order_from, 0, s.SymbolIndex, &orderModel.MyPlaceOrderReq{
						OrigPrice:     limit,
						OrigVol:       orderNum,
						ClientOrderId: toUpBitDataStatic.GetMakerClientOrderId(),
						StaticMeta:    s.StMeta,
						OrderType:     execute.ORDER_TYPE_POST_ONLY,
						OrderMode:     execute.ORDER_BUY_OPEN,
					}); err != nil {
						toUpBitDataStatic.DyLog.GetLog().Errorf("每秒limit_maker订单失败: %v", err)
					}
				}
			}
		}
	})
	s.Pos = toUpbitListPos.NewPosCal()
	s.FirstPriceBuy = limit.Mul(dec103).Truncate(s.pScale)
}

func (s *Single) tryBuyLoopBeforeNews() {
	//开启每秒抢一次的协程,来抢未来十秒的订单
	safex.SafeGo("to_upbit_bn_open_second", func() {
		var i int32
		defer func() {
			toUpBitDataStatic.DyLog.GetLog().Infof("每秒抽奖协程结束,抽奖次数[当前抽奖序号:%d,max:%d]", i, 3)
		}()
		for i = 1; i < 5; i++ {
			select {
			case <-s.ctxStop.Done():
				toUpBitDataStatic.DyLog.GetLog().Infof("收到关闭信号,退出每秒抽奖协程")
				return
			default:
				// 睡到下一秒的5毫秒后
				now := time.Now()
				secStart := now.Truncate(time.Second)
				target := secStart.Add(965 * time.Millisecond)

				// 如果已经超过 965ms，就睡到下一秒的 965ms
				if !now.Before(target) {
					target = target.Add(time.Second)
				}
				time.Sleep(time.Until(target))

				//已经完全开满
				if s.hasAllFilled.Load() {
					toUpBitDataStatic.DyLog.GetLog().Infof("完全成交,退出每秒抽奖协程")
					break
				}
				if s.hasTreeNews.Load() {
					toUpBitDataStatic.DyLog.GetLog().Infof("收到 TreeNews,退出每秒抽奖协程")
					break
				}
				// 进入每秒抽奖循环
				placeIndex := uint8(getCurIndex(i))           // 该秒的下单账户id
				s.SecondArr[placeIndex].start()               // 重置该秒状态
				s.thisOrderAccountId.Store(int32(placeIndex)) // 当前订单使用的资金账户Id
				fromAccountId := getPreIndex(i)               // 该秒的撤单账户id
				s.toAccountId.Store(trans[fromAccountId])     // 当前应该接收资金的账户,新的一秒开始就更新

				dynamicLog.Log.GetLog().Infof("==========[循环序号:%d,下单账户:%d,撤单账户:%d]秒下单=========", i, placeIndex, fromAccountId)

				// 撤销上一轮的订单
				go s.cancelAndTransfer(i, fromAccountId)

				//探测逻辑
				go s.monitorPer(placeIndex)

				//真实下单逻辑
				go s.placePer(i, placeIndex)
			}
		}
	})
}
