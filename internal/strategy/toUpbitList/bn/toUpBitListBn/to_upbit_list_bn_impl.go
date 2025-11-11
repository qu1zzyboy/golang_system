package toUpBitListBn

import (
	"context"
	"time"

	"upbitBnServer/internal/quant/market/aggTrade/aggTradeSubBn"
	"upbitBnServer/internal/quant/market/bookTick/bookTickSubBn"
	"upbitBnServer/internal/quant/market/markPrice/markPriceSubBn"
	"upbitBnServer/internal/quant/market/symbolInfo/coinMesh"
	"upbitBnServer/internal/strategy/toUpbitList/bn/toUpbitListBnSymbolArr"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/server/instance"
)

func (e *Engine) onSymbolList(ctx context.Context, s *coinMesh.CoinMesh) error {
	if err := bookTickSubBn.GetManager().AddParamAnd(ctx, s.BnFuUsdtName); err != nil {
		return err
	}
	if err := aggTradeSubBn.GetManager().AddParamAnd(ctx, s.BnFuUsdtName); err != nil {
		return err
	}
	if err := markPriceSubBn.GetManager().AddParamAnd(ctx, s.BnFuUsdtName); err != nil {
		return err
	}
	return nil
}

func (e *Engine) OnSymbolList(ctx context.Context, s *coinMesh.CoinMesh) error {
	toUpBitDataStatic.DyLog.GetLog().Infof("upbit上币:%v", s)
	symbolName := s.BnFuUsdtName
	if symbolName == "" {
		return nil
	}
	// 下标不存在则添加,要不然会出现订阅的品种没有索引
	if _, ok := toUpBitDataStatic.SymbolIndex.Load(symbolName); !ok {

		symbolIndex := toUpBitDataStatic.SymbolIndex.Length()
		e.thisCalCount++
		if e.thisCalCount == limit {
			e.thisAccountKeyId++
			e.thisCalCount = 0
		}
		if err := toUpbitListBnSymbolArr.GetSymbolObj(symbolIndex).Start(e.thisAccountKeyId, symbolIndex, symbolName); err != nil {
			return err
		}
	}
	// 再订阅
	if err := e.onSymbolList(ctx, s); err != nil {
		toUpBitDataStatic.DyLog.GetLog().Errorf("bn_upbit上币失败,err:%v", err)
		return err
	}
	return nil
}

func (e *Engine) onSymbolDel(ctx context.Context, s *coinMesh.CoinMesh) error {
	if s.BnFuUsdtName == "" {
		return nil
	}
	if err := bookTickSubBn.GetManager().RemoveParamAnd(ctx, s.BnFuUsdtName); err != nil {
		return err
	}
	if err := aggTradeSubBn.GetManager().RemoveParamAnd(ctx, s.BnFuUsdtName); err != nil {
		return err
	}
	if err := markPriceSubBn.GetManager().RemoveParamAnd(ctx, s.BnFuUsdtName); err != nil {
		return err
	}
	return nil
}

func (e *Engine) OnSymbolDel(ctx context.Context, s *coinMesh.CoinMesh) error {
	toUpBitDataStatic.DyLog.GetLog().Infof("upbit下币:%v", s)
	if err := e.onSymbolDel(ctx, s); err != nil {
		toUpBitDataStatic.DyLog.GetLog().Errorf("bn_upbit下币失败,err:%v", err)
		return err
	}
	return nil
}

func (e *Engine) OnStop(ctx context.Context) error {
	bookTickSubBn.GetManager().CloseSub(ctx)  //关闭订阅
	aggTradeSubBn.GetManager().CloseSub(ctx)  //关闭订阅
	markPriceSubBn.GetManager().CloseSub(ctx) //关闭订阅
	time.Sleep(3 * time.Second)               //等待数据处理完
	toUpbitListBnSymbolArr.CancelAllOrders()  //撤销所有挂单
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
		bookTickSubBn.GetManager().OpenSub(ctx)
		aggTradeSubBn.GetManager().OpenSub(ctx)
		markPriceSubBn.GetManager().OpenSub(ctx)
		toUpBitDataStatic.DyLog.GetLog().Info("今天交易开始,打开订阅")
	}
}

func (e *Engine) OnDayEnd() {
	ctx := context.Background()
	bookTickSubBn.GetManager().CloseSub(ctx)  //关闭订阅
	aggTradeSubBn.GetManager().CloseSub(ctx)  //关闭订阅
	markPriceSubBn.GetManager().CloseSub(ctx) //关闭订阅
	time.Sleep(3 * time.Second)               //等待数据处理完
	toUpbitListBnSymbolArr.CancelAllOrders()  //撤销所有挂单
	toUpbitListBnSymbolArr.ClearByDayEnd()    //清理涨幅相关数据
	toUpBitDataStatic.DyLog.GetLog().Info("今天交易结束,关闭订阅")
}
