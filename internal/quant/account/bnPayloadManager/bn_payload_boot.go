package bnPayloadManager

import (
	"context"

	"github.com/hhh500/quantGoInfra/infra/observe/notify"
	"github.com/hhh500/upbitBnServer/internal/quant/account/accountConfig"
	"github.com/hhh500/upbitBnServer/internal/strategy/toUpbitList/toUpBitListDataAfter"
)

const MODULE_ID = "bn_payload_manager"

type Boot struct {
}

func NewBoot(fnSuccess_ toUpBitListDataAfter.OnSuccessOrder) *Boot {
	fnSuccess = fnSuccess_
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
	if err := GetManager().init(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Boot) Stop(ctx context.Context) error {
	return nil
}
