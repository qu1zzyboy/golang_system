package toUpbitBybitSymbol

import (
	"math"
	"time"
	"upbitBnServer/internal/cal/u64Cal"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/byBitOrderAppManager"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitBnMode"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitParam"
	"upbitBnServer/pkg/utils/time2str"
)

func (s *Single) placePer(i int32, accountIndex uint8) {
	var j int
	var post, limit, market int
	var maxNotional float64 //这一秒的能开的仓位的上限,第3秒之后要判断钱够不够开
	if i >= 3 {
		// val := s.secondArr[accountIndex].maxNotional.Load()
		// if val == nil {
		// 	maxNotional = s.maxNotional
		// } else {
		// 	maxNotional = val.(float64)
		// }
	}
	defer func() {
		toUpBitDataStatic.DyLog.GetLog().Infof("账户[%d],抽奖[总:%d,maker:%d,limit:%d,market:%d]次,上限[%.2f],协程结束",
			accountIndex, j, post, limit, market, maxNotional)
	}()
	tsSec := time.Now().Unix() //该秒的开始时间戳,1760516599
	hasReceive := false        //没有收到标记价格
	var priceBuy float64       //应该下单的价格

	// 上一秒有标记价格,预测这一秒为上一秒的1.03倍
	// 上一秒没有标记价格,用第一次的价格的1.03倍
	if buyMaxPrice, ok := s.trigPriceMax.Load(tsSec - 1); ok {
		priceBuy = buyMaxPrice * 1.03
	} else {
		priceBuy = s.priceMaxBuy * 1.03
	}
OUTER:
	for j = 0; j <= 10; j++ {
		select {
		case <-s.ctxStop.Done():
			toUpBitDataStatic.DyLog.GetLog().Infof("收到关闭信号,退出每秒下单协程")
			break OUTER
		default:
			//有成交或者本轮挂单成功
			if s.hasAllFilled.Load() {
				break OUTER
			}
			orderMode := execute.BUY_OPEN_LIMIT_MAKER
			if s.hasTreeNews {
				// 上一次循环没有收到这一秒的标记价格
				if !hasReceive {

					// 拿到了新的价格上限,更新买入价格
					buyMaxPrice, ok := s.trigPriceMax.Load(tsSec)
					if ok {
						hasReceive = true
						priceBuy = buyMaxPrice
					}
				}
				// 再次判断是否拿到了标记价格
				if hasReceive {
					orderMode = execute.BUY_OPEN_LIMIT
					limit++
				} else {
					orderMode = execute.BUY_OPEN_MARKET
					market++
				}
			} else {
				post++
			}

			if toUpbitBnMode.Mode.ShouldExitOnTakeProfit(priceBuy, s.takeProfitPrice) {
				toUpBitDataStatic.DyLog.GetLog().Infof("超出止盈价格:[买入价:%.8f,止盈价:%.8f],退出每秒下单协程", priceBuy, s.takeProfitPrice)
				return
			}

			//0.3*(总仓位-当前仓位)
			num := toUpbitParam.F03 * (s.posTotalNeed - s.getPosLong())
			if i >= 3 {
				num = math.Min(num, maxNotional/(priceBuy))
			}

			// 每次只开剩余应开仓位
			if err := byBitOrderAppManager.GetTradeManager().SendPlaceOrder(accountIndex, orderModel.MyPlaceOrderReq{
				SymbolName:    s.symbolName,
				ClientOrderId: time2str.GetNowTimeStampMicroSlice16(),
				Pvalue:        u64Cal.FromF64(priceBuy, s.pScale.Uint8()),
				Qvalue:        u64Cal.FromF64(num, s.qScale.Uint8()),
				Pscale:        s.pScale,
				Qscale:        s.qScale,
				OrderMode:     orderMode,
				SymbolIndex:   s.symbolIndex,
				SymbolLen:     s.symbolLen,
				ReqFrom:       instanceEnum.TO_UPBIT_LIST_BYBIT,
				UsageFrom:     to_upbit_main,
			}); err != nil {
				toUpBitDataStatic.DyLog.GetLog().Errorf("每秒创建订单失败: %v", err)
			}
			time.Sleep(300 * time.Microsecond) // 休眠 300 微秒
		}
	}
}
