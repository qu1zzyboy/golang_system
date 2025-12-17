package bnDriveOrderWrap

import (
	"time"
	"upbitBnServer/internal/quant/execute/order/bnOrderAppManager"
	"upbitBnServer/internal/quant/execute/order/orderModel"
	"upbitBnServer/internal/quant/execute/plan/cancelPlan"
	"upbitBnServer/internal/quant/execute/plan/placePlan"
	"upbitBnServer/internal/quant/execute/plan/updatePlan"
	"upbitBnServer/internal/strategy/newsDrive/common/driverOrderDedup"
	"upbitBnServer/internal/strategy/newsDrive/common/driverStatic"
)

func PlaceOrderWithPlan(accountKeyId uint8, req *orderModel.MyPlaceOrderReq) error {
	//下小订单
	if err := bnOrderAppManager.GetTradeManager().SendPlaceOrder(accountKeyId, req); err != nil {
		driverStatic.DyLog.GetLog().Errorf("[%s,%s,%s]创建订单错误: %s", req.SymbolName, req.UsageFrom, req.ClientOrderId, err.Error())
		return err
	}
	// 下单成功就存入交易计划
	driverOrderDedup.PlaceManager.Store(req.ClientOrderId, &placePlan.PlacePlan{Req: req, UpdateAt: time.Now().UnixMilli()})
	return nil
}

func ModifyOrderWithPlan(accountKeyId uint8, req *orderModel.MyModifyOrderReq) error {
	if err := bnOrderAppManager.GetTradeManager().SendModifyOrder(accountKeyId, req); err != nil {
		driverStatic.DyLog.GetLog().Errorf("[%s,%s,%s]修改订单错误: %s", req.SymbolName, req.UsageFrom, req.ClientOrderId, err.Error())
		return err
	}
	driverOrderDedup.ModifyManager.Store(req.ClientOrderId, &updatePlan.UpdatePlan{Req: req, UpdateAt: time.Now().UnixMilli()})
	return nil
}

func CancelOrderWithPlan(accountKeyId uint8, req *orderModel.MyQueryOrderReq) error {
	if err := bnOrderAppManager.GetTradeManager().SendCancelOrder(accountKeyId, req); err != nil {
		driverStatic.DyLog.GetLog().Errorf("[%s,%s,%s]撤销订单错误: %s", req.SymbolName, req.UsageFrom, req.ClientOrderId, err.Error())
		return err
	}
	driverOrderDedup.CancelManager.Store(req.ClientOrderId, &cancelPlan.CancelPlan{Req: req, UpdateAt: time.Now().UnixMilli()})
	return nil
}
