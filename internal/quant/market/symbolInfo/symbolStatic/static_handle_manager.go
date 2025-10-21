package symbolStatic

import (
	"context"

	"github.com/hhh500/quantGoInfra/define/defineJson"
	"github.com/hhh500/quantGoInfra/pkg/singleton"
	"github.com/hhh500/upbitBnServer/internal/resource/registerHandler"
	"github.com/hhh500/upbitBnServer/server/serverInstanceEnum"
)

type SymbolListDelta interface {
	OnSymbolList(ctx context.Context, s StaticTrade) error
	OnSymbolDel(ctx context.Context, s StaticTrade) error
}

var (
	handleSingleton = singleton.NewSingleton(func() *Handle {
		return &Handle{
			handlers: registerHandler.NewRegistry[SymbolListDelta](),
		}
	})
)

func GetHandle() *Handle {
	return handleSingleton.Get()
}

type Handle struct {
	handlers *registerHandler.Registry[SymbolListDelta] //事件处理器注册中心
}

func (m *Handle) Register(ctx context.Context, instanceId serverInstanceEnum.Type, fields map[string]string, handler SymbolListDelta) error {
	fields[defineJson.From] = "StaticSymbolHandle"
	return m.handlers.RegisterOrReplace(ctx, instanceId, fields, handler)
}

func (m *Handle) OnSymbolList(ctx context.Context, static StaticTrade) {
	m.handlers.Range(func(s serverInstanceEnum.Type, delta SymbolListDelta) bool {
		delta.OnSymbolList(ctx, static)
		return true
	})
}

func (m *Handle) OnSymbolDel(ctx context.Context, static StaticTrade) {
	m.handlers.Range(func(s serverInstanceEnum.Type, delta SymbolListDelta) bool {
		delta.OnSymbolDel(ctx, static)
		return true
	})
}
