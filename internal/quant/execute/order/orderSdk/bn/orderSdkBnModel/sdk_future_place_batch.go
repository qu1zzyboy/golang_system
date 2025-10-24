package orderSdkBnModel

import (
	"net/url"

	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/pkg/utils/convertx"
	"upbitBnServer/pkg/utils/jsonUtils"
)

type paramBatch map[string]any

// FuturePlaceBatchSdk rest批量下单 (TRADE)
type FuturePlaceBatchSdk struct {
	BatchOrders []*FuturePlaceLimitSdk //YES	批量下单 最多支持5个
}

func (s *FuturePlaceBatchSdk) AddOrders(orderReqs ...*FuturePlaceLimitSdk) *FuturePlaceBatchSdk {
	if s.BatchOrders == nil {
		s.BatchOrders = make([]*FuturePlaceLimitSdk, 0)
	}
	s.BatchOrders = append(s.BatchOrders, orderReqs...)
	return s
}

func (s *FuturePlaceBatchSdk) SetOrders(orderReqs []*FuturePlaceLimitSdk) *FuturePlaceBatchSdk {
	s.BatchOrders = make([]*FuturePlaceLimitSdk, 0, len(orderReqs))
	s.BatchOrders = append(s.BatchOrders, orderReqs...)
	return s
}

func (s *FuturePlaceBatchSdk) ParseRestRequest() (string, error) {
	var orders []paramBatch
	// for _, order := range s.BatchOrders {
	// 	m := paramBatch{
	// 		p_SYMBOL: order.symbolName,
	// 		p_SIDE:   order.side,
	// 		p_TYPE:   order.orderType,
	// 	}
	// 	m[p_NEW_CLIENT_ORDER_ID] = order.ClientOrderId
	// 	if order.timeInForce != nil {
	// 		m[p_TIME_IN_FORCE] = *order.timeInForce
	// 	}
	// 	if order.origPrice != nil {
	// 		m[p_PRICE] = order.origPrice
	// 	}
	// 	if order.origVolume != nil {
	// 		m[p_QUANTITY] = order.origVolume
	// 	}
	// 	if order.orderRespType != nil {
	// 		m[p_NEW_ORDER_RESP_TYPE] = *order.orderRespType
	// 	}
	// 	if order.positionSide != nil {
	// 		m[p_POSITION_SIDE] = *order.positionSide
	// 	}
	// 	orders = append(orders, m)
	// }
	b, err := jsonUtils.MarshalStructToByteArray(orders)
	if err != nil {
		return "", err
	}
	param := url.Values{}
	param.Set(p_BATCH_ORDERS, string(b))
	param.Set(p_TIME_STAMP, convertx.GetNowTimeStampMilliStr())
	// 编码 query & form
	return param.Encode(), nil
}

// NewFuturePlaceBatchSdk  rest批量下单 (TRADE)
func NewFuturePlaceBatchSdk() *FuturePlaceBatchSdk {
	return &FuturePlaceBatchSdk{}
}

func GetFuturePlaceBatchSdk(req *orderModel.MyPlaceOrderBatchReq) *FuturePlaceBatchSdk {
	var res []*FuturePlaceLimitSdk
	for k, v := range req.Orders {
		if k >= max_order_place_batch {
			break
		}
		res = append(res, GetFuturePlaceLimitSdk(v))
	}
	return NewFuturePlaceBatchSdk().SetOrders(res)
}
