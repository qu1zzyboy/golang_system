package toUpbitListBnSymbol

import (
	"errors"
	"time"

	"github.com/hhh500/quantGoInfra/infra/observe/notify/notifyTg"
	"github.com/hhh500/quantGoInfra/pkg/container/map/myMap"
	"github.com/hhh500/quantGoInfra/pkg/container/pool/antPool"
	"github.com/hhh500/quantGoInfra/pkg/container/pool/byteBufPool"
	"github.com/hhh500/quantGoInfra/pkg/utils/convertx"
	"github.com/hhh500/quantGoInfra/quant/exchanges/binance/bnConst"
	"github.com/hhh500/upbitBnServer/internal/cal/u64Cal"
	"github.com/hhh500/upbitBnServer/internal/quant/execute"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderModel"
	"github.com/hhh500/upbitBnServer/internal/quant/market/symbolInfo/coinMesh"
	"github.com/hhh500/upbitBnServer/internal/quant/market/symbolInfo/symbolDynamic"
	"github.com/hhh500/upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

var (
	dec2               = decimal.RequireFromString("2.00")      //2倍最小下单金额
	dec5               = decimal.RequireFromString("0.33")      //小订单比例
	dec1               = decimal.RequireFromString("1.0")       //1.0
	clientOrders       = myMap.NewMySyncMap[string, struct{}]() //clientOrderId-->占位符,所有的挂单状态的订单
	clientOrderSig     = myMap.NewMySyncMap[string, struct{}]() //clientOrderId-->占位符,有就不下单
	ClientOrderIsCheck = myMap.NewMySyncMap[string, struct{}]() //clientOrderId-->占位符,有就不检查
)

func (s *Single) CancelPreOrder() {
	if !s.hasInit {
		return
	}
	bnOrderAppManager.GetTradeManager().SendCancelOrder(ws_req_from, s.accountKeyId, &orderModel.MyQueryOrderReq{
		ClientOrderId: s.clientOrderIdSmall,
		StaticMeta:    s.StMeta,
	})
}

func OnWsOrder(isOnline bool, clientOrderId string) {
	// 挂单状态就存,非挂单状态就删
	if isOnline {
		clientOrders.Store(clientOrderId, struct{}{})
	} else {
		clientOrders.Delete(clientOrderId)
	}
}

func (s *Single) onMarkPrice(len int, bufPtr *[]byte) {
	defer byteBufPool.ReleaseBuffer(bufPtr)
	data := (*bufPtr)[:len]

	if toUpBitListDataAfter.LoadTrig() {
		if s.symbolIndex != toUpBitListDataAfter.TrigSymbolIndex {
			return
		}
		/*********************上币已经触发**************************/
		// 1、计算价格上限并存储
		results := gjson.GetManyBytes(data, "p", jsonEvent)
		markPrice_u8 := convertx.PriceStringToUint64(results[0].String(), bnConst.PScale_8)
		markPrice_u10 := markPrice_u8 * s.upLimitPercent_2
		toUpBitListDataAfter.TrigPriceMax_10.Store(results[1].Int()/1000, markPrice_u10)
		toUpBitListDataStatic.DyLog.GetLog().Infof("%s最新[u8:%d,u10:%d]标记价格: %s", toUpBitListDataAfter.TrigSymbolName, markPrice_u8, markPrice_u10, string(data))
	} else {
		results := gjson.GetManyBytes(data, "p", jsonEvent)
		// 1、计算网络接受延迟
		markPrice_8 := convertx.PriceStringToUint64(results[0].String(), bnConst.PScale_8)
		eventTs := 1000 * results[1].Int()
		// 2、计算价格上限
		s.markPrice_8 = markPrice_8
		s.priceMax_10 = markPrice_8 * s.upLimitPercent_2
		s.mpLatencyTotal.Record(s.StMeta.SymbolName, float64(time.Now().UnixMicro()-eventTs))

		if toUpBitListDataStatic.IsDebug {
			return
		}
		// 3、回调函数更新预挂单
		if !s.hasInit {
			s.lastMarkPrice_8 = markPrice_8
			if err := s.initPreOrder(); err != nil {
				toUpBitListDataStatic.DyLog.GetLog().Errorf("%s 初始化交易对失败:  %s", s.StMeta.SymbolName, err.Error())
				return
			}
			//下小订单
			thisMarkPriceDec := decimal.New(int64(markPrice_8), -bnConst.PScale_8)
			if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(ws_req_from, s.accountKeyId, s.symbolIndex,
				&orderModel.MyPlaceOrderReq{
					OrigPrice:     thisMarkPriceDec.Mul(s.smallPercent).Truncate(s.pScale),
					OrigVol:       s.orderNum,
					ClientOrderId: s.clientOrderIdSmall,
					StaticMeta:    s.StMeta,
					OrderType:     execute.ORDER_TYPE_LIMIT,
					OrderMode:     execute.ORDER_SELL_OPEN,
				}); err != nil {
				toUpBitListDataStatic.DyLog.GetLog().Errorf("创建小订单错误: %s", err.Error())
			}
			return
		}
		// 价格距离上次变动有1%以上,开始移动订单
		if u64Cal.IsDiffMoreThanPercent100(markPrice_8, s.lastMarkPrice_8, 1) {
			// 缩小10的8次方倍
			thisMarkPriceDec := decimal.New(int64(markPrice_8), -bnConst.PScale_8)
			symbolKey := symbolStatic.GetSymbol().GetSymbolKey(s.StMeta.SymbolKeyId)
			modifySmallPrice := thisMarkPriceDec.Mul(s.smallPercent).Truncate(s.pScale)
			s.lastMarkPrice_8 = markPrice_8
			if _, ok := clientOrders.Load(s.clientOrderIdSmall); ok {
				toUpBitListDataStatic.DyLog.GetLog().Infof("[%d,%s] 触发[%d,%d,%d],准备更新预挂单:%s", s.accountKeyId, symbolKey, s.lastMarkPrice_8, markPrice_8, s.pScale, modifySmallPrice)
				//更新订单价格
				if err := bnOrderAppManager.GetTradeManager().SendModifyOrder(ws_req_from, s.accountKeyId,
					&orderModel.MyModifyOrderReq{
						ModifyPrice:   modifySmallPrice,
						OrigVol:       s.orderNum,
						StaticMeta:    s.StMeta,
						ClientOrderId: s.clientOrderIdSmall,
						OrderMode:     execute.ORDER_SELL_OPEN,
					}); err != nil {
					notifyTg.GetTg().SendToUpBitMsg(map[string]string{
						"symbol": symbolKey,
						"op":     "更新5%订单失败",
						"error":  err.Error(),
					})
					toUpBitListDataStatic.DyLog.GetLog().Errorf("%s修改小订单错误: %s", symbolKey, err.Error())
				}
			} else {
				toUpBitListDataStatic.DyLog.GetLog().Infof("[%d,%s] 触发[%d,%d,%d],准备重下预挂单:%s", s.accountKeyId, symbolKey, s.lastMarkPrice_8, markPrice_8, s.pScale, modifySmallPrice)
				if _, ok := clientOrderSig.Load(s.clientOrderIdSmall); ok {
					toUpBitListDataStatic.DyLog.GetLog().Infof("订单限流中,无法下单:%s", s.clientOrderIdSmall)
					return
				}
				//下小订单
				if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(ws_req_from, s.accountKeyId, s.symbolIndex,
					&orderModel.MyPlaceOrderReq{
						OrigPrice:     modifySmallPrice,
						OrigVol:       s.orderNum,
						ClientOrderId: s.clientOrderIdSmall,
						StaticMeta:    s.StMeta,
						OrderType:     execute.ORDER_TYPE_LIMIT,
						OrderMode:     execute.ORDER_SELL_OPEN,
					}); err != nil {
					notifyTg.GetTg().SendToUpBitMsg(map[string]string{
						"symbol": symbolKey,
						"op":     "预挂5%订单失败",
						"error":  err.Error(),
					})
					toUpBitListDataStatic.DyLog.GetLog().Errorf("%s创建小订单错误: %s", symbolKey, err.Error())
				}
			}
		}
	}
}

func (s *Single) onPreFilled(clientOrderId string) {
	antPool.SubmitToIoPool("预挂单成交"+clientOrderId, func() {
		if _, ok := clientOrders.Load(clientOrderId); ok {
			// 是属于预挂单的订单,5秒后删除订单标记,防止频繁成交
			clientOrders.Delete(clientOrderId)
			clientOrderSig.Store(clientOrderId, struct{}{})
			// 下买入平空单
			if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(ws_req_from, s.accountKeyId, s.symbolIndex,
				&orderModel.MyPlaceOrderReq{
					OrigPrice:     decimal.New(int64(s.lastMarkPrice_8), -bnConst.PScale_8).Truncate(s.pScale),
					OrigVol:       s.orderNum,
					ClientOrderId: toUpBitListDataStatic.GetClientOrderIdBy("close_pre"),
					StaticMeta:    s.StMeta,
					OrderType:     execute.ORDER_TYPE_LIMIT,
					OrderMode:     execute.ORDER_BUY_CLOSE,
				}); err != nil {
				symbolKey := symbolStatic.GetSymbol().GetSymbolKey(s.StMeta.SymbolKeyId)
				notifyTg.GetTg().SendToUpBitMsg(map[string]string{
					"symbol": symbolKey,
					"op":     "下买入平空单失败",
				})
				toUpBitListDataStatic.DyLog.GetLog().Errorf("%s下买入平空单错误: %s", symbolKey, err.Error())
			}
			// 等待能再次下单
			time.Sleep(5 * time.Second)
			toUpBitListDataStatic.DyLog.GetLog().Infof("预挂单成交5秒后,删除订单限流标记:%s", clientOrderId)
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
	lastMarkPriceDec := decimal.New(int64(s.lastMarkPrice_8), -bnConst.PScale_8)

	//挂单量=最大(最小下单量,2*最小下单金额/最新标记价格)
	s.orderNum = decimal.Max(dyMeta.LotSize, dec2.Mul(dyMeta.MinQty).Div(lastMarkPriceDec)).Truncate(dyMeta.QScale)
	s.clientOrderIdSmall = toUpBitListDataStatic.GetClientOrderIdBy(mesh.CmcAsset) //小订单id
	//1.0+0.33*(1.15-1.0)
	s.smallPercent = dec5.Mul(dyMeta.UpLimitPercent.Sub(dec1)).Add(dec1)
	s.hasInit = true
	return nil
}
