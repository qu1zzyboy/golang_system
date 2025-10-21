package orderSdkBnWsSign

import (
	"github.com/hhh500/quantGoInfra/pkg/container/pool/byteBufPool"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnModel"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/wsRequestCache"
)

func (s *FutureClient) QueryOrder(reqFrom orderBelongEnum.Type, api *orderSdkBnModel.FutureQuerySdk) error {
	rawData, err := api.ParseWsReqFast("Q", s.apiKey, "order.status", s.secretByte)
	defer byteBufPool.ReleaseBuffer(rawData)
	if rawData == nil || err != nil {
		return err
	}
	if err = s.conn.WriteAsync(*rawData); err != nil {
		return err
	}
	wsRequestCache.GetCache().StoreMeta("Q"+api.ClientOrderId, &wsRequestCache.WsRequestMeta{
		Json:    string(*rawData),
		ReqType: wsRequestCache.QUERY_ORDER,
		ReqFrom: reqFrom,
	})
	return err
}

// {
// 	"id": "s7-fP-btc7304766620344156161_query",
// 	"status": 200,
// 	"result": {
// 		"orderId": 615801527858,
// 		"symbol": "BTCUSDT",
// 		"status": "NEW",
// 		"clientOrderId": "s7-fP-btc7304766620344156161",
// 		"origPrice": "80000.00",
// 		"avgPrice": "0.00",
// 		"origQty": "0.002",
// 		"executedQty": "0.000",
// 		"cumQuote": "0.00000",
// 		"timeInForce": "GTX",
// 		"type": "LIMIT",
// 		"reduceOnly": false,
// 		"closePosition": false,
// 		"side": "BUY",
// 		"positionSide": "LONG",
// 		"stopPrice": "0.00",
// 		"workingType": "CONTRACT_PRICE",
// 		"priceProtect": false,
// 		"origType": "LIMIT",
// 		"priceMatch": "NONE",
// 		"selfTradePreventionMode": "EXPIRE_MAKER",
// 		"goodTillDate": 0,
// 		"time": 1741592078302,  //订单时间
// 		"updateTime": 1741592078316 //更新时间
// 	},
// 	"rateLimits": [{
// 		"rateLimitType": "REQUEST_WEIGHT",
// 		"interval": "MINUTE",
// 		"intervalNum": 1,
// 		"limit": 2400,
// 		"count": 6
// 	}]
// }
