package wsDialMarketImpl

import (
	"context"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/pkg/container/map/myMap"
)

type DynamicSymbol struct {
	params myMap.MySyncMap[string, struct{}] //所有需要订阅的参数
	subI   wsDefine.SubParamSdk              //加减订阅的接口
}

func NewDynamicSymbol(subI wsDefine.SubParamSdk) *DynamicSymbol {
	return &DynamicSymbol{
		subI:   subI,
		params: myMap.NewMySyncMap[string, struct{}](),
	}
}

func (s *DynamicSymbol) SetInitSymbols(resourceTypes []resourceEnum.ResourceType, symbolName string) error {
	for _, resourceType := range resourceTypes {
		fn, err := s.subI.ParamBuild(resourceType)
		if err != nil {
			return err
		}
		s.params.Store(fn(symbolName), struct{}{})
	}
	return nil
}

func (s *DynamicSymbol) AddSub(ctx context.Context, resourceType resourceEnum.ResourceType, symbolName string) error {
	fn, err := s.subI.ParamBuild(resourceType)
	if err != nil {
		return err
	}
	param := fn(symbolName)
	s.params.Store(param, struct{}{})
	return s.subI.AddSub(ctx, []string{param})
}

func (s *DynamicSymbol) UnSub(ctx context.Context, resourceType resourceEnum.ResourceType, symbolName string) error {
	fn, err := s.subI.ParamBuild(resourceType)
	if err != nil {
		return err
	}
	param := fn(symbolName)
	s.params.Delete(param)
	return s.subI.UnSub(ctx, []string{param})
}

func (s *DynamicSymbol) DialTo(ctx context.Context) (*wsDefine.SafeWrite, error) {
	var params []string
	s.params.Range(func(k string, v struct{}) bool {
		params = append(params, k)
		return true
	})
	return s.subI.DialToMarket(ctx, params)
}
