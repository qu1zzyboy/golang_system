package bybitAccountAvailable

import (
	"context"
)

const MODULE_ID = "bybit_account_available"

type Boot struct {
}

func NewBoot(length int) *Boot {
	GetManager().init(length)
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
