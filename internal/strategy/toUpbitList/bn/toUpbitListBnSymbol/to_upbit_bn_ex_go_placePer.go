package toUpbitListBnSymbol

import (
	"time"

	"upbitBnServer/internal/quant/exchanges/binance/bnConst"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitBnMode"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"

	"github.com/shopspring/decimal"
)

func (s *Single) placePer(i int32, accountIndex uint8) {
	var j int
	var post, limit, market int
	var maxNotional decimal.Decimal //这一秒的能开的仓位的上限,第3秒之后要判断钱够不够开
	if i >= 3 {
		val := s.secondArr[accountIndex].maxNotional.Load()
		if val == nil {
			maxNotional = s.maxNotional
		} else {
			maxNotional = val.(decimal.Decimal)
		}
	}
	defer func() {
		toUpBitListDataStatic.DyLog.GetLog().Infof("账户[%d_%d],抽奖[总:%d,maker:%d,limit:%d,market:%d]次,上限[%s],协程结束",
			accountIndex, i, j, post, limit, market, maxNotional)
	}()
	tsSec := time.Now().Unix()   //该秒的开始时间戳,1760516599
	hasReceive := false          //没有收到标记价格
	var priceBuy decimal.Decimal //应该下单的价格

	// 上一秒有标记价格,预测这一秒为上一秒的1.03倍
	// 上一秒没有标记价格,用第一次的价格的1.03倍
	if markPrice_u10, ok := s.trigPriceMax_10.Load(tsSec - 1); ok {
		priceBuy = decimal.New(int64(markPrice_u10), -bnConst.PScale_10).Mul(dec103).Truncate(s.pScale)
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
			if s.secondArr[accountIndex].loadStop() || s.hasAllFilled.Load() {
				break OUTER
			}

			orderType := execute.ORDER_TYPE_POST_ONLY
			if s.hasTreeNews {
				// 上一次循环没有收到这一秒的标记价格
				if !hasReceive {

					// 拿到了新的价格上限,更新买入价格
					priceMaxBuy_10, ok := s.trigPriceMax_10.Load(tsSec)
					if ok {
						hasReceive = true
						priceBuy = decimal.New(int64(priceMaxBuy_10), -bnConst.PScale_10).Truncate(s.pScale)
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

			if i == 1 {
				orderType = execute.ORDER_TYPE_POST_ONLY
			}

			if toUpbitBnMode.Mode.ShouldExitOnTakeProfit(priceBuy.InexactFloat64(), s.takeProfitPrice) {
				toUpBitListDataStatic.DyLog.GetLog().Infof("超出止盈价格:[买入价:%.8f,止盈价:%.8f],退出每秒下单协程", priceBuy.InexactFloat64(), s.takeProfitPrice)
				return
			}

			//0.3*(总仓位-当前仓位)
			num := (s.posTotalNeed.Sub(s.pos.GetTotalVol()).Mul(dec03))
			if i >= 3 {
				num = decimal.Min(num, maxNotional.Div(priceBuy))
			}

			// 每次只开剩余应开仓位
			if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(order_from, accountIndex, s.symbolIndex,
				&orderModel.MyPlaceOrderReq{
					OrigPrice:     priceBuy,
					OrigVol:       num.Truncate(s.qScale),
					ClientOrderId: toUpBitListDataStatic.GetClientOrderIdBy("second"),
					StaticMeta:    s.StMeta,
					OrderType:     orderType,
					OrderMode:     execute.ORDER_BUY_OPEN,
				}); err != nil {
				toUpBitListDataStatic.DyLog.GetLog().Errorf("每秒创建订单失败: %v", err)
			}
			time.Sleep(150 * time.Microsecond) // 休眠 300 微秒
		}
	}
}
