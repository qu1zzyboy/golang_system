package wsDialMarketImpl

import (
	"context"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/resource/resourceEnum"
)

type StaticSymbol struct {
	params []string             //所有需要订阅的参数
	subI   wsDefine.SubParamSdk //加减订阅的接口
}

func NewStaticSymbol(subI wsDefine.SubParamSdk) *StaticSymbol {
	return &StaticSymbol{
		subI: subI,
	}
}

func (s *StaticSymbol) SetInitSymbols(resourceTypes []resourceEnum.ResourceType, symbolName string) error {
	for _, resourceType := range resourceTypes {
		fn, err := s.subI.ParamBuild(resourceType)
		if err != nil {
			return err
		}
		s.params = append(s.params, fn(symbolName))
	}
	return nil
}

func (s *StaticSymbol) DialTo(ctx context.Context) (*wsDefine.SafeWrite, error) {
	return s.subI.DialToMarket(ctx, s.params)
}
