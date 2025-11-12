package toUpBitByBit

import (
	"context"
	"time"
	"upbitBnServer/internal/infra/systemx"
	"upbitBnServer/internal/quant/exchanges/bybit/autoMarketBybitSub"
	"upbitBnServer/internal/quant/exchanges/bybit/bybitVar"
	"upbitBnServer/internal/quant/exchanges/bybit/poolMarketBybitSub"
	"upbitBnServer/internal/quant/market/symbolInfo/coinMesh"
	"upbitBnServer/internal/strategy/toUpbitList/bybit/toUpbitBybitSymbolArr"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/server/instance"
)

func (e *Engine) OnSymbolList(ctx context.Context, s *coinMesh.CoinMesh) error {
	toUpBitDataStatic.DyLog.GetLog().Infof("upbit上币:%v", s)
	symbolName := s.BybitFuUsdtName
	if symbolName == "" {
		return nil
	}
	var symbolIndex systemx.SymbolIndex16I
	// 下标不存在则添加,要不然会出现订阅的品种没有索引
	if rawIndex, ok := bybitVar.SymbolIndex.Load(symbolName); !ok {
		symbolIndex = systemx.SymbolIndex16I(bybitVar.SymbolIndex.Length())
		if err := toUpbitBybitSymbolArr.GetSymbolObj(symbolIndex).Start(e.getPreAccountKeyId(), bybitVar.SymbolIndex.Length(), symbolName); err != nil {
			return err
		}
	} else {
		symbolIndex = rawIndex
	}
	// 再订阅
	if err := poolMarketBybitSub.GetSymbolObj(symbolIndex).RegisterReadHandler(ctx, symbolName); err != nil {
		toUpBitDataStatic.DyLog.GetLog().Errorf("bn_upbit上币失败,err:%v", err)
		return err
	}
	if err := autoMarketBybitSub.GetSymbolObj(symbolIndex).RegisterReadHandler(ctx, symbolName); err != nil {
		toUpBitDataStatic.DyLog.GetLog().Errorf("bn_upbit上币失败,err:%v", err)
		return err
	}
	return nil
}

func (e *Engine) OnSymbolDel(ctx context.Context, s *coinMesh.CoinMesh) error {
	toUpBitDataStatic.DyLog.GetLog().Infof("upbit下币:%v", s)
	if s.BybitFuUsdtName == "" {
		return nil
	}
	if rawIndex, ok := bybitVar.SymbolIndex.Load(s.BybitFuUsdtName); ok {
		poolMarketBybitSub.GetSymbolObj(rawIndex).CloseSub(ctx)
		autoMarketBybitSub.GetSymbolObj(rawIndex).CloseSub(ctx)
	}
	return nil
}

func (e *Engine) OnStop(ctx context.Context) error {
	bybitVar.SymbolIndex.Range(func(symbolName string, symbolIndex systemx.SymbolIndex16I) bool {
		poolMarketBybitSub.GetSymbolObj(symbolIndex).CloseSub(ctx)
		autoMarketBybitSub.GetSymbolObj(symbolIndex).CloseSub(ctx)
		return true
	})
	time.Sleep(3 * time.Second)                                          //等待数据处理完
	toUpbitBybitSymbolArr.CancelAllOrders(bybitVar.SymbolIndex.Length()) //撤销所有挂单
	return nil
}

func (e *Engine) OnUpdate(ctx context.Context, param instance.InstanceUpdate) error {
	return nil
}

func (e *Engine) OnReceive() error {
	return nil
}

func (e *Engine) OnDayBegin() {
	today := time.Now()
	weekday := today.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		toUpBitDataStatic.DyLog.GetLog().Info("今天是周末,不交易")
		return
	} else {
		ctx := context.Background()
		stopIndex := bybitVar.SymbolIndex.Length()
		poolMarketBybitSub.OpenSub(ctx, stopIndex)
		autoMarketBybitSub.OpenSub(ctx, stopIndex)
		toUpBitDataStatic.DyLog.GetLog().Info("今天交易开始,打开订阅")
	}
}

func (e *Engine) OnDayEnd() {
	ctx := context.Background()
	stopIndex := bybitVar.SymbolIndex.Length()
	poolMarketBybitSub.CloseSub(ctx, stopIndex)      //关闭订阅
	autoMarketBybitSub.CloseSub(ctx, stopIndex)      //关闭订阅
	time.Sleep(3 * time.Second)                      //等待数据处理完
	toUpbitBybitSymbolArr.CancelAllOrders(stopIndex) //撤销所有挂单
	toUpbitBybitSymbolArr.ClearByDayEnd(stopIndex)   //清理涨幅相关数据
	toUpBitDataStatic.DyLog.GetLog().Info("今天交易结束,关闭订阅")
}
