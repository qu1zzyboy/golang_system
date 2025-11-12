package wsDialMarketImpl

import (
	"context"
	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/ws/wsDefine"

	"github.com/gorilla/websocket"
)

const request_url = "wss://ws-fapi.binance.com/ws-fapi/v1?returnRateLimits=false"

type BnRequest struct {
}

func NewBnRequest() *BnRequest {
	return &BnRequest{}
}

func (s *BnRequest) DialTo(ctx context.Context) (*wsDefine.SafeWrite, error) {
	dynamicLog.Log.GetLog().Infof("开始连接[%s] ", request_url)
	conn, _, err := websocket.DefaultDialer.Dial(request_url, nil)
	if err != nil {
		return nil, wsDefine.ConnErr.WithCause(err).WithMetadata(map[string]string{defineJson.FullUrl: request_url})
	}
	return wsDefine.NewSafeWrite(conn), nil
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
