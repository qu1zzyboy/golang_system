package wsSub

import (
	"context"

	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/resource/resourceEnum"
)

type WsSub interface {
	AddSub(ctx context.Context, params []string) error
	ListSub(ctx context.Context) error
	UnSub(ctx context.Context, params []string) error
	DialToMarket(ctx context.Context, params []string) (*wsDefine.SafeWrite, error)
	ParamBuild(resourceType resourceEnum.ResourceType) (func(symbol string) string, error)
}

type SubParam interface {
	DialTo(ctx context.Context) (*wsDefine.SafeWrite, error)
}
