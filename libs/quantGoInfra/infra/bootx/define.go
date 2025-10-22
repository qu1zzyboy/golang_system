package bootx

import (
	"context"

	"github.com/hhh500/quantGoInfra/pkg/singleton"
)

type Bootable interface {
	ModuleId() string    // 唯一标识
	DependsOn() []string // 启动前依赖的组件名
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

var (
	manager = singleton.NewSingleton(func() *BootManager {
		return &BootManager{components: make(map[string]Bootable)}
	})
)

func GetManager() *BootManager {
	return manager.Get()
}
