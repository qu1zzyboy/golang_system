package symbolLimit

import (
	"upbitBnServer/internal/resource/registerHandler"
	"upbitBnServer/pkg/singleton"
)

type LimitSymbolHandler func(symbolKeyId uint64, dynamicSymbol LimitSymbol)

var (
	handleSingleton = singleton.NewSingleton(func() *Handle {
		return &Handle{
			handlers: registerHandler.NewRegistry[LimitSymbolHandler](),
		}
	})
)

func GetHandle() *Handle {
	return handleSingleton.Get()
}

type Handle struct {
	handlers *registerHandler.Registry[LimitSymbolHandler] //事件处理器注册中心
}

// func (m *Handle) Register(ctx context.Context, instanceId instanceEnum.Type, fields map[string]string, handler LimitSymbolHandler) error {
// 	fields[defineJson.From] = "LimitSymbolHandle"
// 	return m.handlers.RegisterOrReplace(ctx, instanceId, fields, handler)
// }
