package wsDialImpl

import (
	"context"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/infra/ws/wsSdkImpl"
	"upbitBnServer/pkg/container/map/myMap"
)

type BybitPayload struct {
	params myMap.MySyncMap[string, struct{}] //所有需要订阅的参数
	subI   *wsSdkImpl.ByBitPrivate           //加减订阅的接口
}

func NewBybitPayload(apiKey, secretKey string) *BybitPayload {
	return &BybitPayload{
		params: myMap.NewMySyncMap[string, struct{}](),
		subI:   wsSdkImpl.NewByBitPrivate("wss://stream.bybit.com/v5/private", apiKey, secretKey),
	}
}

func (s *BybitPayload) SetInitParams(params []string) error {
	for _, param := range params {
		s.params.Store(param, struct{}{})
	}
	return nil
}

func (s *BybitPayload) getAllParams() []string {
	paramArr := make([]string, 0, s.params.Length())
	s.params.Range(func(key string, _ struct{}) bool {
		paramArr = append(paramArr, key)
		return true
	})
	return paramArr
}

func (s *BybitPayload) DialTo(ctx context.Context) (*wsDefine.SafeWrite, error) {
	return s.subI.DialToPrivate(ctx, s.getAllParams())
}
