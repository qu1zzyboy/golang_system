package treenews

import (
	"context"

	"github.com/hhh500/quantGoInfra/conf"
)

type Boot struct{}

func NewBoot() *Boot {
	return &Boot{}
}

func (b *Boot) ModuleId() string {
	return "treenews_service"
}

func (b *Boot) DependsOn() []string {
	return []string{conf.MODULE_ID}
}

func (b *Boot) Start(ctx context.Context) error {
	return GetService().Start(ctx)
}

func (b *Boot) Stop(ctx context.Context) error {
	return GetService().Stop(ctx)
}
