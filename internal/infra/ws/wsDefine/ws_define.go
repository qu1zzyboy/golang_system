package wsDefine

import (
	"context"
	"time"
	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
	"upbitBnServer/internal/resource/resourceEnum"
)

const (
	KeepAliveInterval = 15 * time.Second //保持活动连接的间隔
)

var (
	ConnErr = errorx.Newf(errCode.CodeWsDoError, "ws连接失败")
)

type ReadAutoHandler func(msg []byte)
type ReadPoolHandler func(len uint16, bufPtr *[]byte)
type PingFunc func(*SafeWrite) error

type SubDial interface {
	DialTo(ctx context.Context) (*SafeWrite, error)
}

type SubParamSdk interface {
	AddSub(ctx context.Context, params []string) error
	ListSub(ctx context.Context) error
	UnSub(ctx context.Context, params []string) error
	DialToMarket(ctx context.Context, params []string) (*SafeWrite, error)
	ParamBuild(resourceType resourceEnum.ResourceType) (func(symbol string) string, error)
}

type ReConnType uint8

const (
	READ_ERROR ReConnType = iota
	START_CONN
)

func (s ReConnType) String() string {
	switch s {
	case READ_ERROR:
		return "READ_ERROR"
	case START_CONN:
		return "START_CONN"
	default:
		return "ERROR"
	}
}
