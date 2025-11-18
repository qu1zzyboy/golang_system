package bnPoolMarketSub

import (
	"context"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/infra/ws/client/wsMarketClient"
	"upbitBnServer/internal/infra/ws/market/wsDialMarketImpl"
	"upbitBnServer/internal/infra/ws/wsSdkImpl"
	"upbitBnServer/internal/quant/exchanges/binance/marketSub/bnPoolMarketChan"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/pkg/container/pool/byteBufPool"

	"github.com/tidwall/gjson"
)

type BnSymbolSub256 struct {
	paramMan    *wsDialMarketImpl.StaticSymbol // ws订阅参数管理器
	wsClient    *wsMarketClient.PoolMarket     // WebSocket 客户端
	symbolIndex systemx.SymbolIndex16I         //
	resource    resourceEnum.ResourceType      // 资源类型
	exType      exchangeEnum.ExchangeType      // 交易所类型
}

func newBnSymbolSub256(symbolIndex systemx.SymbolIndex16I) *BnSymbolSub256 {
	return &BnSymbolSub256{
		paramMan:    wsDialMarketImpl.NewStaticSymbol(wsSdkImpl.NewBnMarket()),
		symbolIndex: symbolIndex,
		resource:    resourceEnum.SYMBOL_SUB_256,
		exType:      exchangeEnum.BINANCE,
	}
}

func (s *BnSymbolSub256) RegisterReadHandler(ctx context.Context, symbolName string) error {
	var err error
	if err = s.paramMan.SetInitSymbols(resourceArr, symbolName); err != nil {
		return err
	}
	s.wsClient, err = wsMarketClient.NewPoolMarket(s.exType, s.resource, s.onSymbolPool, s.paramMan)
	if err != nil {
		return err
	}
	return s.wsClient.StartConn(ctx)
}

func (s *BnSymbolSub256) onSymbolPool(lenData uint16, bufPtr *[]byte) {
	b := (*bufPtr)[:lenData]
	if lenData < 60 {
		// {"result":null,"id":1982675118278578176}
		if !gjson.GetBytes(b, "id").Exists() {
			dynamicLog.Error.GetLog().Errorf("err json: %s", string(b))
		}
		byteBufPool.ReleaseBuffer(bufPtr)
		return
	}
	bnPoolMarketChan.SendPoolMarket(s.symbolIndex, bufPtr, lenData)
}

func (s *BnSymbolSub256) OpenSub(ctx context.Context) error {
	return s.wsClient.StartConn(ctx)
}

func (s *BnSymbolSub256) CloseSub(ctx context.Context) {
	if s.wsClient != nil {
		s.wsClient.CloseSub(ctx)
	}
}
