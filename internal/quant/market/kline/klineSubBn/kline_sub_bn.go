package klineSubBn

import (
	"context"
	"upbitBnServer/internal/infra/ws/client/wsMarketClient"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/quant/market/kline/klineEnum"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/internal/resource/wsSub"
	"upbitBnServer/internal/resource/wsSubParam"
	"upbitBnServer/pkg/singleton"
)

var (
	bnSingleton = singleton.NewSingleton(func() *BnDeltaDepth { return newBnDeltaDepth() })
)

func GetManager() *BnDeltaDepth {
	return bnSingleton.Get()
}

type BnDeltaDepth struct {
	paramMan *wsSubParam.ParamMarket    // ws订阅参数管理器
	wsClient *wsMarketClient.PoolMarket // WebSocket 客户端
	resource resourceEnum.ResourceType  // 资源类型
	exType   exchangeEnum.ExchangeType  // 交易所类型
}

func newBnDeltaDepth() *BnDeltaDepth {
	return &BnDeltaDepth{
		paramMan: wsSubParam.NewParamMarket(wsSub.NewBnKline(klineEnum.KLINE_1h)),
		resource: resourceEnum.KLINE,
		exType:   exchangeEnum.BINANCE,
	}
}

func (s *BnDeltaDepth) RegisterReadHandler(ctx context.Context, initSymbols []string, read wsDefine.ReadPoolHandler) error {
	var err error
	if err = s.paramMan.SetInitSymbols(s.resource, initSymbols); err != nil {
		return err
	}
	s.wsClient, err = wsMarketClient.NewPoolMarket(s.exType, s.resource, read, s.paramMan)
	if err != nil {
		return err
	}
	return s.wsClient.StartConn(ctx)
}

func (s *BnDeltaDepth) AddParamAnd(ctx context.Context, symbolName string) error {
	return s.paramMan.AddParamAnd(ctx, symbolName)
}

func (s *BnDeltaDepth) RemoveParamAnd(ctx context.Context, symbolName string) error {
	return s.paramMan.RemoveParamAnd(ctx, symbolName)
}

func (s *BnDeltaDepth) OpenSub(ctx context.Context) {
	s.wsClient.StartConn(ctx)
}

func (s *BnDeltaDepth) CloseSub(ctx context.Context) {
	s.wsClient.CloseSub(ctx)
}
