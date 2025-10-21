package wsMarketClient

import (
	"context"

	"github.com/hhh500/quantGoInfra/infra/ws/wsDefine"
	"github.com/hhh500/quantGoInfra/infra/ws/wsReConn"
	"github.com/hhh500/quantGoInfra/infra/ws/wsSub"
	"github.com/hhh500/quantGoInfra/pkg/utils/idGen"
	"github.com/hhh500/quantGoInfra/quant/exchanges/exchangeEnum"
	"github.com/hhh500/quantGoInfra/resource/resourceEnum"
)

type Market struct {
	conn   *wsReConn.ReConnMarket // ws连接资源管理
	cancel context.CancelFunc     // 取消连接的函数
}

func NewMarket(exType exchangeEnum.ExchangeType, resource resourceEnum.ResourceType, read wsDefine.ReadMarketHandler, subParam wsSub.SubParam) (*Market, error) {
	// 创建一个可以取消的上下文
	ctxStop, cancel := context.WithCancel(context.Background())
	c := &Market{
		conn:   wsReConn.NewReConnMarket(exType, resource, subParam, read, idGen.BuildName2(exType.String(), resource.String())),
		cancel: cancel,
	}
	c.conn.ReConnLoop(ctxStop)
	return c, nil
}

func (s *Market) StartConn(ctx context.Context) error {
	s.conn.ReceiveSig(wsDefine.START_CONN)
	return nil
}

// 关闭这次订阅
func (s *Market) CloseSub(ctx context.Context) {
	s.conn.CloseSub(ctx)
}

// 关闭整个链接
func (s *Market) CloseConn(ctx context.Context) error {
	s.cancel()
	return nil
}
