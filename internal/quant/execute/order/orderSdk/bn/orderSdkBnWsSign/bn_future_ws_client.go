package orderSdkBnWsSign

import (
	"context"

	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/internal/resource/wsExecuteClient"
	"upbitBnServer/internal/resource/wsSub"
)

type FutureClient struct {
	secretByte []byte
	apiKey     string
	conn       *wsExecuteClient.Execute //ws client
	param      *wsSub.BnOrder
}

func (s *FutureClient) IsConnOk() bool {
	return s.conn.IsConnOk()
}

func NewFutureClient(apiKey, secretKey string) *FutureClient {
	return &FutureClient{
		apiKey:     apiKey,
		secretByte: []byte(secretKey),
		param:      wsSub.NewBnOrder(apiKey, secretKey),
	}
}

func (s *FutureClient) Close(ctx context.Context) error {
	return s.conn.CloseConn(ctx)
}

func (s *FutureClient) RegisterReadHandler(ctx context.Context, accountKeyId uint8, read wsDefine.ReadPrivateHandler) error {
	var err error
	s.conn, err = wsExecuteClient.NewExecute(exchangeEnum.BINANCE, resourceEnum.ORDER_WRITE, read, s.param, accountKeyId)
	if err != nil {
		return err
	}
	if err = s.conn.StartConn(ctx); err != nil {
		return err
	}
	return nil
}
