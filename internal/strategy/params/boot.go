package params

import (
	"context"

	"github.com/hhh500/quantGoInfra/infra/redisx/redisConfig"
)

type Boot struct{}

func NewBoot() *Boot {
	return &Boot{}
}

func (b *Boot) ModuleId() string {
	return ModuleId
}

func (b *Boot) DependsOn() []string {
	return []string{
		redisConfig.MODULE_ID,
	}
}

func (b *Boot) Start(ctx context.Context) error {
	return GetService().Start(ctx)
}

func (b *Boot) Stop(ctx context.Context) error {
	return GetService().Stop(ctx)
}
