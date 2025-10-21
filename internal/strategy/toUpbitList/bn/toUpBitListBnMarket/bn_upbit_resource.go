package toUpBitListBnMarket

import (
	"context"

	"github.com/hhh500/quantGoInfra/pkg/singleton"
	"github.com/hhh500/upbitBnServer/internal/quant/market/aggTrade/aggTradeSubBn"
	"github.com/hhh500/upbitBnServer/internal/quant/market/bookTick/bookTickSubBn"
	"github.com/hhh500/upbitBnServer/internal/quant/market/markPrice/markPriceSubBn"
)

type Market struct {
}

const (
	total      = "BINANCE_TOTAL"
	jsonSymbol = "s"
	jsonEvent  = "E"
)

var serviceSingleton = singleton.NewSingleton(func() *Market {
	return &Market{}
})

func GetMarket() *Market {
	return serviceSingleton.Get()
}

func (s *Market) RegisterBefore(ctx context.Context, symbols []string) error {
	//初始化各个行情数据引擎
	if err := bookTickSubBn.GetManager().RegisterReadHandler(ctx, symbols, s.OnBookTickPool); err != nil {
		return err
	}
	if err := aggTradeSubBn.GetManager().RegisterReadHandler(ctx, symbols, s.OnAggTradePool); err != nil {
		return err
	}
	if err := markPriceSubBn.GetManager().RegisterReadHandler(ctx, symbols, s.OnMarkPricePool); err != nil {
		return err
	}
	return nil
}
