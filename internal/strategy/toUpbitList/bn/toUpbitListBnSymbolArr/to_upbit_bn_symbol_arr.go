package toUpbitListBnSymbolArr

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitListBnSymbol"
)

var (
	symObjArray []*toUpbitListBnSymbol.Single // 交易对信息
)

func InitObjArr(size int) {
	symObjArray = make([]*toUpbitListBnSymbol.Single, size)
	for i := range size {
		symObjArray[i] = &toUpbitListBnSymbol.Single{}
	}
}

func GetSymbolObj(symbolIndex systemx.SymbolIndex16I) *toUpbitListBnSymbol.Single {
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
