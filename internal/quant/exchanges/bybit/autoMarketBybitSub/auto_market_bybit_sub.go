package autoMarketBybitSub

import (
	"context"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/ws/client/wsMarketClient"
	"upbitBnServer/internal/infra/ws/market/wsDialMarketImpl"
	"upbitBnServer/internal/infra/ws/wsSdkImpl"
	"upbitBnServer/internal/quant/exchanges/bybit/autoMarketChanByBit"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/resource/resourceEnum"
)

type ByBitSymbolSubAuto struct {
	paramMan    *wsDialMarketImpl.StaticSymbol // ws订阅参数管理器
	wsClient    *wsMarketClient.AutoMarket     // WebSocket 客户端
	symbolIndex systemx.SymbolIndex16I         //
	resource    resourceEnum.ResourceType      // 资源类型
	exType      exchangeEnum.ExchangeType      // 交易所类型
}

func newByBitSymbolSubAuto(symbolIndex systemx.SymbolIndex16I) *ByBitSymbolSubAuto {
	return &ByBitSymbolSubAuto{
		paramMan:    wsDialMarketImpl.NewStaticSymbol(wsSdkImpl.NewByBitMarket()),
		symbolIndex: symbolIndex,
		resource:    resourceEnum.SYMBOL_SUB_AUTO,
		exType:      exchangeEnum.BYBIT,
	}
}

func (s *ByBitSymbolSubAuto) RegisterReadHandler(ctx context.Context, symbolName string) error {
	var err error
	if err = s.paramMan.SetInitSymbols(resourceArr, symbolName); err != nil {
		return err
	}
	s.wsClient, err = wsMarketClient.NewAutoMarket(s.exType, s.resource, s.onSymbolPool, s.paramMan)
	if err != nil {
		return err
	}
	return s.wsClient.StartConn(ctx)
}

func (s *ByBitSymbolSubAuto) onSymbolPool(data []byte) {
	if data[2] == 's' && data[3] == 'u' {
		return
	}
	autoMarketChanByBit.SendAutoMarket(s.symbolIndex, data)
}

func (s *ByBitSymbolSubAuto) OpenSub(ctx context.Context) error {
	return s.wsClient.StartConn(ctx)
}

func (s *ByBitSymbolSubAuto) CloseSub(ctx context.Context) {
	if s.wsClient != nil {
		s.wsClient.CloseSub(ctx)
	}
}
