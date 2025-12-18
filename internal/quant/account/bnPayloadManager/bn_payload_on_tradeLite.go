package bnPayloadManager

import (
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/quant/exchanges/binance/bnVar"
	"upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"upbitBnServer/internal/quant/execute/order/orderStatic"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitListBnSymbol"
	"upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"

	"github.com/tidwall/gjson"
)

func (s *Payload) processTradeLitePrePoint(data []byte, clientOrderId string, symbolIndex int) {
	// 该笔订单已经被处理过了
	if _, ok := toUpbitListBnSymbol.ClientOrderIsCheck.Load(clientOrderId); ok {
		return
	}
	toUpbitListBnSymbol.ClientOrderIsCheck.Store(clientOrderId, struct{}{})
	toUpbitListChan.SendTradeLite(symbolIndex, data)
}

func (s *Payload) onTradeLite(data []byte) {
	clientOrderId := gjson.GetBytes(data, "c").String()
	orderFrom, _, symbolIndex, ok := orderStatic.GetService().GetOrderInstanceIdAndSymbolId(clientOrderId)
	if !ok {
		// 可能是手动平仓单
		if clientOrderId[0:3] == "ios" || clientOrderId[0:3] == "web" || clientOrderId[0:3] == "ele" {
			return
		}
		if clientOrderId[0:5] == "point" {
			symbolIndex = bnVar.GetOrStoreNoTrade(gjson.GetBytes(data, "s").String())
			s.processTradeLitePrePoint(data, clientOrderId, symbolIndex)
			return
		}
		if clientOrderId[0:6] == "server" {
			return
		}
		dynamicLog.Error.GetLog().Errorf("[%d]TRADE_LITE: [%s] orderFrom not found %s", s.accountKeyId, clientOrderId, string(data))
		return
	}
	switch orderFrom {
	case orderBelongEnum.TO_UPBIT_LIST_PRE:
		{
			s.processTradeLitePrePoint(data, clientOrderId, symbolIndex)
		}
	case orderBelongEnum.TO_UPBIT_LIST_LOOP:
		{
			//暂时不处理
		}
	default:
		dynamicLog.Error.GetLog().Errorf("TRADE_LITE: unknown orderFrom %v", orderFrom)
	}
}
