package bnVar

import (
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/pkg/container/map/myMap"
)

var (
	SymbolIndex = myMap.NewMySyncMap[string, systemx.SymbolIndex16I]() // symbolName --> symbolIndex
)

func GetOrStore(symbolName string) systemx.SymbolIndex16I {
	symbolIndex, ok := SymbolIndex.Load(symbolName)
	if !ok {
		symbolIndex = systemx.SymbolIndex16I(SymbolIndex.Length())
		SymbolIndex.Store(symbolName, symbolIndex)
	}
	return symbolIndex
}
