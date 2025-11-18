package byBitPayload

import (
	"context"
	"upbitBnServer/internal/infra/ws/client/wsExecuteClient"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/infra/ws/wsDialImpl"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/resource/resourceEnum"
)

const (
	orderLINEAR     = "order.linear"
	tradeLITE       = "execution.fast.linear"
	POSITION_LINEAR = "position.linear"
	wallet          = "wallet"
)

type ByBitPayload struct {
	payload *wsExecuteClient.Execute
	param   *wsDialImpl.BybitPayload
}

func NewByBitPayload(apiKey, secretKey string) *ByBitPayload {
	return &ByBitPayload{
		param: wsDialImpl.NewBybitPayload(apiKey, secretKey),
	}
}

func (s *ByBitPayload) RegisterReadHandler(ctx context.Context, accountKeyId uint8, read wsDefine.ReadAutoHandler) error {
	var err error
	if err = s.param.SetInitParams([]string{tradeLITE, orderLINEAR, wallet}); err != nil {
		return err
	}
	s.payload, err = wsExecuteClient.NewExecute(exchangeEnum.BYBIT, resourceEnum.PAYLOAD_READ, read, s.param, accountKeyId)
	if err != nil {
		return err
	}
	if err = s.payload.StartConn(ctx); err != nil {
		return err
	}
	return nil
}
