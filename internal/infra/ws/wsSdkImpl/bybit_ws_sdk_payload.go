package wsSdkImpl

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"upbitBnServer/internal/define/defineJson"
	"upbitBnServer/internal/define/defineOp"
	"upbitBnServer/internal/infra/observe/log/staticLog"
	"upbitBnServer/internal/infra/observe/trace/tracex"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/pkg/utils/idGen"
	"upbitBnServer/pkg/utils/jsonUtils"
	"upbitBnServer/pkg/utils/timeUtils"

	"github.com/gorilla/websocket"
)

type ByBitPrivate struct {
	secretByte []byte              // 私钥
	baseUrl    string              // 当前订阅的URL
	apiKey     string              // API Key
	conn       *wsDefine.SafeWrite // websocket连接
}

func NewByBitPrivate(url, apiKey, secretKey string) *ByBitPrivate {
	return &ByBitPrivate{baseUrl: url, apiKey: apiKey, secretByte: []byte(secretKey)}
}

func (s *ByBitPrivate) sendAuth(ctx context.Context) error {
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
	return s.sendWsData(ctx, s.getTraceId(sub_AUTH), authByte)
}

func (s *ByBitPrivate) DialToPrivate(ctx context.Context, params []string) (*wsDefine.SafeWrite, error) {
	conn, _, err := websocket.DefaultDialer.Dial(s.baseUrl, nil)
	if err != nil {
		return nil, wsDefine.ConnErr.WithCause(err).WithMetadata(map[string]string{defineJson.FullUrl: s.baseUrl})
	}
	s.conn = wsDefine.NewSafeWrite(conn)
	if err := s.sendAuth(ctx); err != nil {
		return nil, err
	}
	if len(params) > 0 {
		return s.conn, s.AddSub(ctx, params)
	}
	return s.conn, nil
}

func (s *ByBitPrivate) AddSub(ctx context.Context, params []string) error {
	staticLog.Log.Infof("WsSubBybitSdk 增加订阅[%d] bybit 数据,参数: %v", len(params), params)
	// 1、获取当前的雪花ID
	// 2、序列化请求
	var jsonBytes []byte
	if err := tracex.WithTrace(ctx, true, s.getTraceId(defineOp.OP_MARSHAL_REQ), func(ctx context.Context) error {
		req := byBitSubScribeReq{
			Op:    "subscribe",
			Args:  params,
			ReqId: convertx.GetNowTimeStampMilliStr(),
		}
		jsonBytes_, err := jsonUtils.MarshalStructToByteArray(req)
		jsonBytes = jsonBytes_
		return err
	}); err != nil {
		return err
	}
	// 3 、发送请求
	return s.sendWsData(ctx, s.getTraceId(sub_ADD), jsonBytes)
}

func (s *ByBitPrivate) ListSub(ctx context.Context) error {
	return nil
}

func (s *ByBitPrivate) UnSub(ctx context.Context, params []string) error {
	staticLog.Log.Infof("WsSubBybitSdk 减少订阅 bn 数据,参数: %v", params)
	// 1、获取当前的雪花ID
	// 2、序列化请求
	var jsonBytes []byte
	if err := tracex.WithTrace(ctx, true, s.getTraceId(defineOp.OP_MARSHAL_REQ), func(ctx context.Context) error {
		req := byBitSubScribeReq{
			Op:    "unsubscribe",
			Args:  params,
			ReqId: convertx.GetNowTimeStampMilliStr(),
		}
		jsonBytes_, err := jsonUtils.MarshalStructToByteArray(req)
		jsonBytes = jsonBytes_
		return err
	}); err != nil {
		return err
	}
	// 3 、发送请求
	return s.sendWsData(ctx, s.getTraceId(sub_DEL), jsonBytes)
}

func (s *ByBitPrivate) sendWsData(ctx context.Context, op string, jsonBytes []byte) error {
	return tracex.WithTrace(ctx, true, op, func(ctx context.Context) error { return s.conn.SafeWriteMsg(websocket.TextMessage, jsonBytes) })
}

func (s *ByBitPrivate) getTraceId(op string) string {
	return idGen.BuildName2("ByBitPrivate", op)
}

func (s *ByBitPrivate) GetUrl() string {
	return s.baseUrl
}
