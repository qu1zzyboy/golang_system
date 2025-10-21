package grpcLowLatencyServer

import (
	"context"

	"github.com/hhh500/quantGoInfra/pkg/container/pool/antPool"
	"github.com/hhh500/upbitBnServer/internal/quant/market/symbolInfo/symbolInfoLoad"
)

const ModuleId = "grpc_lowLatency_server"

type Boot struct {
}

func NewBoot() *Boot {
	return &Boot{}
}

func (s *Boot) ModuleId() string {
	return ModuleId
}

func (s *Boot) DependsOn() []string {
	return []string{
		symbolInfoLoad.MODULE_ID, // 从redis加载交易规范
		antPool.MODULE_ID,        // 协程池
	}
}

func (s *Boot) Start(ctx context.Context) error {
	return nil
}

func (s *Boot) Stop(ctx context.Context) error {
	return nil
}
