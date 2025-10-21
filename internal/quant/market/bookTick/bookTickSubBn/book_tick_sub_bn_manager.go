package bookTickSubBn

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
	subMapping   map[string]uint8                    // symbolName-->订阅序号,方便取消订阅
	subs         myMap.MySyncMap[uint8, *BnBootTick] // 订阅序号-->订阅服务
	thisSubIndex uint8                               // 当前订阅实例索引
}

func newService() *Manager {
	return &Manager{
		subMapping:   make(map[string]uint8),
		subs:         myMap.NewMySyncMap[uint8, *BnBootTick](),
		thisSubIndex: 0,
	}
}

func (s *Manager) RegisterReadHandler(ctx context.Context, initSymbolNames []string, read wsDefine.ReadMarketHandler) error {
	batchSize := 100
	for i := 0; i < len(initSymbolNames); i += batchSize {
		end := min(i+batchSize, len(initSymbolNames))
		batch := initSymbolNames[i:end]
		obj := newBnBootTick()
		if err := obj.RegisterReadHandler(ctx, batch, read); err != nil {
			return err
		} else {
			s.thisSubIndex++
			s.subs.Store(s.thisSubIndex, obj)
			for _, symbolName := range batch {
				s.subMapping[symbolName] = s.thisSubIndex
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
		if subObj, ok := s.subs.Load(s.thisSubIndex); ok {
			s.subMapping[symbolName] = s.thisSubIndex
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
	s.subs.Range(func(key uint8, value *BnBootTick) bool {
		value.OpenSub(ctx)
		return true
	})
}

func (s *Manager) CloseSub(ctx context.Context) {
	s.subs.Range(func(key uint8, value *BnBootTick) bool {
		value.CloseSub(ctx)
		return true
	})
}
