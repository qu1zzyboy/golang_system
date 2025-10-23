package bnOrderAppManager

import (
	"context"

	"github.com/hhh500/quantGoInfra/infra/observe/notify"
	"github.com/hhh500/upbitBnServer/internal/quant/account/accountConfig"
)

const MODULE_ID = "bn_order_app_manager"

type Boot struct {
}

func NewBoot() *Boot {
	// 设置回调函数
	return &Boot{}
}

func (s *Boot) ModuleId() string {
	return MODULE_ID
}

func (s *Boot) DependsOn() []string {
	return []string{
		accountConfig.MODULE_ID, //账户配置信息
		notify.MODULE_ID,        //通知模块
	}
}

func (s *Boot) Start(ctx context.Context) error {
	if err := GetTradeManager().init(ctx); err != nil {
		return err
	}
	if err := GetMonitorManager().init(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Boot) Stop(ctx context.Context) error {
	return nil
}
