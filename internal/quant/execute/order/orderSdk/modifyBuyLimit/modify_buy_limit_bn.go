package modifyBuyLimit

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/pkg/utils/byteUtils"
)

// len195
// {"id":"M762412446009244","method":"order.modify","params":{"origClientOrderId":"1762410636008457","price":"3200.00","quantity":"0.01","side":"BUY","symbol":"ETHUSDT","timestamp":"1762412446009"}}

func GetBnFu_NoSign_u32(symbolName string, pVal, qVal uint64, pScale systemx.PScale, qScale systemx.QScale, reqId, cId systemx.WsId16B) []byte {
	buf := make([]byte, 0, 256)
	buf = append(buf, `{"id":"M`...)
	buf = append(buf, reqId[1:]...)
	buf = append(buf, `","method":"order.modify","params":{"origClientOrderId":"`...)
	buf = append(buf, cId[:]...)
	buf = append(buf, `","price":"`...)
	buf = byteUtils.AppendScaledValue(buf, pVal, int(pScale))
	buf = append(buf, `","quantity":"`...)
	buf = byteUtils.AppendScaledValue(buf, qVal, int(qScale))
	buf = append(buf, `","side":"BUY","symbol":"`...)
	buf = append(buf, symbolName...)
	buf = append(buf, `","timestamp":"`...)
	buf = append(buf, reqId[:13]...)
	buf = append(buf, `"}}`...)
	return buf
}
