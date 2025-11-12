package orderSdkBnWsNoSign

import (
	"context"
	"upbitBnServer/internal/infra/ws/client/wsExecuteClient"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/infra/ws/wsDialImpl"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/resource/resourceEnum"
)

type FutureClient struct {
	conn  *wsExecuteClient.Execute //ws client
	param *wsDialImpl.BnOrderAuth
}

func (s *FutureClient) IsConnOk() bool {
	return s.conn.IsConnOk()
}

func NewFutureClient(apiKey, secretKey string) *FutureClient {
	return &FutureClient{
		param: wsDialImpl.NewBnOrderAuth(apiKey, secretKey),
	}
}

func (s *FutureClient) Close(ctx context.Context) error {
	return s.conn.CloseConn(ctx)
}

func (s *FutureClient) RegisterReadHandler(ctx context.Context, accountKeyId uint8, read wsDefine.ReadAutoHandler) error {
	var err error
	s.conn, err = wsExecuteClient.NewExecute(exchangeEnum.BINANCE, resourceEnum.WS_REQUEST_PRIVATE, read, s.param, accountKeyId)
	if err != nil {
		return err
	}
	if err = s.conn.StartConn(ctx); err != nil {
		return err
	}
	return nil
}
