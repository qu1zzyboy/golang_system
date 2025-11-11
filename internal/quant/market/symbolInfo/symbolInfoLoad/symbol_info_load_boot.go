package symbolInfoLoad

import (
	"context"

	"upbitBnServer/internal/infra/observe/log/staticLog"
	"upbitBnServer/internal/infra/redisx"
	"upbitBnServer/internal/infra/redisx/redisConfig"
	"upbitBnServer/internal/quant/market/symbolInfo/coinMesh"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolDynamic"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolLimit"
	"upbitBnServer/internal/quant/market/symbolInfo/symbolStatic"
)

const MODULE_ID = "symbol_info_load"

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
		redisConfig.MODULE_ID, //redis配置信息入库
	}
}

func (s *Boot) Start(ctx context.Context) error {
	redisClient, err := redisx.LoadClient(redisConfig.CONFIG_ALL_KEY)
	if err != nil {
		return err
	}
	if err := loadStatic(ctx, redisClient); err != nil {
		return err
	}
	staticLog.Log.Info("静态交易对加载完成,共计:", symbolStatic.GetTrade().GetLength())
	if err := loadDynamic(ctx, redisClient); err != nil {
		return err
	}
	staticLog.Log.Info("动态交易对加载完成,共计:", symbolDynamic.GetManager().GetLength())
	if err = loadLimit(ctx, redisClient); err != nil {
		return err
	}
	staticLog.Log.Info("限制交易对加载完成,共计:", symbolLimit.GetManager().GetLength())
	if err := loadCoinMesh(ctx, redisClient); err != nil {
		return err
	}
	staticLog.Log.Info("币种信息加载完成,共计:", coinMesh.GetManager().GetLength())
	return nil
}

func (s *Boot) Stop(ctx context.Context) error {
	return nil
}
