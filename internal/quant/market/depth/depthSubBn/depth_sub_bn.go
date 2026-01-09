package depthSubBn

import (
	"context"
	"time"

	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/internal/resource/wsMarketClient"
	"upbitBnServer/internal/resource/wsSub"
	"upbitBnServer/internal/resource/wsSubParam"
	"upbitBnServer/pkg/container/map/myMap"
	"upbitBnServer/pkg/singleton"
)

var (
	depthSingleton = singleton.NewSingleton(func() *Manager { return newManager() })
)

func GetManager() *Manager {
	return depthSingleton.Get()
}

// Manager 管理多个 BnDepth 实例，支持分批订阅
type Manager struct {
	subMapping map[string]uint8              // symbolName-->订阅序号
	subs       myMap.MySyncMap[uint8, *BnDepth] // 订阅序号-->订阅服务
	subIndex   uint8                          // 当前订阅实例索引
	levels     int                            // 深度档位
	updateSpeed int                           // 更新速度（毫秒）
}

func newManager() *Manager {
	return &Manager{
		subMapping: make(map[string]uint8),
		subs:       myMap.NewMySyncMap[uint8, *BnDepth](),
		subIndex:   0,
		levels:     20,  // 默认20档
		updateSpeed: 500, // 默认500ms
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
		obj := newBnDepth(s.levels, s.updateSpeed)
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
	s.subs.Range(func(key uint8, value *BnDepth) bool {
		value.OpenSub(ctx)
		return true
	})
}

func (s *Manager) CloseSub(ctx context.Context) {
	s.subs.Range(func(key uint8, value *BnDepth) bool {
		value.CloseSub(ctx)
		return true
	})
}

// BnDepth 单个订阅实例
type BnDepth struct {
	paramMan    *wsSubParam.ParamMarket   // ws订阅参数管理器
	wsClient    *wsMarketClient.Market    // WebSocket 客户端
	resource    resourceEnum.ResourceType // 资源类型
	exType      exchangeEnum.ExchangeType // 交易所类型
	levels      int                       // 深度档位
	updateSpeed int                       // 更新速度（毫秒）
}

func newBnDepth(levels, updateSpeed int) *BnDepth {
	return &BnDepth{
		paramMan:    wsSubParam.NewParamMarket(wsSub.NewBnDepth(levels, updateSpeed)),
		resource:    resourceEnum.DELTA_DEPTH,
		exType:      exchangeEnum.BINANCE,
		levels:      levels,
		updateSpeed: updateSpeed,
	}
}

func (s *BnDepth) RegisterReadHandler(ctx context.Context, initSymbols []string, read wsDefine.ReadMarketHandler) error {
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

func (s *BnDepth) AddParamAnd(ctx context.Context, symbolName string) error {
	return s.paramMan.AddParamAnd(ctx, symbolName)
}

func (s *BnDepth) RemoveParamAnd(ctx context.Context, symbolName string) error {
	return s.paramMan.RemoveParamAnd(ctx, symbolName)
}

func (s *BnDepth) OpenSub(ctx context.Context) {
	s.wsClient.StartConn(ctx)
}

func (s *BnDepth) CloseSub(ctx context.Context) {
	s.wsClient.CloseSub(ctx)
}

