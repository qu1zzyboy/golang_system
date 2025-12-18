package toUpbitListBnSymbol

import (
	"time"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/observe/notify/notifyTg"
	"upbitBnServer/internal/infra/safex"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"

	"github.com/shopspring/decimal"
)

const (
	f_per   = 25.0
	buy_sec = 25
)

var (
	clientOrderArr [buy_sec]string
)

func (s *Single) initBnSpotBuyOpen() {
	f14 := s.TrigMartPrice * 1.14
	f15 := s.TrigMartPrice * 1.15
	f16 := s.TrigMartPrice * 1.16
	posLeft := s.PosTotalNeed.Sub(s.Pos.GetTotal()).InexactFloat64()
	perAvg := posLeft / f_per
	s.bnSpotPerNum = decimal.NewFromFloat(perAvg).Truncate(s.QScale)

	var req = &orderModel.MyPlaceOrderReq{
		OrigVol:    s.bnSpotPerNum,
		StaticMeta: s.StMeta,
		OrderType:  execute.ORDER_TYPE_LIMIT,
		OrderMode:  execute.ORDER_BUY_OPEN,
	}

	for i := range 25 {
		clientOrderId := toUpBitDataStatic.GetClientOrderIdBy("server_bn_sp")
		req.ClientOrderId = clientOrderId
		clientOrderArr[i] = clientOrderId
		switch i % 3 {
		case 0:
			req.OrigPrice = decimal.NewFromFloat(f14).Truncate(s.pScale)
		case 1:
			req.OrigPrice = decimal.NewFromFloat(f15).Truncate(s.pScale)
		case 2:
			req.OrigPrice = decimal.NewFromFloat(f16).Truncate(s.pScale)
		}
		if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(order_from, 0, s.SymbolIndex, req); err != nil {
			toUpBitDataStatic.DyLog.GetLog().Errorf("bnSpot 初始化订单失败: %v", err)
		}
	}
}

func (s *Single) tryBuyLoopBnSpot(max int32) {
	//开启每秒抢一次的协程,来抢未来十秒的订单
	safex.SafeGo("to_upbit_bn_open_second", func() {
		var i int32
		defer func() {
			toUpBitDataStatic.DyLog.GetLog().Infof("每秒抽奖协程结束,抽奖次数[当前抽奖序号:%d,max:%d]", i, max)
		}()
		for i = 3; i < max; i++ {
			if i >= 4 {
				s.isStopLossAble.Store(true)
			}
			select {
			case <-s.ctxStop.Done():
				toUpBitDataStatic.DyLog.GetLog().Infof("收到关闭信号,退出每秒抽奖协程")
				return
			case <-s.bnSpotCtxStop.Done():
				toUpBitDataStatic.DyLog.GetLog().Infof("收到19%%关闭信号,退出每秒抽奖协程")
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

func (s *Single) buyLoop() {
	var i = 0
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		i++
		if i >= buy_sec {
			ticker.Stop()
			return
		}
		val := s.bidPrice.Load()
		if val == nil {
			continue
		}
		if err := bnOrderAppManager.GetTradeManager().SendModifyOrder(order_from, 0, &orderModel.MyModifyOrderReq{
			ModifyPrice:   decimal.NewFromFloat(val.(float64)).Truncate(s.pScale),
			OrigVol:       s.bnSpotPerNum,
			StaticMeta:    s.StMeta,
			ClientOrderId: clientOrderArr[i],
			OrderMode:     execute.ORDER_BUY_OPEN,
		}); err != nil {
			notifyTg.GetTg().SendToUpBitMsg(map[string]string{
				"symbol": s.StMeta.SymbolName,
				"op":     "更新bnSpot订单失败",
				"error":  err.Error(),
			})
			toUpBitDataStatic.DyLog.GetLog().Errorf("%s 更新bnSpot订单失败: %s", s.StMeta.SymbolName, err.Error())
		}
	}
}
