package klineSubBn

import (
	"context"
	"time"

	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/quant/market/kline/klineEnum"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/internal/resource/wsMarketClient"
	"upbitBnServer/internal/resource/wsSub"
	"upbitBnServer/internal/resource/wsSubParam"
	"upbitBnServer/pkg/container/map/myMap"
	"upbitBnServer/pkg/singleton"
)

var (
	continousKlineSingleton = singleton.NewSingleton(func() *Manager { return newManager() })
)

func GetContinuousKlineManager() *Manager {
	return continousKlineSingleton.Get()
}

// Manager 管理多个 BnContinuousKline 实例，支持分批订阅
type Manager struct {
	subMapping map[string]uint8                           // symbolName-->订阅序号
	subs       myMap.MySyncMap[uint8, *BnContinuousKline] // 订阅序号-->订阅服务
	subIndex   uint8                                      // 当前订阅实例索引
}

func newManager() *Manager {
	return &Manager{
		subMapping: make(map[string]uint8),
		subs:       myMap.NewMySyncMap[uint8, *BnContinuousKline](),
		subIndex:   0,
	}
}

func (s *Manager) RegisterReadHandler(ctx context.Context, initSymbols []string, read wsDefine.ReadMarketHandler) error {
	batchSize := 100 // 每批最多订阅100个交易对，避免币安拒绝
	for i := 0; i < len(initSymbols); i += batchSize {
		end := i + batchSize
		if end > len(initSymbols) {
			end = len(initSymbols)
		}
		batch := initSymbols[i:end]
		obj := newBnContinuousKline()
		if err := obj.RegisterReadHandler(ctx, batch, read); err != nil {
			return err
		} else {
			s.subIndex++
			s.subs.Store(s.subIndex, obj)
			for _, symbol := range batch {
				s.subMapping[symbol] = s.subIndex
			}
			time.Sleep(200 * time.Millisecond) // 避免同时创建过多连接
		}
	}
	return nil
}

func (s *Manager) AddParamAnd(ctx context.Context, symbolName string) error {
	if subIndex, exists := s.subMapping[symbolName]; exists {
		if subObj, ok := s.subs.Load(subIndex); ok {
			return subObj.AddParamAnd(ctx, symbolName)
		}
	} else {
		// 不存在,添加到当前最末尾的订阅实例
		if subObj, ok := s.subs.Load(s.subIndex); ok {
			s.subMapping[symbolName] = s.subIndex
			return subObj.AddParamAnd(ctx, symbolName)
		}
	}
	return errorx.New(errCode.SYMBOL_NAME_NOT_EXISTS, "AddParamAnd_SYMBOL_NAME_NOT_FOUND, for  "+symbolName)
}

func (s *Manager) RemoveParamAnd(ctx context.Context, symbolName string) error {
	if subIndex, exists := s.subMapping[symbolName]; exists {
		if subObj, ok := s.subs.Load(subIndex); ok {
			return subObj.RemoveParamAnd(ctx, symbolName)
		}
	}
	return errorx.New(errCode.SYMBOL_NAME_NOT_EXISTS, "RemoveParamAnd_SYMBOL_NAME_NOT_FOUND, for  "+symbolName)
}

func (s *Manager) OpenSub(ctx context.Context) {
	s.subs.Range(func(key uint8, value *BnContinuousKline) bool {
		value.OpenSub(ctx)
		return true
	})
}

func (s *Manager) CloseSub(ctx context.Context) {
	s.subs.Range(func(key uint8, value *BnContinuousKline) bool {
		value.CloseSub(ctx)
		return true
	})
}

// BnContinuousKline 单个订阅实例
type BnContinuousKline struct {
	paramMan *wsSubParam.ParamMarket   // ws订阅参数管理器
	wsClient *wsMarketClient.Market    // WebSocket 客户端
	resource resourceEnum.ResourceType // 资源类型
	exType   exchangeEnum.ExchangeType // 交易所类型
}

func newBnContinuousKline() *BnContinuousKline {
	return &BnContinuousKline{
		paramMan: wsSubParam.NewParamMarket(wsSub.NewBnKline(klineEnum.KLINE_1s)),
		resource: resourceEnum.CONTINIOUS_KLINE,
		exType:   exchangeEnum.BINANCE,
	}
}

func (s *BnContinuousKline) RegisterReadHandler(ctx context.Context, initSymbols []string, read wsDefine.ReadMarketHandler) error {
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

func (s *BnContinuousKline) AddParamAnd(ctx context.Context, symbolName string) error {
	return s.paramMan.AddParamAnd(ctx, symbolName)
}

func (s *BnContinuousKline) RemoveParamAnd(ctx context.Context, symbolName string) error {
	return s.paramMan.RemoveParamAnd(ctx, symbolName)
}

func (s *BnContinuousKline) OpenSub(ctx context.Context) {
	s.wsClient.StartConn(ctx)
}

func (s *BnContinuousKline) CloseSub(ctx context.Context) {
	s.wsClient.CloseSub(ctx)
}
