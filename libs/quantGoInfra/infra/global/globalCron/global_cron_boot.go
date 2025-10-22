package globalCron

import (
	"context"
	"sync"

	"github.com/robfig/cron/v3"
)

const (
	MODULE_ID = "global_cron" //全局定时任务
)

var (
	once sync.Once
	c    *cron.Cron
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
	return []string{}
}

func (s *Boot) Start(ctx context.Context) error {
	initDefault() // 确保 cron 启动
	return nil
}

func (s *Boot) Stop(ctx context.Context) error {
	return nil
}
