package toUpBitListBnExecute

import (
	"time"

	"github.com/hhh500/quantGoInfra/infra/observe/log/dynamicLog"
	"github.com/hhh500/quantGoInfra/infra/safex"
	"github.com/hhh500/quantGoInfra/quant/exchanges/binance/bnConst"
	"github.com/hhh500/upbitBnServer/internal/quant/execute"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderModel"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"github.com/shopspring/decimal"
)

var (
	decimal50 = decimal.NewFromFloat(2) //刻意超出价格上限的
	trans     = [11]int32{
		3,  //账户0-->账户3
		4,  //账户1-->账户4
		5,  //账户2-->账户5
		6,  //账户3-->账户6
		7,  //账户4-->账户7
		8,  //账户5-->账户8
		9,  //账户6-->账户9
		10, //账户7-->账户10
		1,  //账户8-->账户1
		2,  //账户9-->账户2
		3,  //账户10-->账户3
	}
)

// 给定一个整数 i,算出它在 1–10 的“循环序列”里的前2个下标
func (s *Execute) getPreIndex(i int32) int32 {
	res := i%10 + 8
	if res > 10 {
		res = res - 10
	}
	return res
}

// 把任意整数 i 映射到 1–10 之间
func (s *Execute) getCurIndex(i int32) int32 {
	res := i % 10
	if res == 0 {
		res = 10
	}
	return res
}

func (s *Execute) TryBuyLoop(max int32) {
	//开启每秒抢一次的协程,来抢未来十秒的订单
	safex.SafeGo("toUpBitListBn_open_second", func() {
		var i int32
		defer func() {
			toUpBitListDataStatic.DyLog.GetLog().Infof("每秒抽奖协程结束,抽奖次数[当前抽奖序号:%d,max:%d]", i, max)
		}()
		for i = 1; i < max; i++ {
			select {
			case <-s.ctxStop.Done():
				toUpBitListDataStatic.DyLog.GetLog().Infof("收到关闭信号,退出每秒抽奖协程")
				return
			default:
				// 睡到下一秒的5毫秒后
				now := time.Now()
				next := now.Truncate(time.Second).Add(time.Second)
				trigger := next.Add(+5 * time.Millisecond)
				sleep := time.Until(trigger)
				if sleep > 0 {
					time.Sleep(sleep)
				}
				//已经完全开满
				if s.hasAllFilled.Load() {
					break
				}
				// 进入每秒抽奖循环
				placeIndex := uint8(s.getCurIndex(i))             // 该秒的下单账户id
				s.stopThisSecondPerArr[placeIndex].Store(false)   // 开启本轮抽奖信号
				s.hasInToSecondPerLoopArr[placeIndex].Store(true) // 确认进入了每秒抽奖循环
				s.thisOrderAccountId.Store(int32(placeIndex))     // 当前订单使用的资金账户Id
				fromAccountId := s.getPreIndex(i)                 // 该秒的撤单账户id
				s.toAccountId.Store(trans[fromAccountId])         // 当前应该接收资金的账户,新的一秒开始就更新

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

func (s *Execute) monitorPer(accountIndex uint8) {
	var i int
	defer func() {
		toUpBitListDataStatic.DyLog.GetLog().Infof("账户[%d],探测[%d]次,协程结束", accountIndex, i)
	}()
	price := s.firstPriceBuy.Mul(decimal50).Truncate(s.PScale)
OUTER:
	for i = 0; i <= 230; i++ {
		select {
		case <-s.ctxStop.Done():
			toUpBitListDataStatic.DyLog.GetLog().Infof("收到关闭信号,退出探测协程")
			break OUTER
		default:
			//有成交或者本轮挂单成功
			if s.stopThisSecondPerArr[accountIndex].Load() || s.hasAllFilled.Load() {
				break OUTER
			}
			if err := bnOrderAppManager.GetMonitorManager().SendMonitorOrder(order_from, accountIndex, s.symbolIndex,
				&orderModel.MyPlaceOrderReq{
					OrigPrice:     price,
					OrigVol:       s.posTotalNeed,
					ClientOrderId: toUpBitListDataStatic.GetClientOrderIdBy("sec-Mo"),
					StaticMeta:    s.StMeta,
					OrderType:     execute.ORDER_TYPE_LIMIT,
					OrderMode:     execute.ORDER_BUY_OPEN,
				}); err != nil {
				toUpBitListDataStatic.DyLog.GetLog().Errorf("每秒探测订单失败: %v", err)
			}
			time.Sleep(300 * time.Microsecond) // 休眠 300 微秒
		}
	}
}

func (s *Execute) placePer(i int32, accountIndex uint8) {
	//真实下单逻辑
	var j int
	var post, limit, market int
	var maxNotional decimal.Decimal
	if i >= 3 {
		val := s.maxNotionalArr[accountIndex].Load()
		if val == nil {
			maxNotional = s.maxNotional
		} else {
			maxNotional = val.(decimal.Decimal)
		}
	}
	defer func() {
		toUpBitListDataStatic.DyLog.GetLog().Infof("账户[%d],抽奖[总:%d,maker:%d,limit:%d,market:%d]次,上限[%s],协程结束",
			accountIndex, j, post, limit, market, maxNotional)
	}()
	tsSec := time.Now().Unix()   //该秒的开始时间戳,1760516599
	hasReceive := false          //没有收到标记价格
	var priceBuy decimal.Decimal //应该下单的价格

	// 上一秒有标记价格,预测这一秒为上一秒的1.03倍
	// 上一秒没有标记价格,用第一次的价格的1.03倍
	if markPrice_u10, ok := toUpBitListDataAfter.TrigPriceMax_10.Load(tsSec - 1); ok {
		priceBuy = decimal.New(int64(markPrice_u10), -bnConst.PScale_10).Mul(dec103).Truncate(s.PScale)
	} else {
		priceBuy = s.firstPriceBuy
	}
OUTER:
	for j = 0; j <= 230; j++ {
		select {
		case <-s.ctxStop.Done():
			toUpBitListDataStatic.DyLog.GetLog().Infof("收到关闭信号,退出每秒下单协程")
			break OUTER
		default:
			//有成交或者本轮挂单成功
			if s.stopThisSecondPerArr[accountIndex].Load() || s.hasAllFilled.Load() {
				break OUTER
			}

			orderType := execute.ORDER_TYPE_POST_ONLY
			if s.hasTreeNews.Load() {
				// 上一次循环没有收到这一秒的标记价格
				if !hasReceive {
					markPrice_u10, ok := toUpBitListDataAfter.TrigPriceMax_10.Load(tsSec)
					if ok {
						// 拿到了新的价格上限,更新买入价格
						hasReceive = true
						priceBuy = decimal.New(int64(markPrice_u10), -bnConst.PScale_10).Truncate(s.PScale)
					}
				}
				// 再次判断是否拿到了标记价格
				if hasReceive {
					orderType = execute.ORDER_TYPE_LIMIT
					limit++
				} else {
					orderType = execute.ORDER_TYPE_MARKET
					market++
				}
			} else {
				post++
			}

			if !toUpBitListDataStatic.IsDebug {
				// 超出价格限制,退出
				if s.takeProfitPrice > 0 && priceBuy.InexactFloat64() > s.takeProfitPrice {
					toUpBitListDataStatic.DyLog.GetLog().Infof("超出止盈价格:[买入价:%.8f,止盈价:%.8f],退出每秒下单协程", priceBuy.InexactFloat64(), s.takeProfitPrice)
					return
				}
			}

			num := (s.posTotalNeed.Sub(s.pos.GetTotal()).Mul(dec03))
			if i >= 3 {
				num = decimal.Min(num, maxNotional.Div(priceBuy))
			}

			// 每次只开剩余应开仓位
			if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(order_from, accountIndex, s.symbolIndex,
				&orderModel.MyPlaceOrderReq{
					OrigPrice: priceBuy,
					//0.3*(总仓位-当前仓位)
					OrigVol:       num.Truncate(s.QScale),
					ClientOrderId: toUpBitListDataStatic.GetClientOrderIdBy("second"),
					StaticMeta:    s.StMeta,
					OrderType:     orderType,
					OrderMode:     execute.ORDER_BUY_OPEN,
				}); err != nil {
				toUpBitListDataStatic.DyLog.GetLog().Errorf("每秒创建订单失败: %v", err)
			}
			time.Sleep(300 * time.Microsecond) // 休眠 300 微秒
		}
	}
}
