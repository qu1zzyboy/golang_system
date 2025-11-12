package wsDialMarketImpl

import (
	"context"
	"upbitBnServer/internal/infra/ws/wsDefine"

	"github.com/gorilla/websocket"
)

type TreeNews struct {
}

func NewTreeNews() *TreeNews {
	return &TreeNews{}
}

func (s *TreeNews) DialTo(ctx context.Context) (*wsDefine.SafeWrite, error) {
	conn, _, err := websocket.DefaultDialer.Dial("wss://news.treeofalpha.com/ws", nil)
	if err != nil {
		return nil, wsDefine.ConnErr.WithCause(err)
	}
	return wsDefine.NewSafeWrite(conn), nil
}
