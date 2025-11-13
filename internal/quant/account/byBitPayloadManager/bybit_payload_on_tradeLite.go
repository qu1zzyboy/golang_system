package byBitPayloadManager

import (
	"fmt"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/systemx/instanceEnum"
	"upbitBnServer/internal/quant/execute/order/orderStatic"
	"upbitBnServer/internal/strategy/toUpbitList/bybit/toUpbitByBitPayloadParse"
	"upbitBnServer/pkg/utils/byteUtils"
)

func (s *Payload) onTradeLite(data []byte) {
	fmt.Println(string(data))
	totalLen := uint16(len(data))
	var clientOrderId systemx.WsId16B
	var sy_start uint16 = 101
	sy_end := byteUtils.FindNextQuoteIndex(data, sy_start, totalLen)

	execId_start := sy_end + 12
	execId_end := byteUtils.FindNextQuoteIndex(data, execId_start, totalLen)

	p_start := execId_end + 15
	p_end := byteUtils.FindNextQuoteIndex(data, p_start, totalLen)

	q_start := p_end + 13
	q_end := byteUtils.FindNextQuoteIndex(data, q_start, totalLen)

	id_start := q_end + 13
	id_end := byteUtils.FindNextQuoteIndex(data, id_start, totalLen)

	m_start := id_end + 12
	var cid_start uint16
	switch data[m_start] {
	case 'f':
		cid_start = m_start + 21
	case 't':
		cid_start = m_start + 20
	default:
		dynamicLog.Error.GetLog().Errorf("TRADE_LITE: [%s] buy maker error %s", clientOrderId, string(data))
	}
	cid_end := cid_start + systemx.ArrLen
	copy(clientOrderId[:], data[cid_start:cid_end])
	meta, ok := orderStatic.GetService().GetOrderMeta(clientOrderId)
	if !ok {
		dynamicLog.Error.GetLog().Errorf("TRADE_LITE: [%s] orderFrom not found %s", clientOrderId, string(data))
		return
	}

	switch meta.ReqFrom {
	case instanceEnum.TO_UPBIT_LIST_BYBIT:
		toUpbitByBitPayloadParse.OnTradeLite(data, clientOrderId, meta, p_start, p_end, s.accountKeyId)
	case instanceEnum.TEST:
	default:
		dynamicLog.Error.GetLog().Errorf("TRADE_LITE: unknown ReqFrom %v", meta.ReqFrom)
	}
}

//{
//	"topic": "execution.fast.linear",
//	"creationTime": 1744688571930, //消息數據創建時間
//	"data": [{
//		"category": "linear",
//		"symbol": "SOLUSDT",
//		"execId": "8049d446-d657-59e2-aa03-a1f9f4d2c017",
//		"execPrice": "130.32",
//		"execQty": "1",
//		"orderId": "fb31640a-0147-4064-9eb5-359b302583c0",
//		"isMaker": false,
//		"orderLinkId": "",
//		"side": "Sell",
//		"execTime": "1744688571928",  //成交時間
//		"seq": 197165523953
//	}]
//}
