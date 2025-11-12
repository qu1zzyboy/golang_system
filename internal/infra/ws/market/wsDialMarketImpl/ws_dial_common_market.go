package wsDialMarketImpl

import (
	"context"
	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/pkg/container/map/myMap"
)

var buildFuncNilErr = errorx.Newf(errCode.POINTER_NIL, "ws订阅参数构建函数为nil")

type CommonMarket struct {
	params     myMap.MySyncMap[string, struct{}] //所有需要订阅的参数
	subI       wsDefine.SubParamSdk              //加减订阅的接口
	paramBuild func(symbol string) string        //参数构建函数
}

func NewCommonMarket(subI wsDefine.SubParamSdk) *CommonMarket {
	return &CommonMarket{
		params: myMap.NewMySyncMap[string, struct{}](),
		subI:   subI,
	}
}

func (s *CommonMarket) SetInitSymbols(resourceType resourceEnum.ResourceType, symbols []string) error {
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

func (s *CommonMarket) AddParamAnd(ctx context.Context, symbolName string) error {
	paramKey := s.paramBuild(symbolName)
	dynamicLog.Log.GetLog().Info("CommonMarket 准备添加订阅symbol:", paramKey)
	s.params.Store(paramKey, struct{}{})
	return s.subI.AddSub(ctx, []string{paramKey})
}

func (s *CommonMarket) RemoveParamAnd(ctx context.Context, symbolName string) error {
	paramKey := s.paramBuild(symbolName)
	dynamicLog.Log.GetLog().Info("CommonMarket 准备减少订阅symbol:", paramKey)
	s.params.Delete(paramKey)
	return s.subI.UnSub(ctx, []string{paramKey})
}

func (s *CommonMarket) getAllParams() []string {
	paramArr := make([]string, 0, s.params.Length())
	s.params.Range(func(key string, _ struct{}) bool {
		paramArr = append(paramArr, key)
		return true
	})
	return paramArr
}

func (s *CommonMarket) DialTo(ctx context.Context) (*wsDefine.SafeWrite, error) {
	return s.subI.DialToMarket(ctx, s.getAllParams())
}
