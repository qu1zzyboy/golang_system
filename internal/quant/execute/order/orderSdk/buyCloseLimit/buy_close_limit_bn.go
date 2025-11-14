package buyCloseLimit

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/pkg/utils/byteUtils"
)

func GetBnFu_NoSign_u32(symbolName string, pVal, qVal uint64, pScale systemx.PScale, qScale systemx.QScale, cId systemx.WsId16B) []byte {
	buf := make([]byte, 0, 512)
	buf = append(buf, `{"id":"`...)
	buf = append(buf, cId[:]...)
	buf = append(buf, `","method":"order.place","params":{"newClientOrderId":"`...)
	buf = append(buf, cId[:]...)
	buf = append(buf, `","positionSide":"SHORT","price":"`...)
	buf = byteUtils.AppendScaledValue(buf, pVal, int(pScale))
	buf = append(buf, `","quantity":"`...)
	buf = byteUtils.AppendScaledValue(buf, qVal, int(qScale))
	buf = append(buf, `","side":"BUY","symbol":"`...)
	buf = append(buf, symbolName...)
	buf = append(buf, `","timeInForce":"GTC","timestamp":"`...)
	buf = append(buf, cId[:13]...)
	buf = append(buf, `","type":"LIMIT"}}`...)
	return buf
}

func GetByBitFu_NoSign_u32(symbolName string, pVal, qVal uint64, pScale systemx.PScale, qScale systemx.QScale, cId systemx.WsId16B) []byte {
	buf := make([]byte, 0, 512)
	buf = append(buf, `{"reqId":"`...)
	buf = append(buf, cId[:]...)
	buf = append(buf, `","header":{"X-BAPI-TIMESTAMP":"`...)
	buf = append(buf, cId[:13]...)

	buf = append(buf, `"},"op":"order.create","args":[{"symbol":"`...)
	buf = append(buf, symbolName...)

	buf = append(buf, `","side":"Buy","orderType":"Limit","positionIdx":2,"qty":"`...)
	buf = byteUtils.AppendScaledValue(buf, qVal, int(qScale))

	buf = append(buf, `","price":"`...)
	buf = byteUtils.AppendScaledValue(buf, pVal, int(pScale))

	buf = append(buf, `","orderLinkId":"`...)
	buf = append(buf, cId[:]...)
	buf = append(buf, `","category":"linear"}]}`...)
	return buf
}
