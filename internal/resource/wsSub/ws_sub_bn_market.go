package wsSub

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
	"upbitBnServer/pkg/utils/idGen"
	"upbitBnServer/pkg/utils/jsonUtils"

	"github.com/gorilla/websocket"
)

type bnSubScribeReq struct {
	Method string   `json:"method"`
	Params []string `json:"params,omitempty"`
	Id     int64    `json:"id"`
}

// {"result":null,"id":1950383391484416000}
// {"result":["btcusdt@bookTicker","ethusdt@bookTicker"],"id":1950383391484416000}

type BnMarket struct {
	conn *wsDefine.SafeWrite // websocket连接
}

func NewBnMarket() *BnMarket {
	return &BnMarket{}
}

func (s *BnMarket) DialToMarket(ctx context.Context, params []string) (*wsDefine.SafeWrite, error) {
	conn, _, err := websocket.DefaultDialer.Dial("wss://fstream.binance.com/ws", nil)
	if err != nil {
		return nil, connErr.WithCause(err)
	}
	s.conn = wsDefine.NewSafeWrite(conn)
	return s.conn, s.AddSub(ctx, params)
}

func (s *BnMarket) ParamBuild(resourceType resourceEnum.ResourceType) (func(symbol string) string, error) {
	switch resourceType {
	case resourceEnum.BOOK_TICK:
		return bookTick.BN.GetFuSubParam, nil
	case resourceEnum.AGG_TRADE:
		return aggTrade.BN.GetFuSubParam, nil
	case resourceEnum.MARK_PRICE:
		return markPrice.BN.GetFuSubParam, nil
	default:
		return nil, resourceType.GetNotSupportError("bnParamBuild")
	}
}

func (s *BnMarket) AddSub(ctx context.Context, params []string) error {
	// 1、获取当前的雪花ID
	var snowId int64
	if err := tracex.WithTrace(ctx, true, s.getTraceId(defineOp.OP_GET_SNOWFLAKE_ID), func(ctx context.Context) error {
		id, err := idGen.GetSnowIdInt64()
		snowId = id
		return err
	}); err != nil {
		return err
	}
	// 2、序列化请求
	var jsonBytes []byte
	if err := tracex.WithTrace(ctx, true, s.getTraceId(defineOp.OP_MARSHAL_REQ), func(ctx context.Context) error {
		jsonBytes_, err := jsonUtils.MarshalStructToByteArray(&bnSubScribeReq{
			Method: "SUBSCRIBE",
			Params: params,
			Id:     snowId,
		})
		jsonBytes = jsonBytes_
		return err
	}); err != nil {
		return err
	}
	// 3 、发送请求
	return s.sendWsData(ctx, s.getTraceId(sub_ADD), jsonBytes)
}

func (s *BnMarket) ListSub(ctx context.Context) error {
	// 1、获取当前的雪花ID
	var snowId int64
	if err := tracex.WithTrace(ctx, true, s.getTraceId(defineOp.OP_GET_SNOWFLAKE_ID), func(ctx context.Context) error {
		id, err := idGen.GetSnowIdInt64()
		snowId = id
		return err
	}); err != nil {
		return err
	}
	// 2、序列化请求
	var jsonBytes []byte
	if err := tracex.WithTrace(ctx, true, s.getTraceId(defineOp.OP_MARSHAL_REQ), func(ctx context.Context) error {
		jsonBytes_, err := jsonUtils.MarshalStructToByteArray(&bnSubScribeReq{
			Method: "LIST_SUBSCRIPTIONS",
			Id:     snowId,
		})
		jsonBytes = jsonBytes_
		return err
	}); err != nil {
		return err
	}
	// 3 、发送请求
	return s.sendWsData(ctx, s.getTraceId(sub_LIST), jsonBytes)
}

func (s *BnMarket) UnSub(ctx context.Context, params []string) error {
	// 1、获取当前的雪花ID
	var snowId int64
	if err := tracex.WithTrace(ctx, true, s.getTraceId(defineOp.OP_GET_SNOWFLAKE_ID), func(ctx context.Context) error {
		id, err := idGen.GetSnowIdInt64()
		snowId = id
		return err
	}); err != nil {
		return err
	}
	// 2、序列化请求
	var jsonBytes []byte
	if err := tracex.WithTrace(ctx, true, s.getTraceId(defineOp.OP_MARSHAL_REQ), func(ctx context.Context) error {
		jsonBytes_, err := jsonUtils.MarshalStructToByteArray(&bnSubScribeReq{
			Method: "UNSUBSCRIBE",
			Params: params,
			Id:     snowId,
		})
		jsonBytes = jsonBytes_
		return err
	}); err != nil {
		return err
	}
	// 3 、发送请求
	return s.sendWsData(ctx, s.getTraceId(sub_DEL), jsonBytes)
}

func (s *BnMarket) sendWsData(ctx context.Context, op string, jsonBytes []byte) error {
	return tracex.WithTrace(ctx, true, op, func(ctx context.Context) error {
		dynamicLog.Log.GetLog().Infof("BnMarket: %s", string(jsonBytes))
		return s.conn.SafeWriteMsg(websocket.TextMessage, jsonBytes)
	})
}

func (s *BnMarket) getTraceId(op string) string {
	return idGen.BuildName2("BnMarket", op)
}
