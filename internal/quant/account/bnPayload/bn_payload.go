package bnPayload

import (
	"context"

	"github.com/hhh500/quantGoInfra/infra/ws/wsDefine"
	"github.com/hhh500/quantGoInfra/quant/exchanges/exchangeEnum"
	"github.com/hhh500/quantGoInfra/resource/resourceEnum"
	"github.com/hhh500/upbitBnServer/internal/resource/wsExecuteClient"
	"github.com/hhh500/upbitBnServer/internal/resource/wsSub"
)

const (
	ORDER_TRADE_UPDATE    = "ORDER_TRADE_UPDATE"
	ACCOUNT_UPDATE        = "ACCOUNT_UPDATE"
	TRADE_LITE            = "TRADE_LITE"
	ALGO_UPDATE           = "ALGO_UPDATE"
	ACCOUNT_CONFIG_UPDATE = "ACCOUNT_CONFIG_UPDATE"
)

type BnPayload struct {
	payload *wsExecuteClient.Execute
	param   *wsSub.BnPayload
}

func NewBnPayload(apiKey, secretKey string) *BnPayload {
	return &BnPayload{
		param: wsSub.NewBnPayload(apiKey, secretKey),
	}
}

func (s *BnPayload) RegisterReadHandler(ctx context.Context, accountKeyId uint8, read wsDefine.ReadPrivateHandler) error {
	var err error
	s.payload, err = wsExecuteClient.NewExecute(exchangeEnum.BINANCE, resourceEnum.PAYLOAD_READ, read, s.param, accountKeyId)
	if err != nil {
		return err
	}
	if err = s.payload.StartConn(ctx); err != nil {
		return err
	}
	return nil
}
