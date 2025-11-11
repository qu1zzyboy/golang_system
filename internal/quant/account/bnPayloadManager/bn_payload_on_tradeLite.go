package bnPayloadManager

import (
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"upbitBnServer/internal/quant/execute/order/orderStatic"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitListBnSymbol"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"

	"github.com/tidwall/gjson"
)

func (s *Payload) onTradeLite(data []byte) {
	clientOrderId := gjson.GetBytes(data, "c").String()
	orderFrom, orderMode, symbolIndex, ok := orderStatic.GetService().GetOrderInstanceIdAndSymbolId(clientOrderId)
	if !ok {
		// 可能是手动平仓单
		if clientOrderId[0:3] == "ios" || clientOrderId[0:3] == "web" || clientOrderId[0:3] == "ele" {
			return
		}
		dynamicLog.Error.GetLog().Errorf("[%d]TRADE_LITE: [%s] orderFrom not found %s", s.accountKeyId, clientOrderId, string(data))
		return
	}
	switch orderFrom {
	case orderBelongEnum.TO_UPBIT_LIST_PRE:
		{
			// 卖出开仓成交,主要是用来驱动策略触发
			if orderMode != execute.ORDER_SELL_OPEN {
				return
			}
			// 该笔订单已经被处理过了
			if _, ok = toUpbitListBnSymbol.ClientOrderIsCheck.Load(clientOrderId); ok {
				return
			}
			toUpbitListBnSymbol.ClientOrderIsCheck.Store(clientOrderId, struct{}{})
			toUpbitListChan.SendTradeLite(symbolIndex, data)
		}
	case orderBelongEnum.TO_UPBIT_LIST_LOOP:
		{
			//暂时不处理
		}
	default:
		dynamicLog.Error.GetLog().Errorf("TRADE_LITE: unknown orderFrom %v", orderFrom)
	}
}
