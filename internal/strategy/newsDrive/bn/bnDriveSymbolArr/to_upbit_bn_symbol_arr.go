package bnDriveSymbolArr

import (
	"upbitBnServer/internal/strategy/newsDrive/bn/bnDriveSymbol"
)

var (
	symObjArray []*bnDriveSymbol.Single // 交易对信息
)

func Init(size int) {
	symObjArray = make([]*bnDriveSymbol.Single, size)
	for i := range size {
		symObjArray[i] = &bnDriveSymbol.Single{}
	}
}

func GetSymbolObj(symbolIndex int) *bnDriveSymbol.Single {
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
