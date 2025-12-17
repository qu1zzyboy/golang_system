package toUpbitBnMode

import (
	"context"
)

const MODULE_ID = "to_upbit_bn_mode"

type Boot struct{}

func NewBoot(mode ModeBehavior) *Boot {
	Mode = mode
	return &Boot{}
}

func (b *Boot) ModuleId() string {
	return MODULE_ID
}

func (b *Boot) DependsOn() []string {
	return []string{}
}

func (b *Boot) Start(ctx context.Context) error {
	return nil
}

func (b *Boot) Stop(ctx context.Context) error {
	return nil
}
