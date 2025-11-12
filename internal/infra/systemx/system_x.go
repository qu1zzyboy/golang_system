package systemx

const (
	ArrLen uint16 = 16
)

type SymbolIndex16I int16 //交易对索引,[-32768,32767]
type WsId16B [ArrLen]byte //reqId和clientOrderId的类型

type ScaledValue uint32 //定点整数值

type PScale uint8 //价格精度类型
type QScale uint8 //数量精度类型

func (p PScale) Uint8() uint8        { return uint8(p) }
func (q QScale) Uint8() uint8        { return uint8(q) }
func (v ScaledValue) Uint32() uint32 { return uint32(v) }

type Job struct {
	Buf *[]byte
	Len uint16
}
