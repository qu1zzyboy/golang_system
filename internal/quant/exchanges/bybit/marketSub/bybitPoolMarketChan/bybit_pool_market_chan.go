package bybitPoolMarketChan

import "upbitBnServer/internal/infra/systemx"

var (
	chanPoolMarketArr []chan systemx.Job
)

func InitChanArr(size int) {
	chanPoolMarketArr = make([]chan systemx.Job, size)
}

func Register(symbolIndex systemx.SymbolIndex16I, poolMarketChan chan systemx.Job) {
	chanPoolMarketArr[symbolIndex] = poolMarketChan
}

func SendPoolMarket(symbolIndex systemx.SymbolIndex16I, buf *[]byte, len uint16) {
	chanPoolMarketArr[symbolIndex] <- systemx.Job{Buf: buf, Len: len}
}
