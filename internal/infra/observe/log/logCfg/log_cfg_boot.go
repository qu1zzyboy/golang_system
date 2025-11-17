package logCfg

import (
	"context"
)

const (
	MODULE_ID = "log_cfg" //全局定时任务
)

type Boot struct {
}

func NewBoot(logLevel LogLevel) *Boot {
	G_LOG_LEVEL = logLevel
	return &Boot{}
}

func (s *Boot) ModuleId() string {
	return MODULE_ID
}

func (s *Boot) DependsOn() []string {
	return []string{}
}

func (s *Boot) Start(ctx context.Context) error {
	return nil
}

func (s *Boot) Stop(ctx context.Context) error {
	return nil
}
