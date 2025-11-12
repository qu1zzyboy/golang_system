package wsDialImpl

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/infra/observe/trace/tracex"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/pkg/utils/idGen"
	"upbitBnServer/pkg/utils/jsonUtils"
	"upbitBnServer/pkg/utils/timeUtils"

	"github.com/gorilla/websocket"
)

type ByBitOrder struct {
	secretByte []byte              // 私钥
	baseUrl    string              // 当前订阅的URL
	apiKey     string              // API Key
	conn       *wsDefine.SafeWrite // websocket连接
}

func NewByBitOrder(apiKey, secretKey string) *ByBitOrder {
	return &ByBitOrder{
		baseUrl:    "wss://stream.bybit.com/v5/trade",
		apiKey:     apiKey,
		secretByte: []byte(secretKey),
	}
}

func (s *ByBitOrder) sendAuth(ctx context.Context) error {
	expires := timeUtils.GetNowTimeUnixMilli() + 10000
	val := fmt.Sprintf("GET/realtime%d", expires)
	h := hmac.New(sha256.New, s.secretByte)
	h.Write([]byte(val))
	signature := hex.EncodeToString(h.Sum(nil))
	authMessage := map[string]any{
		"req_id": convertx.GetNowTimeStampMilliStr(),
		"op":     "auth",
		"args":   []any{s.apiKey, expires, signature},
	}
	authByte, err := jsonUtils.MarshalStructToByteArray(authMessage)
	if err != nil {
		return err
	}
	return s.sendWsData(ctx, s.getTraceId("auth"), authByte)
}

func (s *ByBitOrder) DialTo(ctx context.Context) (*wsDefine.SafeWrite, error) {
	conn, _, err := websocket.DefaultDialer.Dial(s.baseUrl, nil)
	if err != nil {
		return nil, wsDefine.ConnErr.WithCause(err).WithMetadata(map[string]string{defineJson.FullUrl: s.baseUrl})
	}
	s.conn = wsDefine.NewSafeWrite(conn)
	if err = s.sendAuth(ctx); err != nil {
		return nil, err
	}
	return s.conn, nil
}

func (s *ByBitOrder) sendWsData(ctx context.Context, op string, jsonBytes []byte) error {
	return tracex.WithTrace(ctx, true, op, func(ctx context.Context) error { return s.conn.SafeWriteMsg(websocket.TextMessage, jsonBytes) })
}

func (s *ByBitOrder) getTraceId(op string) string {
	return idGen.BuildName2("ByBitOrder", op)
}
