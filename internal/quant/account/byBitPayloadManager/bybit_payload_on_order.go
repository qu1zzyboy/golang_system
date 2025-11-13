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
	case instanceEnum.TO_UPBIT_LIST_BN:
		toUpbitByBitPayloadParse.OnPayloadOrder(data, clientOrderId, meta, totalLen, cid_end, s.accountKeyId)
	case instanceEnum.TEST:
	default:
		dynamicLog.Error.GetLog().Errorf("TRADE_LITE: unknown ReqFrom %v", meta.ReqFrom)
	}
}

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
