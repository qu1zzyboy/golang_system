package bybitQueryOrder

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/pkg/utils/byteUtils"
)

// len146

func GetByBitFu_C(symbolName string, reqId, cId systemx.WsId16B) []byte {
	buf := make([]byte, 0, 256)
	buf = append(buf, `{"reqId":"C`...)
	buf = append(buf, reqId[1:]...)
	buf = append(buf, `","header":{"X-BAPI-TIMESTAMP":"`...)
	buf = append(buf, reqId[:13]...)

	buf = append(buf, `"},"op":"order.cancel","args":[{"symbol":"`...)
	buf = append(buf, symbolName...)

	buf = append(buf, `","orderLinkId":"`...)
	buf = append(buf, cId[:]...)
	buf = append(buf, `","category":"linear"}]}`...)
	return buf
}

func GetByBitFu_M(symbolName string, pVal, qVal uint64, pScale systemx.PScale, qScale systemx.QScale, reqId, cId systemx.WsId16B) []byte {
	buf := make([]byte, 0, 256)
	buf = append(buf, `{"reqId":"M`...)
	buf = append(buf, reqId[1:]...)
	buf = append(buf, `","header":{"X-BAPI-TIMESTAMP":"`...)
	buf = append(buf, reqId[:13]...)

	buf = append(buf, `"},"op":"order.amend","args":[{"symbol":"`...)
	buf = append(buf, symbolName...)

	if qVal > 0 {
		// 修改後的訂單數量. 若不修改，請不要傳該字段
		buf = append(buf, `","qty":"`...)
		buf = byteUtils.AppendScaledValue(buf, qVal, int(qScale))
	}
	if pVal > 0 {
		// 修改後的訂單價格. 若不修改，請不要傳該字段
		buf = append(buf, `","price":"`...)
		buf = byteUtils.AppendScaledValue(buf, pVal, int(pScale))
	}
	buf = append(buf, `","orderLinkId":"`...)
	buf = append(buf, cId[:]...)
	buf = append(buf, `","category":"linear"}]}`...)
	return buf
}
