package safex

import (
	"context"

	"upbitBnServer/internal/infra/observe/notify"
)

const (
	MODULE_ID = "safex"
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
		notify.MODULE_ID,
	}
}

func (s *Boot) Start(ctx context.Context) error {
	return nil
}

func (s *Boot) Stop(ctx context.Context) error {
	return nil
}
