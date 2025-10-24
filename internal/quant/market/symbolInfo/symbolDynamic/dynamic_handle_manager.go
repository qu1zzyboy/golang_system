package symbolDynamic

import (
	"context"

	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/resource/registerHandler"
	"upbitBnServer/pkg/singleton"
	"upbitBnServer/server/serverInstanceEnum"
)

type DynamicSymbolHandler func(symbolKeyId uint64, dynamicSymbol DynamicSymbol)

var (
	handleSingleton = singleton.NewSingleton(func() *Handle {
		return &Handle{
			handlers: registerHandler.NewRegistry[DynamicSymbolHandler](),
		}
	})
)

func GetHandle() *Handle {
	return handleSingleton.Get()
}

type Handle struct {
	handlers *registerHandler.Registry[DynamicSymbolHandler] //事件处理器注册中心
}

func (m *Handle) Register(ctx context.Context, instanceId serverInstanceEnum.Type, fields map[string]string, handler DynamicSymbolHandler) error {
	fields[defineJson.From] = "DynamicSymbolHandle"
	return m.handlers.RegisterOrReplace(ctx, instanceId, fields, handler)
}
