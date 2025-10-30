package orderSdkBnModel

import (
	"net/url"
	"upbitBnServer/pkg/utils/time2str"

	"upbitBnServer/internal/quant/execute/order/orderModel"

	"github.com/shopspring/decimal"
)

type SpotPlaceSdk struct {
	ClientOrderId string           //No	必须满足正则规则 ^[\.A-Z\:/a-z0-9_-]{1,36}$
	symbolName    string           //Yes	交易对
	side          string           //Yes	买卖方向 SELL, BUY
	orderType     string           //Yes	订单类型 LIMIT, MARKET...
	origVolume    *decimal.Decimal //No	下单数量,使用closePosition不支持此参数。
	origPrice     *decimal.Decimal //No	委托价格
	stopPrice     *decimal.Decimal //No	止损价格
	orderRespType *string          //No	"ACK", "RESULT", 默认 "ACK"
	timeInForce   *string          //No	有效方法
}

func (api *SpotPlaceSdk) ClientOrderId_(newClientOrderId string) *SpotPlaceSdk {
	api.ClientOrderId = newClientOrderId
	return api
}
func (api *SpotPlaceSdk) Symbol_(symbol string) *SpotPlaceSdk {
	api.symbolName = symbol
	return api
}
func (api *SpotPlaceSdk) Side_(side string) *SpotPlaceSdk {
	api.side = side
	return api
}
func (api *SpotPlaceSdk) Type_(orderType string) *SpotPlaceSdk {
	api.orderType = orderType
	return api
}

func (api *SpotPlaceSdk) Volume_(quantity decimal.Decimal) *SpotPlaceSdk {
	api.origVolume = &quantity
	return api
}
func (api *SpotPlaceSdk) Price_(price decimal.Decimal) *SpotPlaceSdk {
	api.origPrice = &price
	return api
}

func (api *SpotPlaceSdk) StopPrice_(stopPrice decimal.Decimal) *SpotPlaceSdk {
	api.stopPrice = &stopPrice
	return api
}

func (api *SpotPlaceSdk) TimeInForce_(timeInForce string) *SpotPlaceSdk {
	api.timeInForce = &timeInForce
	return api
}
func (api *SpotPlaceSdk) OrderRespType_(newOrderRespType string) *SpotPlaceSdk {
	api.orderRespType = &newOrderRespType
	return api
}

func (api *SpotPlaceSdk) ParseRestRequest() string {
	param := url.Values{}
	param.Set(p_SYMBOL, api.symbolName)
	param.Set(p_SIDE, api.side)
	param.Set(p_TYPE, api.orderType)
	param.Set(p_NEW_CLIENT_ORDER_ID, api.ClientOrderId)
	if api.timeInForce != nil {
		param.Set(p_TIME_IN_FORCE, *api.timeInForce)
	}
	if api.origPrice != nil {
		param.Set(p_PRICE, api.origPrice.String())
	}
	if api.stopPrice != nil {
		param.Set(p_STOP_PRICE, api.stopPrice.String())
	}
	if api.origVolume != nil {
		param.Set(p_QUANTITY, api.origVolume.String())
	}
	if api.orderRespType != nil {
		param.Set(p_NEW_ORDER_RESP_TYPE, *api.orderRespType)
	}
	param.Set(p_TIME_STAMP, time2str.GetNowTimeStampMilliStr())
	// 编码 query & form
	return param.Encode()
}

// NewSpotPlaceSdk    rest下单 (TRADE)
func NewSpotPlaceSdk() *SpotPlaceSdk {
	return &SpotPlaceSdk{}
}

func GetSpotPlaceSdk(req *orderModel.MyPlaceOrderReq) *SpotPlaceSdk {
	return nil
	// side, _ := getBnOrderMode(req.OrderMode)
	// timeInForce := gtc
	// orderType := limit
	// switch req.OrderType {
	// case execute.ORDER_TYPE_POST_ONLY:
	// 	orderType = limit_maker
	// 	timeInForce = ""
	// case execute.ORDER_TYPE_IOC:
	// 	timeInForce = ioc
	// case execute.ORDER_TYPE_GTD:
	// 	timeInForce = gtd
	// default:
	// }
	// client := NewSpotPlaceSdk().
	// 	Symbol_(req.StaticMeta.SymbolName).
	// 	Side_(side).
	// 	Type_(orderType).
	// 	Volume_(req.OrigVol).
	// 	Price_(req.OrigPrice).
	// 	OrderRespType_(oRDER_RESP_RESULT)
	// if timeInForce != "" {
	// 	client.TimeInForce_(timeInForce)
	// }
	// if req.ClientOrderId != "" {
	// 	client.ClientOrderId_(req.ClientOrderId)
	// }
	// return client
}
