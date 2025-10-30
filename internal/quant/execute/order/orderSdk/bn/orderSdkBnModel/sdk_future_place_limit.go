package orderSdkBnModel

import (
	"upbitBnServer/internal/infra/errorx/errDefine"
	"upbitBnServer/internal/quant/execute"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/utils/myCrypto"
	"upbitBnServer/pkg/container/pool/byteBufPool"
	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/pkg/utils/timeUtils"

	"github.com/shopspring/decimal"
)

type FuturePlaceLimitSdk struct {
	ClientOrderId string           //YES 必须满足正则规则 ^[\.A-Z\:/a-z0-9_-]{1,36}$
	symbolName    string           //Yes 交易对
	origVolume    *decimal.Decimal //No	下单数量,使用closePosition不支持此参数。
	origPrice     *decimal.Decimal //No	委托价格
	side          orderSide        //Yes 买卖方向 SELL, BUY
	orderType     orderType        //Yes 订单类型 LIMIT, MARKET...
	positionSide  positionSide     //No	持仓方向,单向持仓模式下非必填，默认且仅可填BOTH;在双向持仓模式下必填,且仅可选择 LONG 或 SHORT
	timeInForce   timeInForce      //No	有效方法
	// orderRespType *string          //No	"ACK", "RESULT", 默认 "ACK"
}

func (api *FuturePlaceLimitSdk) ClientOrderId_(newClientOrderId string) *FuturePlaceLimitSdk {
	api.ClientOrderId = newClientOrderId
	return api
}
func (api *FuturePlaceLimitSdk) Symbol_(symbol string) *FuturePlaceLimitSdk {
	api.symbolName = symbol
	return api
}
func (api *FuturePlaceLimitSdk) Side_(side orderSide) *FuturePlaceLimitSdk {
	api.side = side
	return api
}

func (api *FuturePlaceLimitSdk) OrderType_(orderType orderType) *FuturePlaceLimitSdk {
	api.orderType = orderType
	return api
}
func (api *FuturePlaceLimitSdk) PositionSide_(positionSide positionSide) *FuturePlaceLimitSdk {
	api.positionSide = positionSide
	return api
}
func (api *FuturePlaceLimitSdk) Volume_(quantity decimal.Decimal) *FuturePlaceLimitSdk {
	api.origVolume = &quantity
	return api
}
func (api *FuturePlaceLimitSdk) Price_(price decimal.Decimal) *FuturePlaceLimitSdk {
	api.origPrice = &price
	return api
}

func (api *FuturePlaceLimitSdk) TimeInForce_(timeInForce timeInForce) *FuturePlaceLimitSdk {
	api.timeInForce = timeInForce
	return api
}

// func (api *FuturePlaceLimitSdk) OrderRespType_(newOrderRespType string) *FuturePlaceLimitSdk {
// 	api.orderRespType = &newOrderRespType
// 	return api
// }

// NewFuturePlaceSdk   rest下单 (TRADE)
func NewFuturePlaceSdk() *FuturePlaceLimitSdk {
	return &FuturePlaceLimitSdk{}
}

func GetFuturePlaceLimitSdk(req *orderModel.MyPlaceOrderReq) *FuturePlaceLimitSdk {
	side, psSide := getBnOrderMode(req.OrderMode)
	timeInForce := timeInForceGTC
	switch req.OrderType {
	case execute.ORDER_TYPE_POST_ONLY:
		timeInForce = timeInForceGTX
	case execute.ORDER_TYPE_IOC:
		timeInForce = timeInForceIOC
	case execute.ORDER_TYPE_MARKET:
		return NewFuturePlaceSdk().
			ClientOrderId_(req.ClientOrderId).
			Symbol_(req.StaticMeta.SymbolName).
			Side_(side).
			PositionSide_(psSide).
			Volume_(req.OrigVol).
			OrderType_(orderTypeMarket).
			TimeInForce_(timeInForce)
	default:
	}
	return NewFuturePlaceSdk().
		ClientOrderId_(req.ClientOrderId).
		Symbol_(req.StaticMeta.SymbolName).
		Side_(side).
		PositionSide_(psSide).
		Price_(req.OrigPrice).
		Volume_(req.OrigVol).
		OrderType_(orderTypeLimit).
		TimeInForce_(timeInForce)
}

// fast可以少遍历一个元素
var placeSortedKeyFast = []string{p_API_KEY, p_NEW_CLIENT_ORDER_ID, p_NEW_ORDER_RESP_TYPE,
	p_POSITION_SIDE, p_PRICE, p_QUANTITY, p_SIDE, p_STOP_PRICE, p_SYMBOL, p_TIME_IN_FORCE, p_TIME_STAMP, p_TYPE}

// ParseRestReqFast 551.8 ns/op	     480 B/op	      21 allocs/op
func (api *FuturePlaceLimitSdk) ParseRestReqFast() *[]byte {
	orig := byteBufPool.AcquireBuffer(256)
	*orig = append(*orig, b_NEW_CLIENT_ORDER_ID...)
	*orig = append(*orig, api.ClientOrderId...)

	*orig = append(*orig, b_SYMBOL...)
	*orig = append(*orig, api.symbolName...)

	*orig = append(*orig, b_SIDE...)
	*orig = append(*orig, orderSideArr[api.side]...)

	*orig = append(*orig, b_TYPE...)
	*orig = append(*orig, orderTypeArr[api.orderType]...)

	*orig = append(*orig, b_TIME_IN_FORCE...)
	*orig = append(*orig, timeInForceArr[api.timeInForce]...)

	if api.origPrice != nil {
		*orig = append(*orig, b_PRICE...)
		*orig = append(*orig, api.origPrice.String()...)
	}
	if api.origVolume != nil {
		*orig = append(*orig, b_QUANTITY...)
		*orig = append(*orig, api.origVolume.String()...)
	}
	// if api.orderRespType != nil {
	// 	*orig = append(*orig, b_ORDER_RESP_TYPE...)
	// 	*orig = append(*orig, *api.orderRespType...)
	// }
	*orig = append(*orig, b_POSITION_SIDE...)
	*orig = append(*orig, positionSideArr[api.positionSide]...)
	*orig = append(*orig, b_TIME_STAMP...)
	*orig = convertx.AppendValueToBytes(*orig, timeUtils.GetNowTimeUnixMilli())
	return orig
}

// ParseWsReqFast
// 性能说明: 1970 ns/op	1754 B/op  38 allocs/op
func (api *FuturePlaceLimitSdk) ParseWsReqFast(apiKey string, secretByte []byte) (*[]byte, error) {
	if api.ClientOrderId == "" {
		return nil, errDefine.ClientOrderIdEmpty
	}
	m := make(map[string]any)
	m[p_SYMBOL] = api.symbolName
	m[p_SIDE] = orderSideArr[api.side]
	m[p_TYPE] = orderTypeArr[api.orderType]
	m[p_NEW_CLIENT_ORDER_ID] = api.ClientOrderId
	if api.orderType != orderTypeMarket {
		m[p_TIME_IN_FORCE] = timeInForceArr[api.timeInForce]
	}
	if api.origPrice != nil {
		m[p_PRICE] = api.origPrice.String()
	}
	if api.origVolume != nil {
		m[p_QUANTITY] = api.origVolume.String()
	}
	m[p_POSITION_SIDE] = positionSideArr[api.positionSide]
	//统一逻辑
	m[p_API_KEY] = apiKey
	m[p_TIME_STAMP] = timeUtils.GetNowTimeUnixMilli()
	signData := buildQueryBytePool(256, m, placeSortedKeyFast) //从池子中获取256位签名数据
	signRes := byteBufPool.AcquireBuffer(64)                   //从池子中获取64位
	defer byteBufPool.ReleaseBuffer(signData)                  //释放签名数据
	defer byteBufPool.ReleaseBuffer(signRes)                   //释放签名值
	if err := myCrypto.HmacSha256Fast(secretByte, *signData, signRes); err != nil {
		return nil, err
	}
	return buildWsReqFast(512, "P"+api.ClientOrderId, "order.place", m, placeSortedKeyFast, signRes), nil
}

// {
// 	"id": "Pfu0-u-HEMI7383069262832632231",
// 	"method": "order.place",
// 	"params": {
// 		"apiKey": "c2Y1zMXaZcz85k6MrQY1Qo3FEEy81ookmcI3js0KJrMfT0EL5pgvwSgHKfSbu7aH",
// 		"newClientOrderId": "fu0-u-HEMI7383069262832632231",
// 		"newOrderRespType": "ACK",
// 		"positionSide": "SHORT",
// 		"price": "0.076831",
// 		"quantity": "149",
// 		"side": "SELL",
// 		"symbol": "HEMIUSDT",
// 		"timeInForce": "GTC",
// 		"timestamp": "1760260883052",
// 		"type": "LIMIT",
// 		"signature": "a414f80d1d27686d282f14796a2f5286bc140fe1b7669a1d7d579b54ef6b1d8f"
// 	}
// }
