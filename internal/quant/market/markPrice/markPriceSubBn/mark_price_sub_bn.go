package markPriceSubBn

import (
	"context"

	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/internal/resource/wsMarketClient"
	"upbitBnServer/internal/resource/wsSub"
	"upbitBnServer/internal/resource/wsSubParam"
)

type BnMarkPrice struct {
	paramMan *wsSubParam.ParamMarket   // ws订阅参数管理器
	wsClient *wsMarketClient.Market    // WebSocket 客户端
	resource resourceEnum.ResourceType // 资源类型
	exType   exchangeEnum.ExchangeType // 交易所类型
}

func newBnMarkPrice() *BnMarkPrice {
	return &BnMarkPrice{
		paramMan: wsSubParam.NewParamMarket(wsSub.NewBnMarket()),
		resource: resourceEnum.MARK_PRICE,
		exType:   exchangeEnum.BINANCE,
	}
}

func (s *BnMarkPrice) RegisterReadHandler(ctx context.Context, initSymbols []string, read wsDefine.ReadMarketHandler) error {
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

func (s *BnMarkPrice) AddParamAnd(ctx context.Context, symbolName string) error {
	return s.paramMan.AddParamAnd(ctx, symbolName)
}

func (s *BnMarkPrice) RemoveParamAnd(ctx context.Context, symbolName string) error {
	return s.paramMan.RemoveParamAnd(ctx, symbolName)
}

func (s *BnMarkPrice) OpenSub(ctx context.Context) {
	s.wsClient.StartConn(ctx)
}

func (s *BnMarkPrice) CloseSub(ctx context.Context) {
	s.wsClient.CloseSub(ctx)
}
