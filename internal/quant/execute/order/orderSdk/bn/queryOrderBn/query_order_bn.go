package queryOrderBn

import (
	"upbitBnServer/internal/infra/systemx"
)

// len146
// {"id":"C762412447010163","method":"order.cancel","params":{"origClientOrderId":"1762410636008457","symbol":"ETHUSDT","timestamp":"1762412447010"}}

func GetBnFu_NoSign_Q(symbolName string, reqId, cId systemx.WsId16B) []byte {
	buf := make([]byte, 0, 256)
	buf = append(buf, `{"id":"Q`...)
	buf = append(buf, reqId[1:]...)
	buf = append(buf, `","method":"order.status","params":{"origClientOrderId":"`...)
	buf = append(buf, cId[:]...)
	buf = append(buf, `","symbol":"`...)
	buf = append(buf, symbolName...)
	buf = append(buf, `","timestamp":"`...)
	buf = append(buf, reqId[:13]...)
	buf = append(buf, `"}}`...)
	return buf
}

func GetBnFu_NoSign_C(symbolName string, reqId, cId systemx.WsId16B) []byte {
	buf := make([]byte, 0, 256)
	buf = append(buf, `{"id":"C`...)
	buf = append(buf, reqId[1:]...)
	buf = append(buf, `","method":"order.cancel","params":{"origClientOrderId":"`...)
	buf = append(buf, cId[:]...)
	buf = append(buf, `","symbol":"`...)
	buf = append(buf, symbolName...)
	buf = append(buf, `","timestamp":"`...)
	buf = append(buf, reqId[:13]...)
	buf = append(buf, `"}}`...)
	return buf
}
