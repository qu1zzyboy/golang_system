package treeNewsSub

import (
	"context"
	"upbitBnServer/internal/infra/ws/client/wsMarketClient"
	"upbitBnServer/internal/infra/ws/market/wsDialMarketImpl"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/pkg/singleton"
)

var (
	bnSingleton = singleton.NewSingleton(func() *Sub { return newService() })
)

func Get() *Sub {
	return bnSingleton.Get()
}

type Sub struct {
	paramMan *wsDialMarketImpl.TreeNews // ws订阅参数管理器
	wsClient *wsMarketClient.AutoMarket // WebSocket 客户端
	resource resourceEnum.ResourceType  // 资源类型
	exType   exchangeEnum.ExchangeType  // 交易所类型
}

func newService() *Sub {
	return &Sub{
		paramMan: wsDialMarketImpl.NewTreeNews(),
		resource: resourceEnum.TREE_NEWS,
		exType:   exchangeEnum.TREE_NEWS,
	}
}

func (s *Sub) RegisterReadHandler(ctx context.Context, read wsDefine.ReadAutoHandler) error {
	var err error
	s.wsClient, err = wsMarketClient.NewAutoMarket(s.exType, s.resource, read, s.paramMan)
	if err != nil {
		return err
	}
	return s.wsClient.StartConn(ctx)
}
