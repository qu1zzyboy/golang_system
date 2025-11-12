package sellOpenMarket

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/pkg/utils/byteUtils"
)

func GetBnFu_NoSign_u32(symbolName string, qVal uint64, qScale systemx.QScale, cId systemx.WsId16B) []byte {
	buf := make([]byte, 0, 256)
	buf = append(buf, `{"id":"`...)
	buf = append(buf, cId[:]...)
	buf = append(buf, `","method":"order.place","params":{"newClientOrderId":"`...)
	buf = append(buf, cId[:]...)
	buf = append(buf, `","positionSide":"SHORT","quantity":"`...)
	buf = byteUtils.AppendScaledValue(buf, qVal, int(qScale))
	buf = append(buf, `","side":"SELL","symbol":"`...)
	buf = append(buf, symbolName...)
	buf = append(buf, `","timestamp":"`...)
	buf = append(buf, cId[:13]...)
	buf = append(buf, `","type":"MARKET"}}`...)
	return buf
}

func GetByBitFu_NoSign_u32(symbolName string, qVal uint64, qScale systemx.QScale, cId systemx.WsId16B) []byte {
	buf := make([]byte, 0, 512)
	buf = append(buf, `{"reqId":"`...)
	buf = append(buf, cId[:]...)
	buf = append(buf, `","header":{"X-BAPI-TIMESTAMP":"`...)
	buf = append(buf, cId[:13]...)

	buf = append(buf, `"},"op":"order.create","args":[{"symbol":"`...)
	buf = append(buf, symbolName...)

	buf = append(buf, `","side":"Sell","orderType":"Market","positionIdx":2,"qty":"`...)
	buf = byteUtils.AppendScaledValue(buf, qVal, int(qScale))

	buf = append(buf, `","orderLinkId":"`...)
	buf = append(buf, cId[:]...)
	buf = append(buf, `","category":"linear"}]}`...)
	return buf
}
