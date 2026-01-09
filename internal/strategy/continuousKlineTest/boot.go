package continuousKlineTest

import (
	"context"

	"upbitBnServer/internal/infra/redisx/redisConfig"
)

type Boot struct{}

func NewBoot() *Boot {
	return &Boot{}
}

func (b *Boot) ModuleId() string {
	return MODULE_ID
}

func (b *Boot) DependsOn() []string {
	return []string{
		redisConfig.MODULE_ID, // Redis 配置（用于存储 K 线数据）
	}
}

func (b *Boot) Start(ctx context.Context) error {
	return GetService().Start(ctx)
}

func (b *Boot) Stop(ctx context.Context) error {
	return GetService().Stop(ctx)
}
