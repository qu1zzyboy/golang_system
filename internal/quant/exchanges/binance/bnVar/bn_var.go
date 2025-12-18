package bnVar

import (
	"upbitBnServer/pkg/container/map/myMap"
)

var (
	Symbol2Index = myMap.NewMySyncMap[string, int]() // symbolName --> symbolIndex
)

// GetOrStoreNoTrade 给不需要交易的服务调用
func GetOrStoreNoTrade(symbolName string) int {
	symbolIndex, ok := Symbol2Index.Load(symbolName)
	if ok {
		return symbolIndex
	}
	symbolIndex = int(Symbol2Index.Length())
	Symbol2Index.Store(symbolName, symbolIndex)
	return symbolIndex
}
