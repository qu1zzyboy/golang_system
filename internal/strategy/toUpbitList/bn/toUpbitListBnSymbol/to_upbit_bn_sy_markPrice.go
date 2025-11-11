package toUpbitListBnSymbol

import (
	"errors"
	"time"

	"upbitBnServer/internal/infra/observe/notify/notifyTg"
	"upbitBnServer/internal/quant/exchanges/binance/bnConst"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/quant/market/symbolInfo/coinMesh"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolDynamic"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolLimit"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitBnMode"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/pkg/container/map/myMap"
	"upbitBnServer/pkg/container/pool/antPool"
	"upbitBnServer/pkg/container/pool/byteBufPool"
	"upbitBnServer/pkg/utils/convertx"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

var (
	dec12              = decimal.RequireFromString("12.00")     //2倍最小下单金额
	dec2               = decimal.RequireFromString("2.00")      //2倍最小下单金额
	dec5               = decimal.RequireFromString("0.33")      //小订单比例
	dec1               = decimal.RequireFromString("1.0")       //1.0
	clientOrders       = myMap.NewMySyncMap[string, struct{}]() //clientOrderId-->占位符,所有的挂单状态的订单
	clientOrderSig     = myMap.NewMySyncMap[string, struct{}]() //clientOrderId-->占位符,有就不下单
	ClientOrderIsCheck = myMap.NewMySyncMap[string, struct{}]() //clientOrderId-->占位符,有就不检查
)

func OnOrderUpdate(isOnline bool, clientOrderId string) {
	// 挂单状态就存,非挂单状态就删
	if isOnline {
		clientOrders.Store(clientOrderId, struct{}{})
	} else {
		clientOrders.Delete(clientOrderId)
	}
}

func (s *Single) CancelPreOrder() {
	if !s.hasInit {
		return
	}
	bnOrderAppManager.GetTradeManager().SendCancelOrder(ws_req_from, s.preAccountKeyId, &orderModel.MyQueryOrderReq{
		ClientOrderId: s.clientOrderIdSmall,
		StaticMeta:    s.StMeta,
	})
}

func (s *Single) onMarkPrice(len int, bufPtr *[]byte) {
	defer byteBufPool.ReleaseBuffer(bufPtr)
	data := (*bufPtr)[:len]

	if toUpBitListDataAfter.LoadTrig() {
		/*********************上币已经触发**************************/
		if s.symbolIndex != toUpBitListDataAfter.TrigSymbolIndex {
			return
		}
		// 1、计算价格上限并存储
		results := gjson.GetManyBytes(data, "p", jsonEvent)
		markPrice_u8 := convertx.PriceStringToUint64(results[0].String(), bnConst.PScale_8)
		markPrice_u10 := markPrice_u8 * s.upLimitPercent_2
		s.trigPriceMax_10.Store(results[1].Int()/1000, markPrice_u10)
		toUpBitDataStatic.DyLog.GetLog().Infof("%s最新[u8:%d,u10:%d]标记价格: %s", s.StMeta.SymbolName, markPrice_u8, markPrice_u10, string(data))
	} else {
		results := gjson.GetManyBytes(data, "p", jsonEvent)
		// 1、计算网络接受延迟
		markPrice_8 := convertx.PriceStringToUint64(results[0].String(), bnConst.PScale_8)
		s.markPriceTs = results[1].Int()
		s.minPriceAfterMp = markPrice_8
		// 2、计算价格上限
		s.markPrice_8 = markPrice_8
		s.priceMaxBuy_10 = markPrice_8 * s.upLimitPercent_2
		s.mpLatencyTotal.Record(s.StMeta.SymbolName, float64(time.Now().UnixMicro()-1000*s.markPriceTs))
		if !toUpbitBnMode.Mode.IsPlacePreOrder() {
			return
		}
		// 3、回调函数更新预挂单
		s.checkPreOrder(markPrice_8)
	}
}

func (s *Single) onPreFilled(clientOrderId string) {
	_ = antPool.SubmitToIoPool("预挂单成交"+clientOrderId, func() {
		if _, ok := clientOrders.Load(clientOrderId); ok {
			// 是属于预挂单的订单,5秒后删除订单标记,防止频繁成交
			ClientOrderIsCheck.Delete(clientOrderId)
			clientOrders.Delete(clientOrderId)
			clientOrderSig.Store(clientOrderId, struct{}{})
			// 下买入平空单
			if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(ws_req_from, s.preAccountKeyId, s.symbolIndex,
				&orderModel.MyPlaceOrderReq{
					OrigPrice:     decimal.New(int64(s.lastMarkPrice_8), -bnConst.PScale_8).Truncate(s.pScale),
					OrigVol:       s.orderNum,
					ClientOrderId: toUpBitDataStatic.GetClientOrderIdBy("close_pre"),
					StaticMeta:    s.StMeta,
					OrderType:     execute.ORDER_TYPE_MARKET,
					OrderMode:     execute.ORDER_BUY_CLOSE,
				}); err != nil {
				symbolKey := symbolStatic.GetSymbol().GetSymbolKey(s.StMeta.SymbolKeyId)
				notifyTg.GetTg().SendToUpBitMsg(map[string]string{
					"symbol": symbolKey,
					"op":     "下买入平空单失败",
				})
				toUpBitDataStatic.DyLog.GetLog().Errorf("%s下买入平空单错误: %s", symbolKey, err.Error())
			}
			// 等待能再次下单
			time.Sleep(60 * time.Second)
			toUpBitDataStatic.DyLog.GetLog().Infof("预挂单成交5秒后,删除订单限流标记:%s", clientOrderId)
			clientOrderSig.Delete(clientOrderId)
		}
	})
}

func (s *Single) initPreOrder() error {
	dyMeta, err := symbolDynamic.GetManager().Get(s.StMeta.SymbolKeyId)
	if err != nil {
		return err
	}
	mesh, ok := coinMesh.GetManager().Get(s.StMeta.TradeId)
	if !ok {
		return errors.New("coinMesh.GetManager().Get not found")
	}

	limit, err := symbolLimit.GetManager().Get(s.StMeta.SymbolKeyId)
	if err != nil {
		toUpBitDataStatic.DyLog.GetLog().Errorf("symbolKeyId %d not found", s.StMeta.SymbolKeyId)
		return err
	}
	lastMarkPriceDec := decimal.New(int64(s.lastMarkPrice_8), -bnConst.PScale_8)

	//挂单量=最大(最小下单量,2*最小下单金额/最新标记价格)
	s.orderNum = decimal.Max(dyMeta.LotSize, dec12.Mul(dyMeta.MinQty).Div(lastMarkPriceDec)).Truncate(dyMeta.QScale)
	s.clientOrderIdSmall = toUpBitDataStatic.GetClientOrderIdBy(mesh.CmcAsset) //小订单id
	//1.0+0.33*(1.15-1.0)
	s.smallPercent = dec5.Mul(limit.UpLimitPercent.Sub(dec1)).Add(dec1)
	s.hasInit = true
	return nil
}
