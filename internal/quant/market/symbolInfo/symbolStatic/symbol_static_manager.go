package symbolStatic

import (
	"github.com/hhh500/quantGoInfra/define/defineJson"
	"github.com/hhh500/quantGoInfra/infra/debugx"
	"github.com/hhh500/quantGoInfra/infra/errorx"
	"github.com/hhh500/quantGoInfra/infra/errorx/errCode"
	"github.com/hhh500/quantGoInfra/pkg/container/map/myMap"
	"github.com/hhh500/quantGoInfra/pkg/singleton"
)

type TradeManager struct {
	statics myMap.MySyncMap[uint64, StaticTrade] //symbolKeyId-->StaticTrade
}

var (
	serviceSingleton = singleton.NewSingleton(func() *TradeManager {
		return &TradeManager{
			statics: myMap.NewMySyncMap[uint64, StaticTrade](),
		}
	})
)

func GetTrade() *TradeManager {
	return serviceSingleton.Get()
}

// Set 由[marketRefresh]调用rest sdk更新
func (m *TradeManager) Set(s StaticTrade) error {
	// 参数校验
	if err := errorx.ValidateWithWrap(s); err != nil {
		return err
	}
	m.statics.Store(s.SymbolKeyId, s) // 更新或添加
	return nil
}

func (m *TradeManager) Delete(symbolKeyId uint64) {
	m.statics.Delete(symbolKeyId)
}

func (m *TradeManager) Get(symbolKeyId uint64) (StaticTrade, error) {
	if v, ok := m.statics.Load(symbolKeyId); ok {
		return v, nil
	}
	return StaticTrade{}, errorx.Newf(errCode.STATIC_SYMBOL_NOT_FOUND, "[%v]未找到", symbolKeyId).WithMetadata(map[string]string{
		defineJson.SymbolKey: GetSymbol().GetSymbolKey(symbolKeyId),
		"func_caller":        debugx.GetCaller(3),
	})
}

func (m *TradeManager) GetLength() int {
	return m.statics.Length()
}
