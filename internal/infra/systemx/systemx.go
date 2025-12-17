package systemx

import "github.com/shopspring/decimal"

const (
	ArrLen uint16 = 16
)

type SymbolIndex16I int16         //交易对索引,[-32768,32767]
type WsId16B string               //reqId和clientOrderId的类型
type OrderSdkType decimal.Decimal //订单的价格和数量的类型
type PScale int32                 //价格精度类型
type QScale int32                 //数量精度类型
func (p PScale) Uint8() int32     { return int32(p) }
func (q QScale) Uint8() int32     { return int32(q) }

type Job struct {
	Buf *[]byte
	Len uint16
}
