package bnPayloadManager

import (
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/quant/execute/order/orderStatic"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitBnPayloadParse"
	"upbitBnServer/pkg/utils/byteUtils"
)

func (s *Payload) onTradeLite(data []byte) {
	totalLen := uint16(len(data))
	var clientOrderId systemx.WsId16B
	symbolEnd := byteUtils.FindNextQuoteIndex(data, 59, totalLen)
	qStart := symbolEnd + 7
	qEnd := byteUtils.FindNextQuoteIndex(data, qStart, totalLen)
	pStart := qEnd + 7
	pEnd := byteUtils.FindNextQuoteIndex(data, pStart, totalLen)
	mStart := pEnd + 6
	cidStart := mStart + 10
	copy(clientOrderId[:], data[cidStart:cidStart+systemx.ArrLen])
	meta, ok := orderStatic.GetService().GetOrderMeta(clientOrderId)
	if !ok {
		// 可能是手动平仓单
		switch {
		case clientOrderId[0] == 'i' && clientOrderId[1] == 'o' && clientOrderId[2] == 's':
			return
		case clientOrderId[0] == 'w' && clientOrderId[1] == 'e' && clientOrderId[2] == 'b':
			return
		case clientOrderId[0] == 'e' && clientOrderId[1] == 'l' && clientOrderId[2] == 'e':
			return
		case clientOrderId[0] == 'a' && clientOrderId[1] == 'n' && clientOrderId[2] == 'd':
			return
		}
		dynamicLog.Error.GetLog().Errorf("[%d]TRADE_LITE: [%s] orderFrom not found %s", s.accountKeyId, string(clientOrderId[:]), string(data))
		return
	}
	switch meta.ReqFrom {
	case instanceEnum.TO_UPBIT_LIST_BN:
		toUpbitBnPayloadParse.OnTradeLite(data, clientOrderId, meta, pStart, pEnd, s.accountKeyId)
	case instanceEnum.TEST:
	default:
		dynamicLog.Error.GetLog().Errorf("TRADE_LITE: unknown ReqFrom %v", meta.ReqFrom)
	}
}
