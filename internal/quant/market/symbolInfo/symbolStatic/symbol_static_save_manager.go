package symbolStatic

import (
	"github.com/hhh500/quantGoInfra/define/defineJson"
	"github.com/hhh500/quantGoInfra/infra/debugx"
	"github.com/hhh500/quantGoInfra/infra/errorx"
	"github.com/hhh500/quantGoInfra/infra/errorx/errCode"
	"github.com/hhh500/quantGoInfra/pkg/container/map/myMap"
	"github.com/hhh500/quantGoInfra/pkg/singleton"
	"github.com/hhh500/quantGoInfra/quant/exchanges/exchangeEnum"
)

const SYMBOL_INFO_STATIC_MANAGER = "symbol_info_static_manager"

type SaveManager struct {
	statics myMap.MySyncMap[uint64, StaticSave] //symbolKeyId-->StaticSave
}

var (
	saveSingleton = singleton.NewSingleton(func() *SaveManager {
		return &SaveManager{
			statics: myMap.NewMySyncMap[uint64, StaticSave](),
		}
	})
)

func GetSave() *SaveManager {
	return saveSingleton.Get()
}

func (m *SaveManager) GetAllKeyMapBn() map[uint64]struct{} {
	resp := make(map[uint64]struct{})
	m.statics.Range(func(symbolKeyId uint64, v StaticSave) bool {
		if v.ExType == exchangeEnum.BINANCE {
			resp[symbolKeyId] = struct{}{}
		}
		return true
	})
	return resp
}

func (m *SaveManager) GetAllKeyMapByBit() map[uint64]struct{} {
	resp := make(map[uint64]struct{})
	m.statics.Range(func(k uint64, v StaticSave) bool {
		if v.ExType == exchangeEnum.BYBIT {
			resp[k] = struct{}{}
		}
		return true
	})
	return resp
}

// Set 由[marketRefresh]调用rest sdk更新
func (m *SaveManager) Set(s StaticSave) error {
	// 参数校验
	if err := errorx.ValidateWithWrap(s); err != nil {
		return err
	}
	m.statics.Store(s.SymbolKeyId, s) // 更新或添加
	return nil
}

func (m *SaveManager) Delete(symbolKeyId uint64) {
	m.statics.Delete(symbolKeyId)
}

func (m *SaveManager) Get(symbolKeyId uint64) (StaticSave, error) {
	if v, ok := m.statics.Load(symbolKeyId); ok {
		return v, nil
	}
	return StaticSave{}, errorx.Newf(errCode.STATIC_SYMBOL_NOT_FOUND, "[%v]未找到", symbolKeyId).WithMetadata(map[string]string{
		defineJson.SymbolKey: GetSymbol().GetSymbolKey(symbolKeyId),
		"func_caller":        debugx.GetCaller(3),
	})
}

func (m *SaveManager) GetLength() int {
	return m.statics.Length()
}
