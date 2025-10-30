package toUpbitListBnSymbolArr

import (
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitListBnSymbol"
)

var (
	symObjArray []*toUpbitListBnSymbol.Single // 交易对信息
)

func Init(size int) {
	symObjArray = make([]*toUpbitListBnSymbol.Single, size)
	for i := range size {
		symObjArray[i] = &toUpbitListBnSymbol.Single{}
	}
}

func GetSymbolObj(symbolIndex int) *toUpbitListBnSymbol.Single {
	return symObjArray[symbolIndex]
}

func CancelAllOrders() {
	for _, sym := range symObjArray {
		sym.CancelPreOrder()
	}
}

func ClearByDayEnd() {
	for _, sym := range symObjArray {
		sym.Clear()
	}
}
