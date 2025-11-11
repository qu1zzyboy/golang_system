package symbolLimit

import (
	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/debugx"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
	"upbitBnServer/pkg/container/map/myMap"
	"upbitBnServer/pkg/singleton"
)

const SYMBOL_INFO_LIMIT_MANAGER = "symbol_info_limit_manager"

var (
	serviceSingleton = singleton.NewSingleton(func() *Manager {
		return &Manager{
			limits: myMap.NewMySyncMap[uint64, LimitSymbol](),
		}
	})
)

func GetManager() *Manager {
	return serviceSingleton.Get()
}

type Manager struct {
	limits myMap.MySyncMap[uint64, LimitSymbol] //symbolKeyId -> DynamicSymbol
}

func (m *Manager) Delete(symbolKeyId uint64) {
	m.limits.Delete(symbolKeyId)
}

func (m *Manager) SetDirect(symbolKeyId uint64, new LimitSymbol) error {
	if err := new.Check(); err != nil {
		return err.WithMetadata(map[string]string{
			"reqType":            "DynamicSymbol",
			defineJson.SymbolKey: symbolStatic.GetSymbol().GetSymbolKey(symbolKeyId),
		})
	}
	m.limits.Store(symbolKeyId, new)
	return nil
}

func (m *Manager) Get(symbolKeyId uint64) (LimitSymbol, error) {
	if v, ok := m.limits.Load(symbolKeyId); ok {
		return v, nil
	}
	return LimitSymbol{}, ErrDynamicNotFound.WithMetadata(map[string]string{
		defineJson.SymbolKey: symbolStatic.GetSymbol().GetSymbolKey(symbolKeyId),
		"func_caller":        debugx.GetCaller(3),
	})
}

func (m *Manager) Exists(symbolKeyId uint64) bool {
	_, ok := m.limits.Load(symbolKeyId)
	return ok
}

func (m *Manager) GetLength() int {
	return m.limits.Length()
}
