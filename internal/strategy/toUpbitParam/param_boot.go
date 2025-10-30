package toUpbitParam

import (
	"context"

	"upbitBnServer/internal/infra/global/globalCron"
	"upbitBnServer/internal/infra/redisx"
	"upbitBnServer/internal/infra/redisx/redisConfig"
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
		redisConfig.MODULE_ID, //依赖redis数据
		globalCron.MODULE_ID,  //依赖定时器
	}
}

func (b *Boot) Start(ctx context.Context) error {
	redisClient, err := redisx.LoadClient(redisConfig.CONFIG_ALL_KEY)
	if err != nil {
		return err
	}
	if err := GetService().Start(ctx, redisClient); err != nil {
		return err
	}
	return nil
}

func (b *Boot) Stop(ctx context.Context) error {
	return GetService().Stop(ctx)
}
