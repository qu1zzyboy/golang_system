package markPriceSubBn

import (
	"context"
	"time"

	"github.com/hhh500/quantGoInfra/infra/errorx"
	"github.com/hhh500/quantGoInfra/infra/errorx/errCode"
	"github.com/hhh500/quantGoInfra/infra/ws/wsDefine"
	"github.com/hhh500/quantGoInfra/pkg/container/map/myMap"
	"github.com/hhh500/quantGoInfra/pkg/singleton"
)

var (
	bnSingleton = singleton.NewSingleton(func() *Manager { return newService() })
)

func GetManager() *Manager {
	return bnSingleton.Get()
}

type Manager struct {
	subMapping map[string]uint8                     // symbolName-->订阅序号
	subs       myMap.MySyncMap[uint8, *BnMarkPrice] // 订阅序号-->订阅服务
	subIndex   uint8                                // 当前订阅实例索引
}

func newService() *Manager {
	return &Manager{
		subMapping: make(map[string]uint8),
		subs:       myMap.NewMySyncMap[uint8, *BnMarkPrice](),
		subIndex:   0,
	}
}

func (s *Manager) RegisterReadHandler(ctx context.Context, initSymbols []string, read wsDefine.ReadMarketHandler) error {
	batchSize := 100
	for i := 0; i < len(initSymbols); i += batchSize {
		end := min(i+batchSize, len(initSymbols))
		batch := initSymbols[i:end]
		obj := newBnMarkPrice()
		if err := obj.RegisterReadHandler(ctx, batch, read); err != nil {
			return err
		} else {
			s.subIndex++
			s.subs.Store(s.subIndex, obj)
			for _, symbol := range batch {
				s.subMapping[symbol] = s.subIndex
			}
			// 避免同时创建过多连接
			time.Sleep(200 * time.Millisecond)
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
	s.subs.Range(func(key uint8, value *BnMarkPrice) bool {
		value.OpenSub(ctx)
		return true
	})
}

func (s *Manager) CloseSub(ctx context.Context) {
	s.subs.Range(func(key uint8, value *BnMarkPrice) bool {
		value.CloseSub(ctx)
		return true
	})
}
