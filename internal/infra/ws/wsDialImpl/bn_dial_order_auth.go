package wsDialImpl

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/url"
	"time"
	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/pkg/utils/jsonUtils"
	"upbitBnServer/pkg/utils/time2str"

	"github.com/gorilla/websocket"
)

type BnOrderAuth struct {
	secret  string              // 私钥
	baseUrl string              // 当前订阅的URL
	apiKey  string              // API Key
	conn    *wsDefine.SafeWrite // websocket连接
}

func NewBnOrderAuth(apiKey, secretKey string) *BnOrderAuth {
	return &BnOrderAuth{
		baseUrl: "wss://ws-fapi.binance.com/ws-fapi/v1?returnRateLimits=false",
		apiKey:  apiKey,
		secret:  secretKey,
	}
}

func (s *BnOrderAuth) DialTo(ctx context.Context) (*wsDefine.SafeWrite, error) {
	dynamicLog.Log.GetLog().Infof("开始连接[%s] ", s.baseUrl)
	conn, _, err := websocket.DefaultDialer.Dial(s.baseUrl, nil)
	if err != nil {
		return nil, wsDefine.ConnErr.WithCause(err).WithMetadata(map[string]string{defineJson.FullUrl: s.baseUrl})
	}
	s.conn = wsDefine.NewSafeWrite(conn)
	if err := s.sendAuth(); err != nil {
		return nil, err
	}
	return s.conn, nil
}

func (s *BnOrderAuth) GetUrl() string {
	return s.baseUrl
}

func (s *BnOrderAuth) sendAuth() error {
	// 1. Base64 解码
	der, err := base64.StdEncoding.DecodeString(s.secret)
	if err != nil {
		errorx.PanicWithCaller(err.Error())
	}

	// 2. 解析 PKCS#8
	key, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		errorx.PanicWithCaller(err.Error())
	}

	// 3. 类型断言为 ed25519.PrivateKey
	privKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		panic("not an Ed25519 private key")
	}

	ts := time.Now().UnixMilli()

	// 模拟请求参数
	params := url.Values{}
	params.Add("apiKey", s.apiKey)
	params.Add("timestamp", fmt.Sprintf("%d", ts))

	// 1. 生成 query string
	queryString := params.Encode()

	// 3. 签名
	signature := ed25519.Sign(privKey, []byte(queryString))

	// 4. base64 编码
	sigBase64 := base64.StdEncoding.EncodeToString(signature)

	reqId := time2str.GetNowTimeStampMicroSlice16()
	// 5. 拼接 URL
	rawData, err := jsonUtils.MarshalStructToByteArray(struct {
		ID     string `json:"id"`
		Method string `json:"method"`
		Params struct {
			ApiKey    string `json:"apiKey"`
			Signature string `json:"signature"`
			Timestamp int64  `json:"timestamp"`
		} `json:"params"`
	}{
		ID:     string(reqId[:]),
		Method: "session.logon",
		Params: struct {
			ApiKey    string `json:"apiKey"`
			Signature string `json:"signature"`
			Timestamp int64  `json:"timestamp"`
		}{
			ApiKey:    s.apiKey,
			Timestamp: ts,
			Signature: sigBase64,
		},
	})
	if err != nil {
		return err
	}
	return s.conn.SafeWriteMsg(websocket.TextMessage, rawData)
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
