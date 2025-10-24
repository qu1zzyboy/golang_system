package treenews

import (
	"context"

	"upbitBnServer/internal/conf"
)

const MODULE_ID = "treenews_service"

type Boot struct{}

func NewBoot() *Boot {
	return &Boot{}
}

func (b *Boot) ModuleId() string {
	return MODULE_ID
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
