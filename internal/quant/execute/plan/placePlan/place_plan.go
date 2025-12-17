package placePlan

import "upbitBnServer/internal/quant/execute/order/orderModel"

type PlacePlan struct {
	Req      *orderModel.MyPlaceOrderReq
	UpdateAt int64 //更新时间
}
