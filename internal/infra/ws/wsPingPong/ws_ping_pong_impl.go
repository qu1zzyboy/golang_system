package wsPingPong

import (
	"time"
	"upbitBnServer/pkg/utils/time2str"

	"upbitBnServer/internal/infra/errorx"
	"upbitBnServer/internal/infra/errorx/errCode"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/pkg/utils/jsonUtils"

	"github.com/gorilla/websocket"
)

// func (c *PingPong) WrapByBitPongHandler(read wsDefine.ReadHandler) wsDefine.ReadHandler {
// 	return func(data []byte) {
// 		if gjson.GetBytes(data, "op").Exists() {
// 			opStr := gjson.GetBytes(data, "op").String()
// 			if opStr == "pong" || opStr == "ping" {
// 				c.UpdatePong()
// 				return
// 			}
// 		}
// 		read(data)
// 	}
// }

func PingBn(conn *wsDefine.SafeWrite) error {
	if err := conn.SafeWriteControl(websocket.PingMessage, []byte{}, time.Now().Add(3*time.Second)); err != nil {
		return errorx.New(errCode.BN_PING_SEND_ERROR, "binance_ws ping发送失败").WithCause(err)
	}
	return nil
}

func PongBn(msg string, conn *wsDefine.SafeWrite) error {
	if err := conn.SafeWriteControl(websocket.PongMessage, []byte(msg), time.Now().Add(3*time.Second)); err != nil {
		return errorx.New(errCode.BN_PONG_SEND_ERROR, "binance_ws pong发送失败").WithCause(err)
	}
	return nil
}

func PingByBit(conn *wsDefine.SafeWrite) error {
	pingByte, err := jsonUtils.MarshalStructToByteArray(map[string]string{
		"op":     "ping",
		"req_id": time2str.GetNowTimeStampMilliStr(),
	})
	if err != nil {
		return errorx.Newf(errCode.CodeWsParamError, "BYBIT_PING_MARSHAL", "bybit ws ping参数序列化错误").WithCause(err)
	}
	if err := conn.SafeWriteMsg(websocket.TextMessage, pingByte); err != nil {
		return errorx.Newf(errCode.CodeWsDoError, "BYBIT_PING_SEND_ERROR", "bybit_ws ping发送失败").WithCause(err)
	}
	return nil
}

func GetPingFunc(exType exchangeEnum.ExchangeType) (wsDefine.PingFunc, error) {
	switch exType {
	case exchangeEnum.BINANCE:
		return PingBn, nil
	case exchangeEnum.BYBIT:
		return PingByBit, nil
	default:
		return nil, exType.GetNotSupportError("ws_ping")
	}
}
