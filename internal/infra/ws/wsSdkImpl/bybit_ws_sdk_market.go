package wsSdkImpl

import (
	"context"
	"upbitBnServer/internal/define/defineOp"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/observe/trace/tracex"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/quant/market/aggTrade"
	"upbitBnServer/internal/quant/market/bookTick"
	"upbitBnServer/internal/quant/market/markPrice"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/pkg/utils/idGen"
	"upbitBnServer/pkg/utils/jsonUtils"

	"github.com/gorilla/websocket"
)

type byBitSubScribeReq struct {
	ReqId string   `json:"req_id"`
	Op    string   `json:"op"`
	Args  []string `json:"args"`
}

type ByBitMarket struct {
	conn *wsDefine.SafeWrite // websocket连接
}

func NewByBitMarket() *ByBitMarket {
	return &ByBitMarket{}
}

func (s *ByBitMarket) ParamBuild(resourceType resourceEnum.ResourceType) (func(symbol string) string, error) {
	switch resourceType {
	case resourceEnum.BOOK_TICK:
		return bookTick.BYBIT.GetFuSubParam, nil
	case resourceEnum.AGG_TRADE:
		return aggTrade.BYBIT.GetFuSubParam, nil
	case resourceEnum.MARK_PRICE:
		return markPrice.BYBIT.GetFuSubParam, nil
	default:
		return nil, resourceType.GetNotSupportError("ByBitMarket_ParamBuild")
	}
}

func (s *ByBitMarket) DialToMarket(ctx context.Context, params []string) (*wsDefine.SafeWrite, error) {
	conn, _, err := websocket.DefaultDialer.Dial("wss://stream.bybit.com/v5/public/linear", nil)
	if err != nil {
		return nil, wsDefine.ConnErr.WithCause(err)
	}
	s.conn = wsDefine.NewSafeWrite(conn)
	return s.conn, s.AddSub(ctx, params)
}

func (s *ByBitMarket) AddSub(ctx context.Context, params []string) error {
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

func (s *ByBitMarket) ListSub(ctx context.Context) error {
	return nil
}

func (s *ByBitMarket) UnSub(ctx context.Context, params []string) error {
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

func (s *ByBitMarket) sendWsData(ctx context.Context, op string, jsonBytes []byte) error {
	return tracex.WithTrace(ctx, true, op, func(ctx context.Context) error {
		dynamicLog.Log.GetLog().Infof("ByBitMarket: %s", string(jsonBytes))
		return s.conn.SafeWriteMsg(websocket.TextMessage, jsonBytes)
	})
}

func (s *ByBitMarket) getTraceId(op string) string {
	return idGen.BuildName2("ByBitMarket", op)
}
