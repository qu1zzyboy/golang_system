package bnOrderSdkModel

import "upbitBnServer/internal/quant/execute"

const (
	max_order_place_batch = 5
)

const (
	order_resp_ask    = "ACK"
	oRDER_RESP_RESULT = "RESULT"
	ORDER_RESP_FULL   = "FULL"
)

func getBnOrderMode(orderMode execute.MyOrderMode) (orderSide, positionSide) {
	switch orderMode {
	case execute.ORDER_BUY_OPEN:
		return sideBuy, positionSideLONG
	case execute.ORDER_SELL_CLOSE:
		return sideSell, positionSideLONG
	case execute.ORDER_SELL_OPEN:
		return sideSell, positionSideSHORT
	case execute.ORDER_BUY_CLOSE:
		return sideBuy, positionSideSHORT
	default:
		return sideBuy, positionSideLONG
	}
}
