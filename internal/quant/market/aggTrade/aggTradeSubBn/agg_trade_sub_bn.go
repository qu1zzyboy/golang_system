package aggTradeSubBn

import (
	"context"

	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/internal/resource/wsMarketClient"
	"upbitBnServer/internal/resource/wsSub"
	"upbitBnServer/internal/resource/wsSubParam"
)

type BnAggTrade struct {
	paramMan *wsSubParam.ParamMarket   // ws订阅参数管理器
	wsClient *wsMarketClient.Market    // WebSocket 客户端
	resource resourceEnum.ResourceType // 资源类型
	exType   exchangeEnum.ExchangeType // 交易所类型
}

func newBnAggTrade() *BnAggTrade {
	return &BnAggTrade{
		paramMan: wsSubParam.NewParamMarket(wsSub.NewBnMarket()),
		resource: resourceEnum.AGG_TRADE,
		exType:   exchangeEnum.BINANCE,
	}
}

func (s *BnAggTrade) RegisterReadHandler(ctx context.Context, initSymbols []string, read wsDefine.ReadMarketHandler) error {
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

func (s *BnAggTrade) AddParamAnd(ctx context.Context, symbolName string) error {
	return s.paramMan.AddParamAnd(ctx, symbolName)
}

func (s *BnAggTrade) RemoveParamAnd(ctx context.Context, symbolName string) error {
	return s.paramMan.RemoveParamAnd(ctx, symbolName)
}

func (s *BnAggTrade) OpenSub(ctx context.Context) {
	s.wsClient.StartConn(ctx)
}

func (s *BnAggTrade) CloseSub(ctx context.Context) {
	s.wsClient.CloseSub(ctx)
}
