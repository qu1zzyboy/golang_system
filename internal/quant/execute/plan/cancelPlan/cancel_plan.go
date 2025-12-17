package cancelPlan

import "upbitBnServer/internal/quant/execute/order/orderModel"

type CancelPlan struct {
	Req      *orderModel.MyQueryOrderReq
	UpdateAt int64 //更新时间
}
