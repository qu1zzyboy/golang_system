package orderMonitorSdkBnWs

import (
	"context"

	"github.com/hhh500/quantGoInfra/infra/ws/wsDefine"
	"github.com/hhh500/quantGoInfra/pkg/container/pool/byteBufPool"
	"github.com/hhh500/quantGoInfra/quant/exchanges/exchangeEnum"
	"github.com/hhh500/quantGoInfra/resource/resourceEnum"
	"github.com/hhh500/upbitBnServer/internal/quant/execute/order/orderSdk/bn/orderSdkBnModel"
	"github.com/hhh500/upbitBnServer/internal/resource/wsExecuteClient"
	"github.com/hhh500/upbitBnServer/internal/resource/wsSub"
)

type FutureMonitorClient struct {
	secretByte   []byte
	apiKey       string                   // api key
	conn         *wsExecuteClient.Execute // ws client
	param        *wsSub.BnOrder           // 订阅参数
	accountKeyId uint8
}

func (s *FutureMonitorClient) IsConnOk() bool {
	return s.conn.IsConnOk()
}

func NewFutureClient(apiKey, secretKey string, accountKeyId uint8) *FutureMonitorClient {
	return &FutureMonitorClient{
		apiKey:       apiKey,
		secretByte:   []byte(secretKey),
		param:        wsSub.NewBnOrder(apiKey, secretKey),
		accountKeyId: accountKeyId,
	}
}

func (s *FutureMonitorClient) Close(ctx context.Context) error {
	return s.conn.CloseConn(ctx)
}

func (s *FutureMonitorClient) RegisterReadHandler(ctx context.Context, read wsDefine.ReadPrivateHandler) error {
	var err error
	s.conn, err = wsExecuteClient.NewExecute(exchangeEnum.BINANCE, resourceEnum.ORDER_WRITE, read, s.param, s.accountKeyId)
	if err != nil {
		return err
	}
	if err = s.conn.StartConn(ctx); err != nil {
		return err
	}
	return nil
}

func (s *FutureMonitorClient) CreateOrder(api *orderSdkBnModel.FuturePlaceLimitSdk) error {
	rawData, err := api.ParseWsReqFast(s.apiKey, s.secretByte)
	defer byteBufPool.ReleaseBuffer(rawData)
	if rawData == nil || err != nil {
		return err
	}
	if err = s.conn.WriteAsync(*rawData); err != nil {
		return err
	}
	return err
}
