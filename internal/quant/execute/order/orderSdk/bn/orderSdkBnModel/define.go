package orderSdkBnModel

import "upbitBnServer/internal/quant/execute"

const (
	max_order_place_batch = 5
)

const (
	order_resp_ask    = "ACK"
	oRDER_RESP_RESULT = "RESULT"
	ORDER_RESP_FULL   = "FULL"
)

func getBnOrderMode(orderMode execute.OrderMode) (orderSide, positionSide, timeInForce, bool) {
	switch orderMode {
	case execute.BUY_OPEN_LIMIT:
		return sideBuy, positionSideLONG, timeInForceGTC, false
	case execute.BUY_OPEN_LIMIT_MAKER:
		return sideBuy, positionSideLONG, timeInForceGTX, false
	case execute.BUY_OPEN_MARKET:
		return sideBuy, positionSideLONG, timeInForceGTC, true

	case execute.SELL_CLOSE_LIMIT:
		return sideSell, positionSideLONG, timeInForceGTC, false
	case execute.SELL_CLOSE_LIMIT_MAKER:
		return sideSell, positionSideLONG, timeInForceGTX, false

	case execute.SELL_OPEN_LIMIT:
		return sideSell, positionSideSHORT, timeInForceGTC, false
	case execute.SELL_OPEN_LIMIT_MAKER:
		return sideSell, positionSideSHORT, timeInForceGTX, false
	case execute.SELL_OPEN_MARKET:
		return sideSell, positionSideSHORT, timeInForceGTC, true

	case execute.BUY_CLOSE_LIMIT:
		return sideBuy, positionSideSHORT, timeInForceGTC, false
	case execute.BUY_CLOSE_LIMIT_MAKER:
		return sideBuy, positionSideSHORT, timeInForceGTX, false
	default:
		return sideBuy, positionSideLONG, timeInForceGTC, false
	}
}
