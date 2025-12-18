package toUpbitListBnSymbol

import (
	"time"
	"upbitBnServer/internal/infra/observe/notify/notifyTg"
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
