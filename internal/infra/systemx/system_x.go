package systemx

const (
	ArrLen uint16 = 16
)

type PosCal64U uint64     // 仓位计算,[0,1.8446744 × 10^19]
type SymbolIndex16I int16 //交易对索引,[-32768,32767]
type WsId16B [ArrLen]byte //reqId和clientOrderId的类型

type Job struct {
	Buf *[]byte
	Len uint16
}
