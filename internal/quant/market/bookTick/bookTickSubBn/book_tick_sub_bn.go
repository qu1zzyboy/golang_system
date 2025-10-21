package bookTickSubBn

import (
	"context"

	"github.com/hhh500/quantGoInfra/infra/ws/wsDefine"
	"github.com/hhh500/quantGoInfra/quant/exchanges/exchangeEnum"
	"github.com/hhh500/quantGoInfra/resource/resourceEnum"
	"github.com/hhh500/upbitBnServer/internal/resource/wsMarketClient"
	"github.com/hhh500/upbitBnServer/internal/resource/wsSub"
	"github.com/hhh500/upbitBnServer/internal/resource/wsSubParam"
)

type BnBootTick struct {
	paramMan *wsSubParam.ParamMarket   // ws订阅参数管理器
	wsClient *wsMarketClient.Market    // WebSocket 客户端
	resource resourceEnum.ResourceType // 资源类型
	exType   exchangeEnum.ExchangeType // 交易所类型
}

func newBnBootTick() *BnBootTick {
	return &BnBootTick{
		paramMan: wsSubParam.NewParamMarket(wsSub.NewBnMarket()),
		resource: resourceEnum.BOOK_TICK,
		exType:   exchangeEnum.BINANCE,
	}
}

func (s *BnBootTick) RegisterReadHandler(ctx context.Context, initSymbols []string, read wsDefine.ReadMarketHandler) error {
	var err error
	if err = s.paramMan.SetInitSymbols(s.resource, initSymbols); err != nil {
		return err
	}
	s.wsClient, err = wsMarketClient.NewMarket(s.exType, s.resource, read, s.paramMan)
	if err != nil {
		return err
	}
	return s.wsClient.StartConn(ctx)
}

func (s *BnBootTick) AddParamAnd(ctx context.Context, symbolName string) error {
	return s.paramMan.AddParamAnd(ctx, symbolName)
}

func (s *BnBootTick) RemoveParamAnd(ctx context.Context, symbolName string) error {
	return s.paramMan.RemoveParamAnd(ctx, symbolName)
}

func (s *BnBootTick) OpenSub(ctx context.Context) {
	s.wsClient.StartConn(ctx)
}

func (s *BnBootTick) CloseSub(ctx context.Context) {
	s.wsClient.CloseSub(ctx)
}
