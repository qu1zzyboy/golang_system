package toUpbitMesh

import (
	"context"

	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/quant/market/symbolInfo/coinMesh"
	"upbitBnServer/internal/resource/registerHandler"
	"upbitBnServer/pkg/singleton"
	"upbitBnServer/server/serverInstanceEnum"
)

type UpBitListDelta interface {
	OnSymbolList(ctx context.Context, s *coinMesh.CoinMesh) error
	OnSymbolDel(ctx context.Context, s *coinMesh.CoinMesh) error
}

var (
	handleSingleton = singleton.NewSingleton(func() *Handle {
		return &Handle{
			handlers: registerHandler.NewRegistry[UpBitListDelta](),
		}
	})
)

func GetHandle() *Handle {
	return handleSingleton.Get()
}

type Handle struct {
	handlers *registerHandler.Registry[UpBitListDelta] //事件处理器注册中心
}

func (m *Handle) Register(ctx context.Context, instanceId serverInstanceEnum.Type, fields map[string]string, handler UpBitListDelta) error {
	fields[defineJson.From] = "to_UpBit_MeshHandle"
	return m.handlers.RegisterOrReplace(ctx, instanceId, fields, handler)
}

func (m *Handle) UnRegister(ctx context.Context, instanceId serverInstanceEnum.Type, fields map[string]string) error {
	return m.handlers.Unregister(ctx, instanceId, fields)
}

func (m *Handle) OnSymbolList(ctx context.Context, mesh *coinMesh.CoinMesh) {
	m.handlers.Range(func(s serverInstanceEnum.Type, delta UpBitListDelta) bool {
		delta.OnSymbolList(ctx, mesh)
		return true
	})
}
func (m *Handle) OnSymbolDel(ctx context.Context, mesh *coinMesh.CoinMesh) {
	m.handlers.Range(func(s serverInstanceEnum.Type, delta UpBitListDelta) bool {
		delta.OnSymbolDel(ctx, mesh)
		return true
	})
}
