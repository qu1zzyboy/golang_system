package symbolDynamic

import (
	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/debugx"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
	"upbitBnServer/pkg/container/map/myMap"
	"upbitBnServer/pkg/singleton"
)

const SYMBOL_INFO_DYNAMIC_MANAGER = "symbol_info_dynamic_manager"

var (
	serviceSingleton = singleton.NewSingleton(func() *Manager {
		return &Manager{
			dynamics: myMap.NewMySyncMap[uint64, DynamicSymbol](),
		}
	})
)

func GetManager() *Manager {
	return serviceSingleton.Get()
}

type Manager struct {
	dynamics myMap.MySyncMap[uint64, DynamicSymbol] //symbolKeyId -> DynamicSymbol
}

func (m *Manager) Delete(symbolKeyId uint64) {
	m.dynamics.Delete(symbolKeyId)
}

func (m *Manager) SetDirect(symbolKeyId uint64, new DynamicSymbol) error {
	if err := new.Check(); err != nil {
		return err.WithMetadata(map[string]string{
			"reqType":            "DynamicSymbol",
			defineJson.SymbolKey: symbolStatic.GetSymbol().GetSymbolKey(symbolKeyId),
		})
	}
	m.dynamics.Store(symbolKeyId, new)
	return nil
}

func (m *Manager) Get(symbolKeyId uint64) (DynamicSymbol, error) {
	if v, ok := m.dynamics.Load(symbolKeyId); ok {
		return v, nil
	}
	return DynamicSymbol{}, ErrDynamicNotFound.WithMetadata(map[string]string{
		defineJson.SymbolKey: symbolStatic.GetSymbol().GetSymbolKey(symbolKeyId),
		"func_caller":        debugx.GetCaller(3),
	})
}

func (m *Manager) Exists(symbolKeyId uint64) bool {
	_, ok := m.dynamics.Load(symbolKeyId)
	return ok
}

func (m *Manager) GetLength() int {
	return m.dynamics.Length()
}
