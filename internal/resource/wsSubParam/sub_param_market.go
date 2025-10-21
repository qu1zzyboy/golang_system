package wsSubParam

import (
	"context"

	"github.com/hhh500/quantGoInfra/infra/errorx"
	"github.com/hhh500/quantGoInfra/infra/errorx/errCode"
	"github.com/hhh500/quantGoInfra/infra/observe/log/dynamicLog"
	"github.com/hhh500/quantGoInfra/infra/ws/wsDefine"
	"github.com/hhh500/quantGoInfra/infra/ws/wsSub"
	"github.com/hhh500/quantGoInfra/pkg/container/map/myMap"
	"github.com/hhh500/quantGoInfra/resource/resourceEnum"
)

var buildFuncNilErr = errorx.Newf(errCode.POINTER_NIL, "ws订阅参数构建函数为nil")

type ParamMarket struct {
	params     myMap.MySyncMap[string, struct{}] //所有需要订阅的参数
	subI       wsSub.WsSub                       //加减订阅的接口
	paramBuild func(symbol string) string        //参数构建函数
}

func NewParamMarket(subI wsSub.WsSub) *ParamMarket {
	return &ParamMarket{
		params: myMap.NewMySyncMap[string, struct{}](),
		subI:   subI,
	}
}

func (s *ParamMarket) SetInitSymbols(resourceType resourceEnum.ResourceType, symbols []string) error {
	var err error
	s.paramBuild, err = s.subI.ParamBuild(resourceType)
	if err != nil {
		return err
	}
	if s.paramBuild == nil {
		return buildFuncNilErr
	}
	for _, symbol := range symbols {
		s.params.Store(s.paramBuild(symbol), struct{}{})
	}
	return nil
}

func (s *ParamMarket) AddParamAnd(ctx context.Context, symbolName string) error {
	paramKey := s.paramBuild(symbolName)
	dynamicLog.Log.GetLog().Info("ParamMarket 准备添加订阅symbol:", paramKey)
	s.params.Store(paramKey, struct{}{})
	return s.subI.AddSub(ctx, []string{paramKey})
}

func (s *ParamMarket) RemoveParamAnd(ctx context.Context, symbolName string) error {
	paramKey := s.paramBuild(symbolName)
	dynamicLog.Log.GetLog().Info("ParamMarket 准备减少订阅symbol:", paramKey)
	s.params.Delete(paramKey)
	return s.subI.UnSub(ctx, []string{paramKey})
}

func (s *ParamMarket) getAllParams() []string {
	paramArr := make([]string, 0, s.params.Length())
	s.params.Range(func(key string, _ struct{}) bool {
		paramArr = append(paramArr, key)
		return true
	})
	return paramArr
}

func (s *ParamMarket) DialTo(ctx context.Context) (*wsDefine.SafeWrite, error) {
	return s.subI.DialToMarket(ctx, s.getAllParams())
}
