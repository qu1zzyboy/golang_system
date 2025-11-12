package autoMarketChanByBit

import "upbitBnServer/internal/infra/systemx"

var (
	chanPoolMarketArr []chan []byte
)

func InitChanArr(size int) {
	chanPoolMarketArr = make([]chan []byte, size)
}

func Register(symbolIndex systemx.SymbolIndex16I, poolMarketChan chan []byte) {
	chanPoolMarketArr[symbolIndex] = poolMarketChan
}

func SendAutoMarket(symbolIndex systemx.SymbolIndex16I, msg []byte) {
	chanPoolMarketArr[symbolIndex] <- msg
}
