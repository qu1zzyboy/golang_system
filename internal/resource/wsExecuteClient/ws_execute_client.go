package wsExecuteClient

import (
	"context"

	"github.com/hhh500/quantGoInfra/infra/ws/wsDefine"
	"github.com/hhh500/quantGoInfra/infra/ws/wsReConn"
	"github.com/hhh500/quantGoInfra/infra/ws/wsSub"
	"github.com/hhh500/quantGoInfra/pkg/utils/convertx"
	"github.com/hhh500/quantGoInfra/pkg/utils/idGen"
	"github.com/hhh500/quantGoInfra/quant/exchanges/exchangeEnum"
	"github.com/hhh500/quantGoInfra/resource/resourceEnum"
)

type Execute struct {
	conn   *wsReConn.ReConnPrivate // ws连接资源管理
	cancel context.CancelFunc      // 取消连接的函数
}

func NewExecute(exType exchangeEnum.ExchangeType, resource resourceEnum.ResourceType, read wsDefine.ReadPrivateHandler, subParam wsSub.SubParam, accountKeyId uint8) (*Execute, error) {
	// 创建一个可以取消的上下文
	ctxStop, cancel := context.WithCancel(context.Background())
	c := &Execute{
		conn:   wsReConn.NewReConnPrivate(exType, resource, subParam, read, idGen.BuildName3(exType.String(), resource.String(), convertx.ToString(accountKeyId))),
		cancel: cancel,
	}
	c.conn.ReConnLoop(ctxStop)
	return c, nil
}

func (s *Execute) StartConn(ctx context.Context) error {
	s.conn.ReceiveSig(wsDefine.START_CONN)
	return nil
}

func (s *Execute) CloseConn(ctx context.Context) error {
	s.cancel() // 取消连接
	return nil
}

func (s *Execute) IsConnOk() bool { return s.conn.IsConnOk() }

func (s *Execute) WriteAsync(data []byte) error {
	return s.conn.WriteAsync(data)
}
