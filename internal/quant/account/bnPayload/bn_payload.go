package bnPayload

import (
	"context"

	"upbitBnServer/internal/infra/ws/client/wsExecuteClient"
	"upbitBnServer/internal/infra/ws/wsDefine"
	"upbitBnServer/internal/infra/ws/wsDialImpl"
	"upbitBnServer/internal/quant/exchanges/exchangeEnum"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/resource/resourceEnum"
	"upbitBnServer/internal/strategy/toUpbitList/toUpBitDataStatic"
	"upbitBnServer/pkg/utils/byteUtils"
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
	param   *wsDialImpl.BnPayload
}

func NewBnPayload(apiKey, secretKey string) *BnPayload {
	return &BnPayload{
		param: wsDialImpl.NewBnPayload(apiKey, secretKey),
	}
}

func (s *BnPayload) RegisterReadHandler(ctx context.Context, accountKeyId uint8, read wsDefine.ReadAutoHandler) error {
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

func ParseOrderStatus(data []byte, cidEnd, totalLen uint16, accountKeyId uint8) (orderStatus execute.OrderStatus, X_len, ap_start, x_end uint16) {
	sell_start := cidEnd + 7
	sell_end := sell_start + 4
	if data[sell_start] == 'B' {
		sell_end = sell_start + 3
	}
	o_start := sell_end + 7
	var o_end uint16
	switch data[o_start] {
	case 'L':
		// LIMIT
		o_end = o_start + 5
	case 'M':
		// MARKET
		o_end = o_start + 6
	}
	q_start := o_end + 17
	q_end := byteUtils.FindNextQuoteIndex(data, q_start, totalLen)

	p_start := q_end + 7
	p_end := byteUtils.FindNextQuoteIndex(data, p_start, totalLen)

	ap_start = p_end + 8
	ap_end := byteUtils.FindNextQuoteIndex(data, ap_start, totalLen)

	x_start := ap_end + 16
	x_end = byteUtils.FindNextQuoteIndex(data, x_start, totalLen)

	switch data[x_end+7] {
	case 'N':
		orderStatus = execute.NEW
		X_len = 3
	case 'P':
		orderStatus = execute.PARTIALLY_FILLED
		X_len = 16
	case 'F':
		orderStatus = execute.FILLED
		X_len = 6
	case 'C':
		orderStatus = execute.CANCELED
		X_len = 8
	case 'R':
		orderStatus = execute.REJECTED
		X_len = 8
	case 'E':
		orderStatus = execute.EXPIRED
		X_len = 7
	default:
		orderStatus = execute.UNKNOWN_ORDER_STATUS
		toUpBitDataStatic.DyLog.GetLog().Errorf("[%d]ORDER_UPDATE: unknown order status[%d], json: %s", accountKeyId, x_end+7, string(data))
	}
	return
}
