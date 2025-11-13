package byBitPayloadManager

import (
	"fmt"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/usageEnum"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/orderStatic"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitPointPreBn"
	"upbitBnServer/internal/strategy/toUpbitList/bybit/toUpbitBybitSymbol"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"
	"upbitBnServer/pkg/utils/byteUtils"
	"upbitBnServer/pkg/utils/convertx/byteConvert"
)

func (s *Payload) onPayloadOrder(data []byte) {

	fmt.Println(string(data))

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

	// switch meta.ReqFrom {
	// case instanceEnum.TO_UPBIT_LIST_BN:
	// 	toUpbitByBitPayloadParse.OnPayloadOrder(data, clientOrderId, meta, totalLen, cidEnd, s.accountKeyId)
	// case instanceEnum.TEST:
	// default:
	// 	dynamicLog.Error.GetLog().Errorf("TRADE_LITE: unknown ReqFrom %v", meta.ReqFrom)
	// }

	switch meta.UsageFrom {
	case usageEnum.TO_UPBIT_PRE:
		{
			// 卖出开仓成交,主要是用来驱动策略触发
			if meta.OrderMode != execute.SELL_OPEN_LIMIT {
				return
			}
			var o_start uint16
			switch data[cid_end+28] {
			case 'B':
				o_start = cid_end + 64
			case 'S':
				o_start = cid_end + 65
			default:
				toUpBitDataStatic.DyLog.GetLog().Errorf("B_S 解析异常:%s", string(data))
				return
			}
			hasFilled := false
			switch data[o_start] {
			case 'N':
				toUpbitBybitSymbol.OnOrderUpdate(true, clientOrderId)
			case 'P':
				hasFilled = true
			case 'F':
				hasFilled = true
			case 'C':
				toUpbitBybitSymbol.OnOrderUpdate(false, clientOrderId)
			case 'R':
				toUpbitBybitSymbol.OnOrderUpdate(false, clientOrderId)
			default:
				toUpBitDataStatic.DyLog.GetLog().Errorf("[%d]ORDER_UPDATE: unknown order status[%d], json: %s", s.accountKeyId, o_start, string(data))
				return
			}
			// 有成交才继续处理
			if !hasFilled {
				return
			}
			// 该笔订单已经被处理过了
			if _, ok = toUpbitPointPreBn.ClientOrderNotOpen.Load(clientOrderId); ok {
				return
			}
			toUpbitPointPreBn.ClientOrderNotOpen.Store(clientOrderId, struct{}{})
			toUpbitListChan.SendTradeLite(meta.SymbolIndex, toUpbitListChan.TrigOrderInfo{
				ClientOrderId: clientOrderId,
				//T:             convertx.BytesToInt64(data[cid_end+28 : cid_end+41]),
				//P:             convertx.PriceByteArrToUint64(data[p_start:p_end], 8),
			})
			toUpBitDataStatic.SigLog.GetLog().Infof("[%d]触发前成交,%s", s.accountKeyId, string(data))
		}
	case usageEnum.TO_UPBIT_MAIN:
		{
			// 1、只管触发标的的订单
			if meta.SymbolIndex != toUpBitListDataAfter.TrigSymbolIndex {
				toUpBitDataStatic.DyLog.GetLog().Errorf("触发后异常订单返回:%s", string(data))
				return
			}
			var o_start uint16
			switch data[cid_end+28] {
			case 'B':
				o_start = cid_end + 64
			case 'S':
				o_start = cid_end + 65
			default:
				toUpBitDataStatic.DyLog.GetLog().Errorf("B_S 解析异常:%s", string(data))
				return
			}
			var orderStatus execute.OrderStatus
			var X_len uint16
			switch data[o_start] {
			case 'N':
				orderStatus = execute.NEW
				X_len = 3
			case 'P':
				orderStatus = execute.PARTIALLY_FILLED
				X_len = 15
			case 'F':
				orderStatus = execute.FILLED
				X_len = 6
			case 'C':
				orderStatus = execute.CANCELED
				X_len = 9
			case 'R':
				orderStatus = execute.REJECTED
				X_len = 8
			default:
				toUpBitDataStatic.DyLog.GetLog().Errorf("[%d]ORDER_UPDATE: unknown order status[%d], json: %s", s.accountKeyId, o_start, string(data))
				return
			}
			isOnline := execute.IsOrderOnLine(orderStatus)
			evt := toUpBitListDataAfter.OnSuccessEvt{
				ClientOrderId: clientOrderId,
				IsOnline:      isOnline,
				OrderMode:     meta.OrderMode,
				AccountKeyId:  s.accountKeyId,
			}
			if !isOnline {
				o_end := o_start + X_len
				p_start := o_end + 98
				p_end := byteUtils.FindNextQuoteIndex(data, p_start, totalLen)

				q_start := p_end + 9
				q_end := byteUtils.FindNextQuoteIndex(data, q_start, totalLen)

				avg_p_start := q_end + 14
				avg_p_end := byteUtils.FindNextQuoteIndex(data, avg_p_start, totalLen)

				// 有成交
				if avg_p_end-avg_p_start > 1 {
					left_q_start := avg_p_end + 15
					left_q_end := byteUtils.FindNextQuoteIndex(data, left_q_start, totalLen)
					left_q_u8 := byteConvert.PriceByteArrToUint64(data[left_q_start:left_q_end], 8)
					if left_q_u8 == 0 {
						// evt.Volume = meta.OrigVolume
					} else {
						// evt.Volume = meta.OrigVolume.Sub(decimal.New(int64(left_q_u8), -8))
					}
				}
			} else {
				// evt.TimeStamp = data.Get("createdTime").Int()
			}
			toUpbitListChan.SendSuOrder(meta.SymbolIndex, evt)
		}
	default:
		dynamicLog.Error.GetLog().Errorf("ORDER_UPDATE: unknown orderFrom %v", meta.UsageFrom)
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
