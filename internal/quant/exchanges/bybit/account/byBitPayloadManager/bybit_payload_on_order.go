package byBitPayloadManager

import (
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/quant/execute/order/orderStatic"
	"upbitBnServer/internal/strategy/toUpbitList/bybit/toUpbitByBitPayloadParse"
	"upbitBnServer/pkg/utils/byteUtils"
)

func (s *Payload) onPayloadOrder(data []byte) {
	totalLen := uint16(len(data))
	var clientOrderId systemx.WsId16B
	var id_start uint16 = 30
	id_end := byteUtils.FindNextQuoteIndex(data, id_start, totalLen)

	sy_start := id_end + 70
	sy_end := byteUtils.FindNextQuoteIndex(data, sy_start, totalLen)

	o_id_start := sy_end + 13
	o_id_end := byteUtils.FindNextQuoteIndex(data, o_id_start, totalLen)

	cid_start := o_id_end + 17
	cid_end := cid_start + systemx.ArrLen
	copy(clientOrderId[:], data[cid_start:cid_end])
	meta, ok := orderStatic.GetService().GetOrderMeta(clientOrderId)
	if !ok {
		dynamicLog.Error.GetLog().Errorf("ORDER_UPDATE: [%s] orderFrom not found %s", clientOrderId, string(data))
		return
	}

	switch meta.ReqFrom {
	case instanceEnum.TO_UPBIT_LIST_BYBIT:
		toUpbitByBitPayloadParse.OnPayloadOrder(data, clientOrderId, meta, totalLen, o_id_start, o_id_end, cid_end, s.accountKeyId)
	case instanceEnum.TEST:
	default:
		dynamicLog.Error.GetLog().Errorf("TRADE_LITE: unknown ReqFrom %v", meta.ReqFrom)
	}
}

// {"topic":"execution.fast.linear","creationTime":1763095434843,"data":[{"category":"linear","symbol":"ESPORTSUSDT","execId":"410c66db-0f20-5f2f-99a4-e20bada7a6ce","execPrice":"0.33656","execQty":"34","orderId":"8e91f790-0d27-4c26-92ec-d3467da90d07","isMaker":true,"orderLinkId":"","side":"Sell","execTime":"1763095434836","seq":262102669252}]}
// {"topic":"order.linear","id":"509665119_ESPORTSUSDT_262102669252","creationTime":1763095434847,"data":[{"category":"linear","symbol":"ESPORTSUSDT","orderId":"8e91f790-0d27-4c26-92ec-d3467da90d07","orderLinkId":"1763090785843843","blockTradeId":"","side":"Sell","positionIdx":2,"orderStatus":"Filled","cancelType":"UNKNOWN","rejectReason":"EC_NoError","timeInForce":"GTC","isLeverage":"","price":"0.33656","qty":"34","avgPrice":"0.33656","leavesQty":"0","leavesValue":"0","cumExecQty":"34","cumExecValue":"11.44304","cumExecFee":"0.00457722","orderType":"Limit","stopOrderType":"","orderIv":"","triggerPrice":"","takeProfit":"","stopLoss":"","triggerBy":"","tpTriggerBy":"","slTriggerBy":"","triggerDirection":0,"placeType":"","lastPriceOnCreated":"0.32268","closeOnTrigger":false,"reduceOnly":false,"smpGroup":0,"smpType":"None","smpOrderId":"","slLimitPrice":"0","tpLimitPrice":"0","tpslMode":"UNKNOWN","createType":"CreateByUser","marketUnit":"","createdTime":"1763090785846","updatedTime":"1763095434846","feeCurrency":"","closedPnl":"0","slippageTolerance":"0","slippageToleranceType":"UNKNOWN","cumFeeDetail":{"USDT":"0.00457722"}}]}

//{
//	"topic": "order.linear",
//	"id": "447392952_SOLUSDT_197165523953",
//	"creationTime": 1744688571933,
//	"data": [{
//		"category": "linear",
//		"symbol": "SOLUSDT",
//		"orderId": "fb31640a-0147-4064-9eb5-359b302583c0",
//		"orderLinkId": "1234567890123456",
//		"blockTradeId": "",
//		"side": "Buy",
//		"positionIdx": 1,
//		"orderStatus": "Filled",
//		"cancelType": "UNKNOWN",
//		"rejectReason": "EC_NoError",
//		"timeInForce": "GTC",
//		"isLeverage": "",
//		"price": "131",
//		"qty": "1",
//		"avgPrice": "130.32",
//		"leavesQty": "0",
//		"leavesValue": "0",
//		"cumExecQty": "1",
//		"cumExecValue": "130.32",
//		"cumExecFee": "0.052128",
//		"orderType": "Limit",
//		"stopOrderType": "",
//		"orderIv": "",
//		"triggerPrice": "",
//		"takeProfit": "",
//		"stopLoss": "",
//		"triggerBy": "",
//		"tpTriggerBy": "",
//		"slTriggerBy": "",
//		"triggerDirection": 0,
//		"placeType": "",
//		"lastPriceOnCreated": "130.32",
//		"closeOnTrigger": false,
//		"reduceOnly": false,
//		"smpGroup": 0,
//		"smpType": "None",
//		"smpOrderId": "",
//		"slLimitPrice": "0",
//		"tpLimitPrice": "0",
//		"tpslMode": "UNKNOWN",
//		"createType": "CreateByUser",
//		"marketUnit": "",
//		"createdTime": "1744688543358",
//		"updatedTime": "1744688571931",
//		"feeCurrency": "",
//		"closedPnl": "0",
//		"slippageTolerance": "0",
//		"slippageToleranceType": "UNKNOWN"
//	}]
//}
