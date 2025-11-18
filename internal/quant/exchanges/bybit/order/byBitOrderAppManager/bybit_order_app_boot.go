package byBitOrderAppManager

import (
	"context"
	"upbitBnServer/internal/infra/observe/notify"
	"upbitBnServer/internal/quant/account/accountConfig"
)

const MODULE_ID = "bybit_order_app_manager"

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
		accountConfig.MODULE_ID, //账户配置信息
		notify.MODULE_ID,        //通知模块
	}
}

func (s *Boot) Start(ctx context.Context) error {
	if err := GetTradeManager().init(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Boot) Stop(ctx context.Context) error {
	return nil
}
