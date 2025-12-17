package updatePlan

import "upbitBnServer/internal/quant/execute/order/orderModel"

type UpdatePlan struct {
	Req      *orderModel.MyModifyOrderReq
	UpdateAt int64 //更新时间
}
