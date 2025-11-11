package bnPayloadManager

import (
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"upbitBnServer/internal/quant/execute/order/orderStatic"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitListBnSymbol"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataStatic"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

/*
一定是订单有效信息的返回
*/
func (s *Payload) onPayloadOrder(data []byte) {
	// 拿到clientOrderId去查内存静态数据
	clientOrderId := gjson.GetBytes(data, "o.c").String()
	orderFrom, orderMode, symbolIndex, ok := orderStatic.GetService().GetOrderInstanceIdAndSymbolId(clientOrderId)

	if !ok {
		// 可能是手动平仓单
		if clientOrderId[0:3] == "ios" || clientOrderId[0:3] == "web" || clientOrderId[0:3] == "ele" {
			return
		}
		dynamicLog.Error.GetLog().Errorf("[%d]订单推送: [%s] orderFrom not found %s", s.accountKeyId, clientOrderId, string(data))
		return
	}

	switch orderFrom {
	case orderBelongEnum.TO_UPBIT_LIST_PRE:
		{
			// 卖出开仓成交,主要是用来驱动策略触发
			if orderMode != execute.ORDER_SELL_OPEN {
				return
			}
			// 订单状态异常
			orderStatus := execute.ParseBnOrderStatus(gjson.GetBytes(data, "o.X").String())
			if orderStatus == execute.UNKNOWN_ORDER_STATUS {
				toUpBitListDataStatic.DyLog.GetLog().Errorf("[%d]ORDER_UPDATE: unknown order status, json: %s", s.accountKeyId, string(data))
				return
			}
			switch orderStatus {
			case execute.NEW:
				toUpbitListBnSymbol.OnOrderUpdate(true, clientOrderId)
			case execute.CANCELED:
				toUpbitListBnSymbol.OnOrderUpdate(false, clientOrderId)
			default:

			}
			// 只管成交的订单
			if orderStatus != execute.PARTIALLY_FILLED && orderStatus != execute.FILLED {
				return
			}
			// 该笔订单已经被处理过了
			if _, ok = toUpbitListBnSymbol.ClientOrderIsCheck.Load(clientOrderId); ok {
				return
			}
			toUpbitListBnSymbol.ClientOrderIsCheck.Store(clientOrderId, struct{}{})
			toUpbitListChan.SendDeltaOrder(symbolIndex, data)
		}
	case orderBelongEnum.TO_UPBIT_LIST_LOOP:
		{
			// 1、只管触发标的的订单
			if symbolIndex != toUpBitListDataAfter.TrigSymbolIndex {
				toUpBitListDataStatic.DyLog.GetLog().Errorf("触发后异常订单:%s", string(data))
				return
			}
			// 订单状态异常
			orderStatus := execute.ParseBnOrderStatus(gjson.GetBytes(data, "o.X").String())
			if orderStatus == execute.UNKNOWN_ORDER_STATUS {
				toUpBitListDataStatic.DyLog.GetLog().Errorf("[%d]ORDER_UPDATE: unknown order status, json: %s", s.accountKeyId, string(data))
				return
			}

			isOnline := execute.IsOrderOnLine(orderStatus)
			evt := toUpBitListDataAfter.OnSuccessEvt{
				ClientOrderId: clientOrderId,
				IsOnline:      isOnline,
				OrderMode:     orderMode,
				InstanceId:    orderFrom,
				AccountKeyId:  s.accountKeyId,
			}
			if !isOnline {
				evt.Volume = decimal.RequireFromString(gjson.GetBytes(data, "o.z").String())
			} else {
				evt.TimeStamp = gjson.GetBytes(data, "T").Int()
			}
			toUpbitListChan.SendSuOrder(symbolIndex, evt)
		}
	default:
		dynamicLog.Error.GetLog().Errorf("ORDER_UPDATE: unknown orderFrom %v", orderFrom)
	}
}
