package wsSub

import (
	"context"

	"github.com/gorilla/websocket"
	"github.com/hhh500/quantGoInfra/define/defineJson"
	"github.com/hhh500/quantGoInfra/infra/observe/log/dynamicLog"
	"github.com/hhh500/quantGoInfra/infra/ws/wsDefine"
)

type BnOrder struct {
	secret  string              // 私钥
	baseUrl string              // 当前订阅的URL
	apiKey  string              // API Key
	conn    *wsDefine.SafeWrite // websocket连接
}

func NewBnOrder(apiKey, secretKey string) *BnOrder {
	return &BnOrder{
		// baseUrl: "wss://ws-fapi.binance.com/ws-fapi/v1",
		baseUrl: "wss://ws-fapi.binance.com/ws-fapi/v1?returnRateLimits=false",
		apiKey:  apiKey,
		secret:  secretKey,
	}
}

func (s *BnOrder) DialTo(ctx context.Context) (*wsDefine.SafeWrite, error) {
	dynamicLog.Log.GetLog().Infof("开始连接[%s] ", s.baseUrl)
	conn, _, err := websocket.DefaultDialer.Dial(s.baseUrl, nil)
	if err != nil {
		return nil, connErr.WithCause(err).WithMetadata(map[string]string{defineJson.FullUrl: s.baseUrl})
	}
	s.conn = wsDefine.NewSafeWrite(conn)
	return s.conn, nil
}

func (s *BnOrder) GetUrl() string {
	return s.baseUrl
}

// {
// 	"id": "c174a2b1-3f51-4580-b200-8528bd237cb7",
// 	"status": 200,
// 	"result": {
// 		"apiKey": "r8lTMO0eDWNLrkK7VYQQwOG12izakNViSKf20pqpXXrUs4NKyyBTF7EjcJiBLkVc",
// 		"authorizedSince": 1758012017879,
// 		"connectedSince": 1758012017873,
// 		"returnRateLimits": true,
// 		"serverTime": 1758012017879
// 	},
// 	"rateLimits": [{
// 		"rateLimitType": "REQUEST_WEIGHT",
// 		"interval": "MINUTE",
// 		"intervalNum": 1,
// 		"limit": 2400,
// 		"count": 7
// 	}]
// }
