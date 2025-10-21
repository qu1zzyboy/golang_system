package bnPayloadManager

import (
	"github.com/hhh500/quantGoInfra/infra/observe/log/dynamicLog"
	"github.com/hhh500/upbitBnServer/internal/quant/execute"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderBelongEnum"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderStatic"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitListBnSymbol"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpbitListChan"
	"github.com/tidwall/gjson"
)

func (s *Payload) onTradeLite(data []byte) {
	clientOrderId := gjson.GetBytes(data, "c").String()
	instanceId, orderMode, symbolIndex, ok := orderStatic.GetService().GetOrderInstanceIdAndSymbolId(clientOrderId)
	if !ok {
		// 可能是手动平仓单
		if gjson.GetBytes(data, "s").String() == toUpBitListDataAfter.TrigSymbolName {
			return
		}
		dynamicLog.Error.GetLog().Errorf("[%d]TRADE_LITE: [%s] orderInstance not found %s", s.accountKeyId, clientOrderId, string(data))
		return
	}
	switch instanceId {
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
		dynamicLog.Error.GetLog().Errorf("TRADE_LITE: unknown instanceId %v", instanceId)
	}
}
