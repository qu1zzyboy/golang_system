package wsSub

import (
	"context"

	"upbitBnServer/internal/define/defineOp"
	"upbitBnServer/internal/infra/observe/log/dynamicLog"
	"upbitBnServer/internal/infra/observe/trace/tracex"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/quant/market/depth"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/pkg/utils/idGen"
	"upbitBnServer/pkg/utils/jsonUtils"

	"github.com/gorilla/websocket"
)

type BnDepth struct {
	conn        *wsDefine.SafeWrite // websocket连接
	levels      int                 // 深度档位
	updateSpeed int                 // 更新速度（毫秒）
}

func NewBnDepth(levels, updateSpeed int) *BnDepth {
	return &BnDepth{
		levels:      levels,
		updateSpeed: updateSpeed,
	}
}

func (s *BnDepth) DialToMarket(ctx context.Context, params []string) (*wsDefine.SafeWrite, error) {
	conn, _, err := websocket.DefaultDialer.Dial("wss://fstream.binance.com/ws", nil)
	if err != nil {
		return nil, connErr.WithCause(err)
	}
	s.conn = wsDefine.NewSafeWrite(conn)
	return s.conn, s.AddSub(ctx, params)
}

func (s *BnDepth) ParamBuild(resourceType resourceEnum.ResourceType) (func(symbol string) string, error) {
	if resourceType != resourceEnum.DELTA_DEPTH {
		return nil, resourceType.GetNotSupportError("bnDepthParamBuild")
	}
	return depth.BN.GetFuSubParam(s.levels, s.updateSpeed), nil
}

func (s *BnDepth) AddSub(ctx context.Context, params []string) error {
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
	// 3、发送请求
	return s.sendWsData(ctx, s.getTraceId(sub_ADD), jsonBytes)
}

func (s *BnDepth) ListSub(ctx context.Context) error {
	var snowId int64
	if err := tracex.WithTrace(ctx, true, s.getTraceId(defineOp.OP_GET_SNOWFLAKE_ID), func(ctx context.Context) error {
		id, err := idGen.GetSnowIdInt64()
		snowId = id
		return err
	}); err != nil {
		return err
	}
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
	return s.sendWsData(ctx, s.getTraceId(sub_LIST), jsonBytes)
}

func (s *BnDepth) UnSub(ctx context.Context, params []string) error {
	var snowId int64
	if err := tracex.WithTrace(ctx, true, s.getTraceId(defineOp.OP_GET_SNOWFLAKE_ID), func(ctx context.Context) error {
		id, err := idGen.GetSnowIdInt64()
		snowId = id
		return err
	}); err != nil {
		return err
	}
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
	return s.sendWsData(ctx, s.getTraceId(sub_DEL), jsonBytes)
}

func (s *BnDepth) sendWsData(ctx context.Context, op string, jsonBytes []byte) error {
	return tracex.WithTrace(ctx, true, op, func(ctx context.Context) error {
		dynamicLog.Log.GetLog().Infof("BnDepth: %s", string(jsonBytes))
		return s.conn.SafeWriteMsg(websocket.TextMessage, jsonBytes)
	})
}

func (s *BnDepth) getTraceId(op string) string {
	return idGen.BuildName2("BnDepth", op)
}
