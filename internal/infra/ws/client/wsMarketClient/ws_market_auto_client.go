package wsMarketClient

import (
	"context"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/infra/ws/wsReConn"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/pkg/utils/idGen"
)

type AutoMarket struct {
	conn   *wsReConn.ReConnAuto // ws连接资源管理
	cancel context.CancelFunc   // 取消连接的函数
}

func NewAutoMarket(exType exchangeEnum.ExchangeType, resource resourceEnum.ResourceType, read wsDefine.ReadAutoHandler, subParam wsDefine.SubDial) (*AutoMarket, error) {
	// 创建一个可以取消的上下文
	ctxStop, cancel := context.WithCancel(context.Background())
	c := &AutoMarket{
		conn:   wsReConn.NewReConnAuto(exType, resource, subParam, read, idGen.BuildName2(exType.String(), resource.String())),
		cancel: cancel,
	}
	c.conn.ReConnLoop(ctxStop)
	return c, nil
}

func (s *AutoMarket) StartConn(ctx context.Context) error {
	s.conn.ReceiveSig(wsDefine.START_CONN)
	return nil
}

// 关闭这次订阅
func (s *AutoMarket) CloseSub(ctx context.Context) {
	s.conn.CloseSub(ctx)
}

// 关闭整个链接
func (s *AutoMarket) CloseConn(ctx context.Context) error {
	s.cancel()
	return nil
}
