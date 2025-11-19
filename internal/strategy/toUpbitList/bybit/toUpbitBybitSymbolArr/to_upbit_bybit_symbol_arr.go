package toUpbitBybitSymbolArr

import (
	"time"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/strategy/toUpbitList/bybit/toUpbitBybitSymbol"
)

var (
	symObjArray []*toUpbitBybitSymbol.Single // 交易对信息
)

func InitObjArr(size int) {
	symObjArray = make([]*toUpbitBybitSymbol.Single, size)
	for i := range size {
		symObjArray[i] = &toUpbitBybitSymbol.Single{}
	}
}

func GetSymbolObj(symbolIndex systemx.SymbolIndex16I) *toUpbitBybitSymbol.Single {
	return symObjArray[symbolIndex]
}

func CancelAllOrders(stopIndex int) {
	for index, sym := range symObjArray {
		if index >= stopIndex {
			break
		}
		sym.CancelPreOrder()
	}
}

func ClearByDayEnd(stopIndex int) {
	for index, sym := range symObjArray {
		if index >= stopIndex {
			break
		}
		sym.Clear()
	}
}

func RefreshByDayBegin(stopIndex int) {
	for index, sym := range symObjArray {
		if index >= stopIndex {
			break
		}
		sym.ClearBegin()
		time.Sleep(100 * time.Millisecond)
	}
}
