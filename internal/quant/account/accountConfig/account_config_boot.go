package accountConfig

import (
	"context"
)

const (
	MODULE_ID = "account_config" //全局定时任务
)

type Boot struct {
}

func NewBoot(trades, monitors []Config) *Boot {
	Trades = trades
	Monitors = monitors
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
