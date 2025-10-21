package symbolStatic

import (
	"github.com/hhh500/quantGoInfra/pkg/container/map/myMap"
	"github.com/hhh500/quantGoInfra/pkg/singleton"
)

type SymbolId struct {
	symbolToU64    myMap.MySyncMap[string, uint64] // symbolName --> symbolId
	u64ToSymbolKey myMap.MySyncMap[uint64, string] // symbolKeyId --> symbolKey
}

var (
	symbolSingleton = singleton.NewSingleton(func() *SymbolId {
		return &SymbolId{
			symbolToU64:    myMap.NewMySyncMap[string, uint64](),
			u64ToSymbolKey: myMap.NewMySyncMap[uint64, string](),
		}
	})
)

func GetSymbol() *SymbolId {
	return symbolSingleton.Get()
}

func (m *SymbolId) SetSymbol(symbolName string, symbolId uint64) {
	m.symbolToU64.Store(symbolName, symbolId)
}

func (m *SymbolId) GetSymbol(symbolName string) (uint64, bool) {
	return m.symbolToU64.Load(symbolName)
}

func (m *SymbolId) SetSymbolKey(symbolKeyId uint64, symbolKey string) {
	m.u64ToSymbolKey.Store(symbolKeyId, symbolKey)
}

func (m *SymbolId) GetSymbolKey(symbolKeyId uint64) string {
	if si, ok := m.u64ToSymbolKey.Load(symbolKeyId); ok {
		return si
	}
	return ""
}
