package modifySellLimit

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/pkg/utils/byteUtils"
)

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
	buf = append(buf, `","side":"SELL","symbol":"`...)
	buf = append(buf, symbolName...)
	buf = append(buf, `","timestamp":"`...)
	buf = append(buf, reqId[:13]...)
	buf = append(buf, `"}}`...)
	return buf
}
