package grpcLowLatencyServer

import (
	"context"

	"upbitBnServer/internal/quant/market/symbolInfo/symbolInfoLoad"
	"upbitBnServer/internal/strategy/toUpbitParam"
	"upbitBnServer/internal/strategy/treenews"
	"upbitBnServer/pkg/container/pool/antPool"
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
		treenews.MODULE_ID,       // treeNews模块
		toUpbitParam.ModuleId,
	}
}

func (s *Boot) Start(ctx context.Context) error {
	return nil
}

func (s *Boot) Stop(ctx context.Context) error {
	return nil
}
