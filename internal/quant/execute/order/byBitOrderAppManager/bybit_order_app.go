package byBitOrderAppManager

import (
	"context"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/quant/account/accountConfig"
	"upbitBnServer/internal/quant/execute/order/orderSdk/bybit/orderSdkBybitWs"
	"upbitBnServer/internal/quant/execute/order/wsRequestCache"
	"upbitBnServer/internal/strategy/toUpbitList/bybit/toUpbitByBitWsParse"
)

type OrderApp struct {
	wsOrder      *orderSdkBybitWs.FutureClient // WS API 客户端
	accountKeyId uint8                         // 账户序号
}

func newOrderApp() *OrderApp {
	return &OrderApp{}
}

func (s *OrderApp) init(ctx context.Context, v accountConfig.Config) error {
	s.accountKeyId = v.AccountId
	s.wsOrder = orderSdkBybitWs.NewFutureClient(v.ApiKeyHmac, v.SecretHmac)
	if err := s.wsOrder.RegisterReadHandler(ctx, v.AccountId, s.OnWsResp); err != nil {
		return err
	}
	return nil
}

func (s *OrderApp) OnWsResp(data []byte) {
	var reqId systemx.WsId16B
	copy(reqId[:], data[10:26])

	reqOk := data[38] == '0'
	wsMeta, ok := wsRequestCache.GetCache().GetMeta(reqId)
	if !ok {
		// 帐号认证的json
		if reqOk {
			return
		}
		dynamicLog.Error.GetLog().Errorf("[%d]WS_REQUEST: [%s]  not found %s", s.accountKeyId, string(reqId[:]), string(data))
		return
	}
	// 最终删除ws_meta
	defer wsRequestCache.GetCache().DelMeta(reqId)

	switch wsMeta.ReqFrom {
	case instanceEnum.TO_UPBIT_LIST_BYBIT:
		toUpbitByBitWsParse.Get().Parse(data, wsMeta, reqId, s.accountKeyId)
	case instanceEnum.TEST:
	default:
		dynamicLog.Error.GetLog().Errorf("WS_REQUEST: unknown ReqFrom %s", wsMeta.ReqFrom.String())
	}
}

// func (s *FutureClient) handleWsData(ctx context.Context, data []byte) {
// 	staticLog.Log.Debugf("接收到%s的ws数据:%s", s.accountKey, string(data))
// 	code := gjson.GetBytes(data, "retCode").Int()

// 	// clientOrderId, action := getClientOrderIdAndAction(data)
// 	if code == 0 {
// 		if gjson.GetBytes(data, "op").String() == "order.create" {
// 			su := orderModel.AcquireUnifiedOrder()
// 			su.ClientOrderId = gjson.GetBytes(data, "data.orderLinkId").String()
// 			su.OrderStatus = bybitConst.NEW
// 			su.AccountKeyId = s.accountKey
// 			orderSuccessCenter.GetService().DispatchOrder(ctx, s.accountKey, su)
// 			return
// 		}
// 	} else {
// 		fa := orderModel.NewFailedOrder(clientOrderId, convertx.ToString(code), gjson.GetBytes(data, "retMsg").String())
// 		if !bnConst.IsOrderErrCodeFilter(fa.ErrReason) {
// 			if jsonReq, ok := s.doJson.Load(clientOrderId); ok {
// 				dynamicLog.Error.GetLog().Errorf("ws_request接收到%s失败,请求数据:%s,返回数据:%s", s.accountKey, jsonReq, string(data))
// 			}
// 		}
// 		fa.From = p_WS_REQUEST + "_" + action
// 		orderFailureCenter.GetService().DispatchOrder(ctx, s.accountKey, fa)
// 		s.doJson.Delete(clientOrderId)
// 	}
// }

//{"retCode":0,"retMsg":"OK","op":"auth","connId":"d2a655dec3otvpcdvhjg-1qn6w"}

// {
// 	"retCode": 0,
// 	"retMsg": "OK",
// 	"op": "order.create",
// 	"data": {
// 		"orderId": "2099726f-9548-448b-84ab-96ea8ce6f0d0",
// 		"orderLinkId": "test12345678"
// 	},
// 	"retExtInfo": {},
// 	"header": {
// 		"X-Bapi-Limit": "20",
// 		"X-Bapi-Limit-Status": "19",
// 		"X-Bapi-Limit-Reset-Timestamp": "1755511671879",
// 		"Traceid": "bc1qskrh34w6zzpc6js0my0ccptyqx3r0r7tydt67d",
// 		"Timenow": "1755511671880"
// 	},
// 	"connId": "d2a655dec3otvpcdvhjg-1qn6w"
// }

// {
// 	"retCode": 0,
// 	"retMsg": "OK",
// 	"op": "order.amend",
// 	"data": {
// 		"orderId": "2099726f-9548-448b-84ab-96ea8ce6f0d0",
// 		"orderLinkId": "test12345678"
// 	},
// 	"retExtInfo": {},
// 	"header": {
// 		"X-Bapi-Limit": "20",  //該類型請求的帳戶總頻率
// 		"X-Bapi-Limit-Status": "19",  //該類型請求的帳戶剩餘可用頻率
// 		"X-Bapi-Limit-Reset-Timestamp": "1755511672867",
// 		"Traceid": "5c071068b1d94a6651d1bf809932812c",
// 		"Timenow": "1755511672867"
// 	},
// 	"connId": "d2a655dec3otvpcdvhjg-1qn6w"
// }

// {
// 	"retCode": 0,
// 	"retMsg": "OK",
// 	"op": "order.cancel",
// 	"data": {
// 		"orderId": "2099726f-9548-448b-84ab-96ea8ce6f0d0",
// 		"orderLinkId": "test12345678"
// 	},
// 	"retExtInfo": {},
// 	"header": {
// 		"Timenow": "1755511673867",
// 		"X-Bapi-Limit": "20",
// 		"X-Bapi-Limit-Status": "19",
// 		"X-Bapi-Limit-Reset-Timestamp": "1755511673867",
// 		"Traceid": "7359e714faab8bcd585eaf0190fe2518"
// 	},
// 	"connId": "d2a655dec3otvpcdvhjg-1qn6w"
// }
