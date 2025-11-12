package buyOpenLimit

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/pkg/utils/byteUtils"
)

// GetBnFu_NoSign_u32
// 193.8 ns/op	     512 B/op	       1 allocs/op
func GetBnFu_NoSign_u32(symbolName string, pVal, qVal uint64, pScale systemx.PScale, qScale systemx.QScale, cId systemx.WsId16B) []byte {
	buf := make([]byte, 0, 512)
	buf = append(buf, `{"id":"`...)
	buf = append(buf, cId[:]...)
	buf = append(buf, `","method":"order.place","params":{"newClientOrderId":"`...)
	buf = append(buf, cId[:]...)
	buf = append(buf, `","positionSide":"LONG","price":"`...)
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

// {
//     "reqId": "test-005",
//     "header": {
//         "X-BAPI-TIMESTAMP": "1711001595207",
//         "X-BAPI-RECV-WINDOW": "8000",
//         "Referer": "bot-001" // for api broker
//     },
//     "op": "order.create",
//     "args": [
//         {
//             "symbol": "ETHUSDT",
//             "side": "Buy",
//             "orderType": "Limit",
//             "qty": "0.2",
//             "price": "2800",
//             "category": "linear",
//             "timeInForce": "PostOnly"
//         }
//     ]
// }

func GetByBitFu_NoSign_u32(symbolName string, pVal, qVal uint64, pScale systemx.PScale, qScale systemx.QScale, cId systemx.WsId16B) []byte {
	buf := make([]byte, 0, 512)
	buf = append(buf, `{"reqId":"`...)
	buf = append(buf, cId[:]...)
	buf = append(buf, `","header":{"X-BAPI-TIMESTAMP":"`...)
	buf = append(buf, cId[:13]...)

	buf = append(buf, `"},"op":"order.create","args":[{"symbol":"`...)
	buf = append(buf, symbolName...)

	buf = append(buf, `","side":"Buy","orderType":"Limit","positionIdx":1,"qty":"`...)
	buf = byteUtils.AppendScaledValue(buf, qVal, int(qScale))

	buf = append(buf, `","price":"`...)
	buf = byteUtils.AppendScaledValue(buf, pVal, int(pScale))

	buf = append(buf, `","orderLinkId":"`...)
	buf = append(buf, cId[:]...)
	buf = append(buf, `","category":"linear"}]}`...)
	return buf
}

// bytes.Buffer pass
// var []byte  pass
