package redisConfig

import (
	"context"

	"github.com/hhh500/quantGoInfra/conf"
	"github.com/hhh500/quantGoInfra/infra/redisx"
)

const (
	MODULE_ID = "redis_config"
)

type Boot struct {
}

func NewBoot() *Boot {
	return &Boot{}
}

func (s *Boot) ModuleId() string {
	return MODULE_ID
}

func (s *Boot) DependsOn() []string {
	return []string{
		conf.MODULE_ID, //配置加载
	}
}

func (s *Boot) Start(ctx context.Context) error {
	if err := redisx.RegisterClient(ctx, conf.RedisCfg.Hosts, conf.RedisCfg.Pass, CONFIG_ALL_KEY, 0); err != nil {
		return err
	}
	return nil
}

func (s *Boot) Stop(ctx context.Context) error {
	return nil
}
