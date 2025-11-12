package wsDialImpl

import (
	"context"
	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/ws/wsDefine"

	"github.com/gorilla/websocket"
)

type BnOrder struct {
	secret  string              // 私钥
	baseUrl string              // 当前订阅的URL
	apiKey  string              // API Key
	conn    *wsDefine.SafeWrite // websocket连接
}

func NewBnOrder(apiKey, secretKey string) *BnOrder {
	return &BnOrder{
		baseUrl: "wss://ws-fapi.binance.com/ws-fapi/v1?returnRateLimits=false",
		apiKey:  apiKey,
		secret:  secretKey,
	}
}

func (s *BnOrder) DialTo(ctx context.Context) (*wsDefine.SafeWrite, error) {
	dynamicLog.Log.GetLog().Infof("开始连接[%s] ", s.baseUrl)
	conn, _, err := websocket.DefaultDialer.Dial(s.baseUrl, nil)
	if err != nil {
		return nil, wsDefine.ConnErr.WithCause(err).WithMetadata(map[string]string{defineJson.FullUrl: s.baseUrl})
	}
	s.conn = wsDefine.NewSafeWrite(conn)
	return s.conn, nil
}

func (s *BnOrder) GetUrl() string {
	return s.baseUrl
}
